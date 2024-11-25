package secret_test

import (
	"fmt"
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/test"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const RESOURCE_NAME = "altinitycloud_env_secret"
const FILE_NAME = RESOURCE_NAME + ".dummy"

func TestAccAltinityCloudSecret_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: GeSecretResource(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAltinityCloudSecretExists(FILE_NAME),
					resource.TestCheckResourceAttr(FILE_NAME, "value", "value"),
				),
			},
		},
	})
}

func testAccCheckAltinityCloudSecretExists(n string) resource.TestCheckFunc {
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

func GeSecretResource() string {
	return fmt.Sprintf(`
resource "%s" "dummy" {
  pem   = "xxx"
  value = "value"
}
`, RESOURCE_NAME)
}
