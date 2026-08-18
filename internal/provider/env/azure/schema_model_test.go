package env

import (
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/schematest"
)

func TestAzureResourceModelMatchesSchema(t *testing.T) {
	schematest.AssertResourceModelMatchesSchema(t, &AzureEnvResource{}, &AzureEnvResourceModel{})
}

func TestAzureDataSourceModelMatchesSchema(t *testing.T) {
	schematest.AssertDataSourceModelMatchesSchema(t, &AzureEnvDataSource{}, &AzureEnvDataSourceModel{})
}
