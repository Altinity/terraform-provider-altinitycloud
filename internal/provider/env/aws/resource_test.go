package env_test

import (
	"fmt"
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/test"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const RESOURCE_NAME = "altinitycloud_env_aws"
const FILE_NAME = RESOURCE_NAME + ".dummy"

var resourceEnvName = test.GenerateRandomEnvName()

func TestAccAltinityCloudEnvAWS_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: GetAWSEnvResource(resourceEnvName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAltinityCloudEnvAWSExists(FILE_NAME),
					resource.TestCheckResourceAttr(FILE_NAME, "name", resourceEnvName),
				),
			},
		},
	})
}

func testAccCheckAltinityCloudEnvAWSExists(n string) resource.TestCheckFunc {
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

func GetAWSEnvResource(envName string) string {
	return fmt.Sprintf(`
resource "%s" "dummy" {
  name           = "%s"
  cidr           = "10.0.0.0/16"
  region         = "us-east-1"
  aws_account_id = "123456789012"
  zones          = ["us-east-1a", "us-east-1b"]

  node_groups = [{
    zones             = ["us-east-1a"]
    node_type         = "t4g.large"
    capacity_per_zone = 1
    reservations      = ["SYSTEM","CLICKHOUSE","ZOOKEEPER"]
  }]

  force_destroy                   = true
  skip_deprovision_on_destroy     = true
  allow_delete_while_disconnected = true
}
`, RESOURCE_NAME, envName)
}
