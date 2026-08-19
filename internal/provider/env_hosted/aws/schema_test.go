package hosted_env

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestAWSEnvHostedResource_Schema_validateImplementation(t *testing.T) {
	t.Parallel()

	var r AWSEnvHostedResource
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	if diags := resp.Schema.ValidateImplementation(context.Background()); diags.HasError() {
		t.Fatalf("schema validation failed: %v", diags)
	}
}

func TestAWSEnvHostedDataSource_Schema_validateImplementation(t *testing.T) {
	t.Parallel()

	var ds AWSEnvHostedDataSource
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if diags := resp.Schema.ValidateImplementation(context.Background()); diags.HasError() {
		t.Fatalf("schema validation failed: %v", diags)
	}
}
