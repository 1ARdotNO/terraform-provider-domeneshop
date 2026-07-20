package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

// mockDomeneshop is an in-memory implementation of the subset of the
// Domeneshop API used by this provider. Every request and response is
// validated against the vendored OpenAPI document (testdata/openapi.json,
// extracted from https://api.domeneshop.no/docs/), so the acceptance tests
// double as a contract check: any traffic that does not match the
// documentation turns into a "contract violation" error response, which in
// turn fails the test that produced it.
//
// Known divergences between the documentation and the generated API client
// are tolerated but reported once on stderr; see the drift notes in
// canonicalizeRecordJSON.
type mockDomeneshop struct {
	mu     sync.Mutex
	router routers.Router

	domains  map[int64]map[string]interface{}
	records  map[int64]map[int64]map[string]interface{}
	forwards map[int64]map[string]map[string]interface{}
	nextID   int64

	driftMu sync.Mutex
	drift   map[string]bool
}

// newMockDomeneshop starts the mock API and returns the server. The caller
// must Close() it.
func newMockDomeneshop() (*httptest.Server, error) {
	m := &mockDomeneshop{
		domains:  map[int64]map[string]interface{}{},
		records:  map[int64]map[int64]map[string]interface{}{},
		forwards: map[int64]map[string]map[string]interface{}{},
		nextID:   1000,
		drift:    map[string]bool{},
	}
	m.domains[1] = map[string]interface{}{
		"id":              int64(1),
		"domain":          "example.com",
		"expiry_date":     "2030-01-01",
		"registered_date": "2020-01-01",
		"renew":           true,
		"registrant":      "Test Testesen",
		"status":          "active",
		"nameservers":     []interface{}{"ns1.hyp.net", "ns2.hyp.net", "ns3.hyp.net"},
		"services": map[string]interface{}{
			"registrar": true,
			"dns":       true,
			"email":     true,
			"webhotel":  "none",
		},
	}
	m.records[1] = map[int64]map[string]interface{}{}
	m.forwards[1] = map[string]map[string]interface{}{}

	srv := httptest.NewServer(m)

	doc, err := openapi3.NewLoader().LoadFromFile("testdata/openapi.json")
	if err != nil {
		srv.Close()
		return nil, fmt.Errorf("loading OpenAPI document: %w", err)
	}
	// The router matches requests against the document's server URLs;
	// point them at this test server.
	doc.Servers = openapi3.Servers{&openapi3.Server{URL: srv.URL}}
	router, err := gorillamux.NewRouter(doc)
	if err != nil {
		srv.Close()
		return nil, fmt.Errorf("building OpenAPI router: %w", err)
	}
	m.router = router

	return srv, nil
}

