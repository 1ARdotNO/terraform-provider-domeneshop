package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceRecord_A(t *testing.T) {
	domain := os.Getenv("DOMENESHOP_DOMAIN")
	host := acctest.RandomWithPrefix("test")
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckRecordDestroy,
		Steps: []resource.TestStep{
			{
				// test create
				Config: testAccResourceRecordConfig(domain, host, "A", "192.0.2.56"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("domeneshop_record.test", "host", host),
					resource.TestCheckResourceAttr("domeneshop_record.test", "type", "A"),
					resource.TestCheckResourceAttr("domeneshop_record.test", "data", "192.0.2.56"),
					resource.TestCheckResourceAttr("domeneshop_record.test", "ttl", "300"),
				),
			},
			{
				// test update
				Config: testAccResourceRecordConfig(domain, host, "A", "192.0.2.57"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("domeneshop_record.test", "host", host),
					resource.TestCheckResourceAttr("domeneshop_record.test", "type", "A"),
					resource.TestCheckResourceAttr("domeneshop_record.test", "data", "192.0.2.57"),
					resource.TestCheckResourceAttr("domeneshop_record.test", "ttl", "300"),
				),
			},
			{
				// test import
				ResourceName:      "domeneshop_record.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["domeneshop_record.test"]
					if !ok {
						return "", fmt.Errorf("resource not found in state")
					}
					return fmt.Sprintf("%s/%s", rs.Primary.Attributes["domain_id"], rs.Primary.ID), nil
				},
			},
		},
	})
}

func TestAccResourceRecord_AAAA(t *testing.T) {
	domain := os.Getenv("DOMENESHOP_DOMAIN")
	host := acctest.RandomWithPrefix("test")
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckRecordDestroy,
		Steps: []resource.TestStep{
			{
				// test create
				Config: testAccResourceRecordConfig(domain, host, "AAAA", "100::"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("domeneshop_record.test", "host", host),
					resource.TestCheckResourceAttr("domeneshop_record.test", "type", "AAAA"),
					resource.TestCheckResourceAttr("domeneshop_record.test", "data", "100::"),
					resource.TestCheckResourceAttr("domeneshop_record.test", "ttl", "300"),
				),
			},
		},
	})
}

func TestAccResourceRecord_CAA(t *testing.T) {
	domain := os.Getenv("DOMENESHOP_DOMAIN")
	host := acctest.RandomWithPrefix("test")
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRecordCAAConfig(domain, host),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("domeneshop_record.test", "host", host),
					resource.TestCheckResourceAttr("domeneshop_record.test", "type", "CAA"),
					resource.TestMatchResourceAttr("domeneshop_record.test", "data", regexp.MustCompile("^test")),
					resource.TestCheckResourceAttr("domeneshop_record.test", "ttl", "300"),
					resource.TestCheckResourceAttr("domeneshop_record.test", "flags", "128"),
					resource.TestCheckResourceAttr("domeneshop_record.test", "tag", "1"),
				),
			},
		},
	})
}

func TestAccResourceRecord_CNAME(t *testing.T) {
	domain := os.Getenv("DOMENESHOP_DOMAIN")
	host := acctest.RandomWithPrefix("test")
	data := fmt.Sprintf("%s.%s.", host, domain)
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRecordConfig(domain, host, "CNAME", data),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("domeneshop_record.test", "host", host),
					resource.TestCheckResourceAttr("domeneshop_record.test", "type", "CNAME"),
					resource.TestCheckResourceAttr("domeneshop_record.test", "data", data),
					resource.TestCheckResourceAttr("domeneshop_record.test", "ttl", "300"),
				),
			},
		},
	})
}

func TestAccResourceRecord_MX(t *testing.T) {
	domain := os.Getenv("DOMENESHOP_DOMAIN")
	host := acctest.RandomWithPrefix("test")
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRecordMXConfig(domain, host),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("domeneshop_record.test", "host", host),
					resource.TestCheckResourceAttr("domeneshop_record.test", "type", "MX"),
					resource.TestMatchResourceAttr("domeneshop_record.test", "data", regexp.MustCompile("^test")),
					resource.TestCheckResourceAttr("domeneshop_record.test", "ttl", "300"),
					resource.TestCheckResourceAttr("domeneshop_record.test", "priority", "10"),
				),
			},
		},
	})
}

