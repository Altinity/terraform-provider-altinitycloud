package env_status

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestAWSEnvStatusDataSource_Schema_validateImplementation(t *testing.T) {
	t.Parallel()

	var ds AWSEnvStatusDataSource
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	diags := resp.Schema.ValidateImplementation(context.Background())
	if diags.HasError() {
		t.Fatalf("schema validation failed: %v", diags)
	}
}
