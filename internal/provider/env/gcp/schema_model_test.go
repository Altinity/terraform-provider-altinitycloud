package env

import (
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/schematest"
)

func TestGCPResourceModelMatchesSchema(t *testing.T) {
	schematest.AssertResourceModelMatchesSchema(t, &GCPEnvResource{}, &GCPEnvResourceModel{})
}

func TestGCPDataSourceModelMatchesSchema(t *testing.T) {
	schematest.AssertDataSourceModelMatchesSchema(t, &GCPEnvDataSource{}, &GCPEnvDataSourceModel{})
}
