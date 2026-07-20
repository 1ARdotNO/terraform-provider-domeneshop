package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceForward_basic(t *testing.T) {
	domain := os.Getenv("DOMENESHOP_DOMAIN")
	host := acctest.RandString(6)
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckForwardDestroy,
		Steps: []resource.TestStep{
			{
				// test create
				Config: testAccResourceForwardConfig(domain, host, "https://example.com/foo"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("domeneshop_forward.test", "host", host),
					resource.TestCheckResourceAttr("domeneshop_forward.test", "url", "https://example.com/foo"),
					resource.TestCheckResourceAttr("domeneshop_forward.test", "frame", "false"),
				),
			},
			{
				// test update
				Config: testAccResourceForwardConfig(domain, host, "https://example.com/bar"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("domeneshop_forward.test", "host", host),
					resource.TestCheckResourceAttr("domeneshop_forward.test", "url", "https://example.com/bar"),
					resource.TestCheckResourceAttr("domeneshop_forward.test", "frame", "false"),
				),
			},
			{
				// test import
				ResourceName:      "domeneshop_forward.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["domeneshop_forward.test"]
					if !ok {
						return "", fmt.Errorf("resource not found in state")
					}
					return fmt.Sprintf("%s/%s", rs.Primary.Attributes["domain_id"], rs.Primary.ID), nil
				},
			},
		},
	})
}

func testAccResourceForwardConfig(domain string, host string, url string) string {
	return fmt.Sprintf(`
data "domeneshop_domains" "test" {
  domain = "%s"
}

resource "domeneshop_forward" "test" {
  domain_id = data.domeneshop_domains.test.domains[0].id
  host      = "%s"
  url       = "%s"
}
`, domain, host, url)
}
