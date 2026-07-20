package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceDomain_basic(t *testing.T) {
	domain := os.Getenv("DOMENESHOP_DOMAIN")
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceDomainConfig(domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.domeneshop_domain.test", "domain", domain),
					resource.TestCheckResourceAttrSet("data.domeneshop_domain.test", "expiry_date"),
					resource.TestCheckResourceAttrSet("data.domeneshop_domain.test", "renew"),
					resource.TestCheckResourceAttrSet("data.domeneshop_domain.test", "registrant"),
					resource.TestCheckResourceAttrSet("data.domeneshop_domain.test", "status"),
					resource.TestCheckResourceAttrSet("data.domeneshop_domain.test", "nameservers.0"),
					resource.TestCheckResourceAttrSet("data.domeneshop_domain.test", "services.0.registrar"),
					resource.TestCheckResourceAttrSet("data.domeneshop_domain.test", "services.0.dns"),
					resource.TestCheckResourceAttrSet("data.domeneshop_domain.test", "services.0.email"),
					resource.TestCheckResourceAttrSet("data.domeneshop_domain.test", "services.0.webhotel"),
				),
			},
		},
	})
}

func testAccDataSourceDomainConfig(domain string) string {
	return fmt.Sprintf(`
data "domeneshop_domains" "test" {
  domain = "%s"
}

data "domeneshop_domain" "test" {
  domain_id = data.domeneshop_domains.test.domains[0].id
}
`, domain)
}
