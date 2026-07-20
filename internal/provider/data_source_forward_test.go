package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceForward_basic(t *testing.T) {
	domain := os.Getenv("DOMENESHOP_DOMAIN")
	host := acctest.RandString(6)
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceForwardConfig(domain, host),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.domeneshop_forward.test", "host", host),
					resource.TestCheckResourceAttr("data.domeneshop_forward.test", "url", "https://example.com/foo"),
					resource.TestCheckResourceAttr("data.domeneshop_forward.test", "frame", "false"),
				),
			},
		},
	})
}

func testAccDataSourceForwardConfig(domain string, host string) string {
	return fmt.Sprintf(`
data "domeneshop_domains" "test" {
  domain = "%s"
}

resource "domeneshop_forward" "test" {
  domain_id = data.domeneshop_domains.test.domains[0].id
  host      = "%s"
  url       = "https://example.com/foo"
}

data "domeneshop_forward" "test" {
  domain_id  = data.domeneshop_domains.test.domains[0].id
  host       = domeneshop_forward.test.host
  depends_on = [domeneshop_forward.test]
}
`, domain, host)
}