func (m *mockDomeneshop) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)

	route, pathParams, err := m.router.FindRoute(r)
	if err != nil && strings.HasSuffix(r.URL.Path, "/forwards") {
		// The documentation declares the forwards collection as
		// /domains/{domainId}/forwards/ (trailing slash), but the generated
		// client requests it without the slash.
		retry := r.Clone(r.Context())
		retry.URL.Path += "/"
		if route, pathParams, err = m.router.FindRoute(retry); err == nil {
			m.noteDrift([]string{"client drift: the forwards collection is requested without the trailing slash documented in the path /domains/{domainId}/forwards/"})
			r = retry
		}
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("contract violation: %s %s is not a documented operation: %v", r.Method, r.URL.Path, err), http.StatusInternalServerError)
		return
	}

	reqInput, err := m.validateRequest(r, route, pathParams, body)
	if err != nil {
		http.Error(w, fmt.Sprintf("contract violation: request %s %s does not match the API documentation: %v", r.Method, r.URL.Path, err), http.StatusInternalServerError)
		return
	}

	if user, _, ok := r.BasicAuth(); !ok || user == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"help": "provide credentials using HTTP basic auth"}`)
		return
	}

	rec := httptest.NewRecorder()
	r.Body = io.NopCloser(bytes.NewReader(body))
	m.handle(rec, r)

	if err := m.validateResponse(rec, reqInput); err != nil {
		http.Error(w, fmt.Sprintf("contract violation: response for %s %s does not match the API documentation: %v", r.Method, r.URL.Path, err), http.StatusInternalServerError)
		return
	}

	for k, vs := range rec.Header() {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(rec.Code)
	w.Write(rec.Body.Bytes())
}

func (m *mockDomeneshop) validateRequest(r *http.Request, route *routers.Route, pathParams map[string]string, body []byte) (*openapi3filter.RequestValidationInput, error) {
	mk := func(b []byte) *openapi3filter.RequestValidationInput {
		req := r.Clone(r.Context())
		req.Body = io.NopCloser(bytes.NewReader(b))
		req.ContentLength = int64(len(b))
		return &openapi3filter.RequestValidationInput{
			Request:    req,
			PathParams: pathParams,
			Route:      route,
			Options:    &openapi3filter.Options{AuthenticationFunc: openapi3filter.NoopAuthenticationFunc},
		}
	}

	base := mk(body)
	err := openapi3filter.ValidateRequest(r.Context(), mk(body))
	if err == nil {
		return base, nil
	}

	canon := canonicalizeRecordJSON(body, true)
	if canon.skip {
		m.noteDrift(canon.notes)
		return base, nil
	}
	for _, v := range canon.variants {
		if openapi3filter.ValidateRequest(r.Context(), mk(v)) == nil {
			m.noteDrift(canon.notes)
			return base, nil
		}
	}
	return nil, err
}

func (m *mockDomeneshop) validateResponse(rec *httptest.ResponseRecorder, reqInput *openapi3filter.RequestValidationInput) error {
	ctx := reqInput.Request.Context()
	body := rec.Body.Bytes()
	mk := func(b []byte) *openapi3filter.ResponseValidationInput {
		in := &openapi3filter.ResponseValidationInput{
			RequestValidationInput: reqInput,
			Status:                 rec.Code,
			Header:                 rec.Header(),
			Options:                reqInput.Options,
		}
		in.SetBodyBytes(b)
		return in
	}

	err := openapi3filter.ValidateResponse(ctx, mk(body))
	if err == nil {
		return nil
	}

	canon := canonicalizeRecordJSON(body, false)
	if canon.skip {
		m.noteDrift(canon.notes)
		return nil
	}
	for _, v := range canon.variants {
		if openapi3filter.ValidateResponse(ctx, mk(v)) == nil {
			m.noteDrift(canon.notes)
			return nil
		}
	}
	return err
}

// Known divergences between the API documentation and the generated client
// (github.com/innovationnorway/go-domeneshop) / real API behavior. When plain
// validation fails, bodies are re-validated in these canonicalized forms; if a
// canonical form passes, the divergence is reported as drift instead of
// failing the test.
var undocumentedRecordTypes = map[string]bool{
	// Supported by the provider and the real API, but absent from the
	// documentation's DNSRecord oneOf schema.
	"CAA": true, "NS": true, "DS": true, "ANAME": true,
}

// Fields the documentation declares as integers but the generated client
// marshals as JSON strings.
var stringlyNumericRecordFields = []string{
	"priority", "weight", "port", "flags", "tag", "alg", "digest", "usage", "selector", "dtype",
}

type canonicalized struct {
	variants [][]byte
	notes    []string
	skip     bool
}

func canonicalizeRecordJSON(body []byte, isRequest bool) canonicalized {
	var out canonicalized
	if len(bytes.TrimSpace(body)) == 0 {
		return out
	}
	var v interface{}
	if err := json.Unmarshal(body, &v); err != nil {
		return out
	}

	coerce := func(rec map[string]interface{}) bool {
		changed := false
		for _, f := range stringlyNumericRecordFields {
			if s, ok := rec[f].(string); ok {
				if n, err := strconv.Atoi(s); err == nil {
					rec[f] = n
					changed = true
					out.notes = append(out.notes, fmt.Sprintf("client drift: DNS record field %q is transmitted as a JSON string, but the documentation specifies an integer", f))
				}
			}
		}
		return changed
	}

	switch t := v.(type) {
	case map[string]interface{}:
		if typ, ok := t["type"].(string); ok && undocumentedRecordTypes[typ] {
			out.notes = append(out.notes, fmt.Sprintf("documentation gap: DNS record type %q is supported by the provider but missing from the documented DNSRecord schema", typ))
			out.skip = true
			return out
		}
		changed := coerce(t)
		if changed {
			b, _ := json.Marshal(t)
			out.variants = append(out.variants, b)
		}
		if isRequest {
			if _, ok := t["id"]; ok {
				delete(t, "id")
				out.notes = append(out.notes, `client drift: read-only field "id" is included in write requests`)
				b, _ := json.Marshal(t)
				out.variants = append(out.variants, b)
			}
		}
	case []interface{}:
		filtered := make([]interface{}, 0, len(t))
		changed := false
		for _, e := range t {
			rec, ok := e.(map[string]interface{})
			if !ok {
				filtered = append(filtered, e)
				continue
			}
			if typ, ok := rec["type"].(string); ok && undocumentedRecordTypes[typ] {
				out.notes = append(out.notes, fmt.Sprintf("documentation gap: DNS record type %q is supported by the provider but missing from the documented DNSRecord schema", typ))
				changed = true
				continue
			}
			if coerce(rec) {
				changed = true
			}
			filtered = append(filtered, rec)
		}
		if changed {
			b, _ := json.Marshal(filtered)
			out.variants = append(out.variants, b)
		}
	}
	return out
}

func (m *mockDomeneshop) noteDrift(notes []string) {
	m.driftMu.Lock()
	defer m.driftMu.Unlock()
	for _, n := range notes {
		if !m.drift[n] {
			m.drift[n] = true
			fmt.Fprintf(os.Stderr, "[api-contract] %s\n", n)
		}
	}
}

// handle implements the API behavior itself; all contract validation happens
// in ServeHTTP before and after this.
func (m *mockDomeneshop) handle(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	if len(parts) >= 1 && parts[0] == "domains" {
		if len(parts) == 1 {
			m.listDomains(w, r)
			return
		}
		domainID, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil || m.domains[domainID] == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		switch {
		case len(parts) == 2:
			m.writeJSON(w, http.StatusOK, m.domains[domainID])
		case parts[2] == "dns":
			m.handleDNS(w, r, domainID, parts[3:])
		case parts[2] == "forwards":
			m.handleForwards(w, r, domainID, parts[3:])
		default:
			w.WriteHeader(http.StatusNotFound)
		}
		return
	}
	w.WriteHeader(http.StatusNotFound)
}

func (m *mockDomeneshop) listDomains(w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("domain")
	ids := make([]int64, 0, len(m.domains))
	for id, d := range m.domains {
		if filter == "" || strings.Contains(d["domain"].(string), filter) {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	list := make([]interface{}, 0, len(ids))
	for _, id := range ids {
		list = append(list, m.domains[id])
	}
	m.writeJSON(w, http.StatusOK, list)
}

func (m *mockDomeneshop) handleDNS(w http.ResponseWriter, r *http.Request, domainID int64, rest []string) {
	records := m.records[domainID]

	if len(rest) == 0 {
		switch r.Method {
		case http.MethodGet:
			host := r.URL.Query().Get("host")
			typ := r.URL.Query().Get("type")
			ids := make([]int64, 0, len(records))
			for id, rec := range records {
				if host != "" && rec["host"] != host {
					continue
				}
				if typ != "" && rec["type"] != typ {
					continue
				}
				ids = append(ids, id)
			}
			sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
			list := make([]interface{}, 0, len(ids))
			for _, id := range ids {
				list = append(list, records[id])
			}
			m.writeJSON(w, http.StatusOK, list)
		case http.MethodPost:
			rec, ok := m.decodeRecord(w, r)
			if !ok {
				return
			}
			m.nextID++
			id := m.nextID
			rec["id"] = id
			records[id] = rec
			w.Header().Set("Location", fmt.Sprintf("/v0/domains/%d/dns/%d", domainID, id))
			m.writeJSON(w, http.StatusCreated, map[string]interface{}{"id": id})
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}

	recordID, err := strconv.ParseInt(rest[0], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	existing := records[recordID]
	switch r.Method {
	case http.MethodGet:
		if existing == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		m.writeJSON(w, http.StatusOK, existing)
	case http.MethodPut:
		if existing == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		rec, ok := m.decodeRecord(w, r)
		if !ok {
			return
		}
		rec["id"] = recordID
		records[recordID] = rec
		w.WriteHeader(http.StatusNoContent)
	case http.MethodDelete:
		if existing == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		delete(records, recordID)
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (m *mockDomeneshop) handleForwards(w http.ResponseWriter, r *http.Request, domainID int64, rest []string) {
	forwards := m.forwards[domainID]

	if len(rest) == 0 {
		switch r.Method {
		case http.MethodGet:
			hosts := make([]string, 0, len(forwards))
			for h := range forwards {
				hosts = append(hosts, h)
			}
			sort.Strings(hosts)
			list := make([]interface{}, 0, len(hosts))
			for _, h := range hosts {
				list = append(list, forwards[h])
			}
			m.writeJSON(w, http.StatusOK, list)
		case http.MethodPost:
			fwd, ok := m.decodeBody(w, r)
			if !ok {
				return
			}
			if _, ok := fwd["frame"]; !ok {
				// The real API always returns the frame flag, defaulting to false.
				fwd["frame"] = false
			}
			host, _ := fwd["host"].(string)
			if _, exists := forwards[host]; exists {
				w.WriteHeader(http.StatusConflict)
				return
			}
			forwards[host] = fwd
			w.Header().Set("Location", fmt.Sprintf("/v0/domains/%d/forwards/%s", domainID, host))
			w.WriteHeader(http.StatusCreated)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}

	host := rest[0]
	existing := forwards[host]
	switch r.Method {
	case http.MethodGet:
		if existing == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		m.writeJSON(w, http.StatusOK, existing)
	case http.MethodPut:
		if existing == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		fwd, ok := m.decodeBody(w, r)
		if !ok {
			return
		}
		fwd["host"] = host
		forwards[host] = fwd
		m.writeJSON(w, http.StatusOK, fwd)
	case http.MethodDelete:
		if existing == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		delete(forwards, host)
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// decodeRecord decodes a DNS record body, dropping the client's spurious
// zero-value "id" and applying the documented TTL default.
func (m *mockDomeneshop) decodeRecord(w http.ResponseWriter, r *http.Request) (map[string]interface{}, bool) {
	rec, ok := m.decodeBody(w, r)
	if !ok {
		return nil, false
	}
	delete(rec, "id")
	if _, ok := rec["ttl"]; !ok {
		rec["ttl"] = 3600
	}
	return rec, true
}

func (m *mockDomeneshop) decodeBody(w http.ResponseWriter, r *http.Request) (map[string]interface{}, bool) {
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil, false
	}
	return body, true
}

func (m *mockDomeneshop) writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
