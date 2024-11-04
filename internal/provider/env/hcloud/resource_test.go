package env_test

import (
	"fmt"
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/test"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const RESOURCE_NAME = "altinitycloud_env_hcloud"
const FILE_NAME = RESOURCE_NAME + ".dummy"

var resourceEnvName = test.GenerateRandomEnvName()

func TestAccAltinityCloudEnvHCloud_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: GetHCloudEnvResource(resourceEnvName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAltinityCloudEnvHCloudExists(FILE_NAME),
					resource.TestCheckResourceAttr(FILE_NAME, "name", resourceEnvName),
				),
			},
		},
	})
}

func testAccCheckAltinityCloudEnvHCloudExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no resource ID is set")
		}

		return nil
	}
}

func GetHCloudEnvResource(envName string) string {
	return fmt.Sprintf(`
resource "%s" "dummy" {
  name           = "%s"
	cidr           = "10.0.0.0/16"
	network_zone         = "us-east"
	locations          = ["hil"]

	node_groups = [{
		locations             = ["hil"]
		node_type         = "c2-standard-16"
		capacity_per_zone = 1
		reservations      = ["SYSTEM","CLICKHOUSE","ZOOKEEPER"]
	}]

	skip_deprovision_on_destroy = true
	force_destroy               = true
}
`, RESOURCE_NAME, envName)
}
