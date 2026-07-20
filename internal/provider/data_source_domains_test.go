package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceDomains_basic(t *testing.T) {
	domain := os.Getenv("DOMENESHOP_DOMAIN")
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceDomainsConfig(domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.domeneshop_domains.test", "domains.0.domain", domain),
					resource.TestCheckResourceAttrSet("data.domeneshop_domains.test", "domains.0.expiry_date"),
					resource.TestCheckResourceAttrSet("data.domeneshop_domains.test", "domains.0.renew"),
					resource.TestCheckResourceAttrSet("data.domeneshop_domains.test", "domains.0.registrant"),
					resource.TestCheckResourceAttrSet("data.domeneshop_domains.test", "domains.0.status"),
					resource.TestCheckResourceAttrSet("data.domeneshop_domains.test", "domains.0.nameservers.0"),
					resource.TestCheckResourceAttrSet("data.domeneshop_domains.test", "domains.0.services.0.registrar"),
					resource.TestCheckResourceAttrSet("data.domeneshop_domains.test", "domains.0.services.0.dns"),
					resource.TestCheckResourceAttrSet("data.domeneshop_domains.test", "domains.0.services.0.email"),
					resource.TestCheckResourceAttrSet("data.domeneshop_domains.test", "domains.0.services.0.webhotel"),
				),
			},
		},
	})
}

func testAccDataSourceDomainsConfig(domain string) string {
	return fmt.Sprintf(`
data "domeneshop_domains" "test" {
  domain = "%s"
}
`, domain)
}
