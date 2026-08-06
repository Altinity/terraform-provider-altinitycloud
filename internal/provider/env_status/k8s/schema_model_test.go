package env_status

import (
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/schematest"
)

func TestK8SEnvStatusDataSourceModelMatchesSchema(t *testing.T) {
	schematest.AssertDataSourceModelMatchesSchema(t, &K8SEnvStatusDataSource{}, &K8SEnvStatusModel{})
}
