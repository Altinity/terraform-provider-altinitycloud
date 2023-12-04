package env_test

import (
	"fmt"
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/test"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var dataEnvName = test.GenerateRandomEnvName()

func TestAccAltinityCloudEnvK8SDataSource_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: GetK8SEnvDatasource(dataEnvName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data."+FILE_NAME, "name", dataEnvName),
				),
			},
		},
	})
}

func GetK8SEnvDatasource(envName string) string {
	return fmt.Sprintf(`
%s

data "%[2]s" "dummy" {
	name = %[2]s.dummy.name
}

`, GetK8SEnvResource(envName), RESOURCE_NAME)
}
