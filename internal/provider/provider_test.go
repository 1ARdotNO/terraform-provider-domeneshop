package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// TestMain starts an in-process mock of the Domeneshop API (validated against
// the vendored OpenAPI documentation, see mockapi_test.go) unless real API
// credentials are provided via DOMENESHOP_TOKEN. This makes the acceptance
// tests hermetic by default: `TF_ACC=1 go test ./internal/provider/` passes
// without any credentials or network access to Domeneshop.
func TestMain(m *testing.M) {
	if os.Getenv("DOMENESHOP_TOKEN") == "" {
		srv, err := newMockDomeneshop()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to start mock Domeneshop API: %v\n", err)
			os.Exit(1)
		}
		os.Setenv("DOMENESHOP_TOKEN", "test-token")
		os.Setenv("DOMENESHOP_SECRET", "test-secret")
		os.Setenv("DOMENESHOP_DOMAIN", "example.com")
		os.Setenv("DOMENESHOP_DOMAIN_ID", "1")
		os.Setenv("DOMENESHOP_HOST", srv.URL)
		code := m.Run()
		srv.Close()
		os.Exit(code)
	}
	os.Exit(m.Run())
}

// providerFactories are used to instantiate a provider during acceptance testing.
// The factory function will be invoked for every Terraform CLI command executed
// to create a provider server to which the CLI can reattach.
var providerFactories = map[string]func() (*schema.Provider, error){
	"domeneshop": func() (*schema.Provider, error) {
		return New("dev")(), nil
	},
}

func TestProvider(t *testing.T) {
	if err := New("dev")().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("DOMENESHOP_TOKEN"); v == "" {
		t.Fatal("DOMENESHOP_TOKEN must be set for acceptance tests")
	}
	if v := os.Getenv("DOMENESHOP_SECRET"); v == "" {
		t.Fatal("DOMENESHOP_SECRET must be set for acceptance tests")
	}
	if v := os.Getenv("DOMENESHOP_DOMAIN"); v == "" {
		t.Fatal("DOMENESHOP_DOMAIN must be set for acceptance tests")
	}
}
