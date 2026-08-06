package env_status

import (
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/schematest"
)

func TestGCPEnvStatusDataSourceModelMatchesSchema(t *testing.T) {
	schematest.AssertDataSourceModelMatchesSchema(t, &GCPEnvStatusDataSource{}, &GCPEnvStatusModel{})
}