func TestAccResourceRecord_NS(t *testing.T) {
	domain := os.Getenv("DOMENESHOP_DOMAIN")
	host := acctest.RandomWithPrefix("test")
	data := fmt.Sprintf("%s.%s.", host, domain)
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRecordConfig(domain, host, "NS", data),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("domeneshop_record.test", "host", host),
					resource.TestCheckResourceAttr("domeneshop_record.test", "type", "NS"),
					resource.TestCheckResourceAttr("domeneshop_record.test", "data", data),
					resource.TestCheckResourceAttr("domeneshop_record.test", "ttl", "300"),
				),
			},
		},
	})
}

func TestAccResourceRecord_SRV(t *testing.T) {
	domain := os.Getenv("DOMENESHOP_DOMAIN")
	host := acctest.RandomWithPrefix("test")
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRecordSRVConfig(domain, host),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("domeneshop_record.test", "host", regexp.MustCompile("^_sip")),
					resource.TestCheckResourceAttr("domeneshop_record.test", "type", "SRV"),
					resource.TestMatchResourceAttr("domeneshop_record.test", "data", regexp.MustCompile("^test")),
					resource.TestCheckResourceAttr("domeneshop_record.test", "ttl", "300"),
					resource.TestCheckResourceAttr("domeneshop_record.test", "priority", "10"),
					resource.TestCheckResourceAttr("domeneshop_record.test", "weight", "60"),
					resource.TestCheckResourceAttr("domeneshop_record.test", "port", "5060"),
				),
			},
		},
	})
}

func TestAccResourceRecord_TLSA(t *testing.T) {
	domain := os.Getenv("DOMENESHOP_DOMAIN")
	host := acctest.RandomWithPrefix("_443._tcp.test")
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRecordTLSAConfig(domain, host),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("domeneshop_record.test", "host", regexp.MustCompile("^_443")),
					resource.TestCheckResourceAttr("domeneshop_record.test", "type", "TLSA"),
					resource.TestCheckResourceAttr("domeneshop_record.test", "data", "b35e305e00c2b663444a9e3f0f36fdb876a2a9f9822a44d4954810aed290f5c2"),
					resource.TestCheckResourceAttr("domeneshop_record.test", "ttl", "300"),
					resource.TestCheckResourceAttr("domeneshop_record.test", "usage", "3"),
					resource.TestCheckResourceAttr("domeneshop_record.test", "selector", "1"),
					resource.TestCheckResourceAttr("domeneshop_record.test", "dtype", "1"),
				),
			},
		},
	})
}

func testAccResourceRecordConfig(domain string, host string, recordType string, data string) string {
	return fmt.Sprintf(`
data "domeneshop_domains" "test" {
  domain = "%s"
}

resource "domeneshop_record" "test" {
  domain_id = data.domeneshop_domains.test.domains[0].id
  host      = "%s"
  type      = "%s"
  data      = "%s"
  ttl       = 300
}
`, domain, host, recordType, data)
}

func testAccResourceRecordCAAConfig(domain string, host string) string {
	return fmt.Sprintf(`
data "domeneshop_domains" "test" {
  domain = "%s"
}

resource "domeneshop_record" "test" {
  domain_id = data.domeneshop_domains.test.domains[0].id
  host      = "%s"
  type      = "CAA"
  data      = "%[2]s.%[1]s."
  ttl       = 300
  flags     = 128
  tag       = 1
}
`, domain, host)
}

func testAccResourceRecordMXConfig(domain string, host string) string {
	return fmt.Sprintf(`
data "domeneshop_domains" "test" {
  domain = "%s"
}

resource "domeneshop_record" "test" {
  domain_id = data.domeneshop_domains.test.domains[0].id
  host      = "%s"
  type      = "MX"
  data      = "%[2]s.%[1]s."
  ttl       = 300
  priority  = 10
}
`, domain, host)
}

func testAccResourceRecordSRVConfig(domain string, host string) string {
	return fmt.Sprintf(`
data "domeneshop_domains" "test" {
  domain = "%s"
}

resource "domeneshop_record" "test" {
  domain_id = data.domeneshop_domains.test.domains[0].id
  host      = "_sip._tcp.%s"
  type      = "SRV"
  data      = "%[2]s.%[1]s."
  ttl       = 300
  priority  = 10
  weight    = 60
  port      = 5060
}
`, domain, host)
}

func testAccResourceRecordTLSAConfig(domain string, host string) string {
	return fmt.Sprintf(`
data "domeneshop_domains" "test" {
  domain = "%s"
}

resource "domeneshop_record" "test" {
  domain_id = data.domeneshop_domains.test.domains[0].id
  host      = "%s"
  type      = "TLSA"
  data      = "b35e305e00c2b663444a9e3f0f36fdb876a2a9f9822a44d4954810aed290f5c2"
  ttl       = 300
  usage     = 3
  selector  = 1
  dtype     = 1
}
`, domain, host)
}
