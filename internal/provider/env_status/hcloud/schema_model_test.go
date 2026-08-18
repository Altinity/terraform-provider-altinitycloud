package env_status

import (
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/schematest"
)

func TestHCloudEnvStatusDataSourceModelMatchesSchema(t *testing.T) {
	schematest.AssertDataSourceModelMatchesSchema(t, &HCloudEnvStatusDataSource{}, &HCloudEnvStatusModel{})
}
