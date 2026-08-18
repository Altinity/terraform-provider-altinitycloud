package env_status

import (
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/schematest"
)

func TestAzureEnvStatusDataSourceModelMatchesSchema(t *testing.T) {
	schematest.AssertDataSourceModelMatchesSchema(t, &AzureEnvStatusDataSource{}, &AzureEnvStatusModel{})
}
