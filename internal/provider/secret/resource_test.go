package secret_test

import (
	"fmt"
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/test"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const RESOURCE_NAME = "altinitycloud_secret"
const FILE_NAME = RESOURCE_NAME + ".dummy"

var resourceEnvName = test.GenerateRandomEnvName()

func TestAccAltinityCloudCertificate_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: GetK8SEnvResource(resourceEnvName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAltinityCloudCertificateExists(FILE_NAME),
					resource.TestCheckResourceAttr(FILE_NAME, "pem", resourceEnvName),
				),
			},
		},
	})
}

func testAccCheckAltinityCloudCertificateExists(n string) resource.TestCheckFunc {
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

func GetK8SEnvResource(envName string) string {
	return fmt.Sprintf(`
resource "altinitycloud_env_k8s" "dummy" {
	name         = "%s"
	distribution = "CUSTOM"
	node_groups  = [{
		zones             = ["us-east-1a"]
		node_type         = "small"
		capacity_per_zone = 1
		reservations      = ["SYSTEM","CLICKHOUSE","ZOOKEEPER"]
	}]

	skip_deprovision_on_destroy = true
	force_destroy               = true
}

resource "%s" "dummy" {
  env_name = altinitycloud_env_k8s.dummy.name
}
`, envName, RESOURCE_NAME)
}
