package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceRecords_basic(t *testing.T) {
	domain := os.Getenv("DOMENESHOP_DOMAIN")
	host := acctest.RandomWithPrefix("test")
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceRecordsConfig(domain, host),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.domeneshop_records.test", "records.0.host", host),
					resource.TestCheckResourceAttr("data.domeneshop_records.test", "records.0.type", "A"),
					resource.TestCheckResourceAttr("data.domeneshop_records.test", "records.0.data", "192.0.2.56"),
					resource.TestCheckResourceAttr("data.domeneshop_records.test", "records.0.ttl", "300"),
				),
			},
		},
	})
}

func testAccDataSourceRecordsConfig(domain string, host string) string {
	return fmt.Sprintf(`
data "domeneshop_domains" "test" {
  domain = "%s"
}

resource "domeneshop_record" "test" {
  domain_id = data.domeneshop_domains.test.domains[0].id
  host      = "%s"
  type      = "A"
  data      = "192.0.2.56"
  ttl       = 300
}

data "domeneshop_records" "test" {
  domain_id  = data.domeneshop_domains.test.domains[0].id
  host       = domeneshop_record.test.host
  depends_on = [domeneshop_record.test]
}
`, domain, host)
}
