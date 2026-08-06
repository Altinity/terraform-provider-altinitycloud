package env

import (
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/schematest"
)

func TestK8SResourceModelMatchesSchema(t *testing.T) {
	schematest.AssertResourceModelMatchesSchema(t, &K8SEnvResource{}, &K8SEnvResourceModel{})
}

func TestK8SDataSourceModelMatchesSchema(t *testing.T) {
	schematest.AssertDataSourceModelMatchesSchema(t, &K8SEnvDataSource{}, &K8SEnvDataSourceModel{})
}
