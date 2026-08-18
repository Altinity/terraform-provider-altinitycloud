package env_status

import (
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/schematest"
)

func TestAWSEnvStatusDataSourceModelMatchesSchema(t *testing.T) {
	schematest.AssertDataSourceModelMatchesSchema(t, &AWSEnvStatusDataSource{}, &AWSEnvStatusModel{})
}
