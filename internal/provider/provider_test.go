package provider

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/innovationnorway/go-domeneshop/api/v0/domeneshop"
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

// testAccAPIClient returns a Domeneshop API client configured from the same
// environment variables the provider under test uses, for out-of-band checks
// such as CheckDestroy.
func testAccAPIClient() (*domeneshop.APIClient, context.Context) {
	config := domeneshop.NewConfiguration()
	if host := os.Getenv("DOMENESHOP_HOST"); host != "" {
		config.Servers = domeneshop.ServerConfigurations{
			{URL: host},
		}
	}
	auth := context.WithValue(context.Background(), domeneshop.ContextBasicAuth, domeneshop.BasicAuth{
		UserName: os.Getenv("DOMENESHOP_TOKEN"),
		Password: os.Getenv("DOMENESHOP_SECRET"),
	})
	return domeneshop.NewAPIClient(config), auth
}

func testAccCheckRecordDestroy(s *terraform.State) error {
	client, auth := testAccAPIClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "domeneshop_record" {
			continue
		}
		domainID, _ := strconv.Atoi(rs.Primary.Attributes["domain_id"])
		recordID, _ := strconv.Atoi(rs.Primary.ID)
		_, r, err := client.DnsApi.GetRecord(auth, int32(domainID), int32(recordID)).Execute()
		if err.Error() == "" {
			return fmt.Errorf("DNS record %s still exists", rs.Primary.ID)
		}
		if r == nil || r.StatusCode != 404 {
			return fmt.Errorf("unexpected error checking DNS record %s: %s", rs.Primary.ID, err.Error())
		}
	}
	return nil
}

func testAccCheckForwardDestroy(s *terraform.State) error {
	client, auth := testAccAPIClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "domeneshop_forward" {
			continue
		}
		domainID, _ := strconv.Atoi(rs.Primary.Attributes["domain_id"])
		_, r, err := client.ForwardsApi.GetForward(auth, int32(domainID), rs.Primary.ID).Execute()
		if err.Error() == "" {
			return fmt.Errorf("HTTP forward %s still exists", rs.Primary.ID)
		}
		if r == nil || r.StatusCode != 404 {
			return fmt.Errorf("unexpected error checking HTTP forward %s: %s", rs.Primary.ID, err.Error())
		}
	}
	return nil
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
