// Package schematest asserts that a resource/data source model struct matches
// its schema. A mismatch is invisible at compile time and crashes every plan or
// read with "mismatch between struct and object", so every pair gets a test.
package schematest

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type SchemaResource interface {
	Schema(context.Context, resource.SchemaRequest, *resource.SchemaResponse)
}

type SchemaDataSource interface {
	Schema(context.Context, datasource.SchemaRequest, *datasource.SchemaResponse)
}

func AssertResourceModelMatchesSchema(t *testing.T, r SchemaResource, model any) {
	t.Helper()
	ctx := context.Background()

	resp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, resp)
	assertNoDiags(t, resp.Diagnostics)
	assertNoDiags(t, resp.Schema.ValidateImplementation(ctx))

	state := tfsdk.State{Schema: resp.Schema, Raw: populated(t, resp.Schema.Type().TerraformType(ctx))}
	assertNoDiags(t, state.Get(ctx, model))
}

func AssertDataSourceModelMatchesSchema(t *testing.T, d SchemaDataSource, model any) {
	t.Helper()
	ctx := context.Background()

	resp := &datasource.SchemaResponse{}
	d.Schema(ctx, datasource.SchemaRequest{}, resp)
	assertNoDiags(t, resp.Diagnostics)
	assertNoDiags(t, resp.Schema.ValidateImplementation(ctx))

	state := tfsdk.State{Schema: resp.Schema, Raw: populated(t, resp.Schema.Type().TerraformType(ctx))}
	assertNoDiags(t, state.Get(ctx, model))
}

// Containers are built known and non-empty on purpose: the framework stops
// descending at a null value, so an all-null fixture would only ever check the
// top-level attributes and miss a mismatch inside a nested model.
func populated(t *testing.T, typ tftypes.Type) tftypes.Value {
	t.Helper()

	switch tt := typ.(type) {
	case tftypes.Object:
		vals := make(map[string]tftypes.Value, len(tt.AttributeTypes))
		for name, attrType := range tt.AttributeTypes {
			vals[name] = populated(t, attrType)
		}
		return tftypes.NewValue(tt, vals)
	case tftypes.List:
		return tftypes.NewValue(tt, []tftypes.Value{populated(t, tt.ElementType)})
	case tftypes.Set:
		return tftypes.NewValue(tt, []tftypes.Value{populated(t, tt.ElementType)})
	case tftypes.Tuple:
		vals := make([]tftypes.Value, 0, len(tt.ElementTypes))
		for _, elemType := range tt.ElementTypes {
			vals = append(vals, populated(t, elemType))
		}
		return tftypes.NewValue(tt, vals)
	case tftypes.Map:
		return tftypes.NewValue(tt, map[string]tftypes.Value{"key": populated(t, tt.ElementType)})
	default:
		return tftypes.NewValue(typ, nil)
	}
}

func assertNoDiags(t *testing.T, diags diag.Diagnostics) {
	t.Helper()
	for _, d := range diags {
		t.Errorf("%s: %s", d.Summary(), d.Detail())
	}
}
