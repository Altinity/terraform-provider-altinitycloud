package env

import (
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/schematest"
)

func TestAWSResourceModelMatchesSchema(t *testing.T) {
	schematest.AssertResourceModelMatchesSchema(t, &AWSEnvResource{}, &AWSEnvResourceModel{})
}

func TestAWSDataSourceModelMatchesSchema(t *testing.T) {
	schematest.AssertDataSourceModelMatchesSchema(t, &AWSEnvDataSource{}, &AWSEnvDataSourceModel{})
}
