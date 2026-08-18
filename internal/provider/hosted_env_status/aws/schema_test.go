package hosted_env_status

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestHostedAWSEnvStatusDataSource_Schema_validateImplementation(t *testing.T) {
	t.Parallel()

	var ds HostedAWSEnvStatusDataSource
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if diags := resp.Schema.ValidateImplementation(context.Background()); diags.HasError() {
		t.Fatalf("schema validation failed: %v", diags)
	}
}
