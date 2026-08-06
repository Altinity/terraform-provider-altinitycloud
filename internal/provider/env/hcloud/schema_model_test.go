package env

import (
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/schematest"
)

func TestHCloudResourceModelMatchesSchema(t *testing.T) {
	schematest.AssertResourceModelMatchesSchema(t, &HCloudEnvResource{}, &HCloudEnvResourceModel{})
}

func TestHCloudDataSourceModelMatchesSchema(t *testing.T) {
	schematest.AssertDataSourceModelMatchesSchema(t, &HCloudEnvDataSource{}, &HCloudEnvDataSourceModel{})
}
