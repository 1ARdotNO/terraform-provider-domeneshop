package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceForwards_basic(t *testing.T) {
	domain := os.Getenv("DOMENESHOP_DOMAIN")
	host := acctest.RandString(6)
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceForwardsConfig(domain, host),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.domeneshop_forwards.test", "forwards.0.host"),
					resource.TestMatchResourceAttr("data.domeneshop_forwards.test", "forwards.0.url", regexp.MustCompile("^http")),
					resource.TestCheckResourceAttr("data.domeneshop_forwards.test", "forwards.0.frame", "false"),
				),
			},
		},
	})
}

func testAccDataSourceForwardsConfig(domain string, host string) string {
	return fmt.Sprintf(`
data "domeneshop_domains" "test" {
  domain = "%s"
}

resource "domeneshop_forward" "test" {
  domain_id = data.domeneshop_domains.test.domains[0].id
  host      = "%s"
  url       = "https://example.com/foo"
}

data "domeneshop_forwards" "test" {
  domain_id  = data.domeneshop_domains.test.domains[0].id
  depends_on = [domeneshop_forward.test]
}
`, domain, host)
}
