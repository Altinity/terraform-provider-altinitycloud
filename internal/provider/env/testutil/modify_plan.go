package testutil

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type SpecRevResource interface {
	Schema(context.Context, resource.SchemaRequest, *resource.SchemaResponse)
	ModifyPlan(context.Context, resource.ModifyPlanRequest, *resource.ModifyPlanResponse)
}

func AssertModifyPlanSpecRevision(t *testing.T, r SpecRevResource) {
	t.Helper()
	ctx := context.Background()

	sresp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, sresp)
	schema := sresp.Schema

	objType, ok := schema.Type().TerraformType(ctx).(tftypes.Object)
	if !ok {
		t.Fatal("schema type is not a tftypes.Object")
	}

	build := func(customDomain string) tftypes.Value {
		vals := map[string]tftypes.Value{}
		for n, at := range objType.AttributeTypes {
			switch n {
			case "name", "id":
				vals[n] = tftypes.NewValue(at, "env")
			case "spec_revision":
				vals[n] = tftypes.NewValue(at, int64(100))
			case "custom_domain":
				vals[n] = tftypes.NewValue(at, customDomain)
			default:
				vals[n] = tftypes.NewValue(at, nil)
			}
		}
		return tftypes.NewValue(objType, vals)
	}

	run := func(state, plan tftypes.Value) types.Int64 {
		resp := &resource.ModifyPlanResponse{Plan: tfsdk.Plan{Raw: plan, Schema: schema}}
		r.ModifyPlan(ctx, resource.ModifyPlanRequest{
			State: tfsdk.State{Raw: state, Schema: schema},
			Plan:  tfsdk.Plan{Raw: plan, Schema: schema},
		}, resp)
		if resp.Diagnostics.HasError() {
			t.Fatalf("ModifyPlan errored: %v", resp.Diagnostics.Errors())
		}
		var got types.Int64
		if d := resp.Plan.GetAttribute(ctx, path.Root("spec_revision"), &got); d.HasError() {
			t.Fatalf("get spec_revision: %v", d.Errors())
		}
		return got
	}

	t.Run("update with change marks spec_revision unknown", func(t *testing.T) {
		if got := run(build("old.example.com"), build("new.example.com")); !got.IsUnknown() {
			t.Fatalf("expected spec_revision unknown on update, got %v", got)
		}
	})

	t.Run("no-op plan keeps spec_revision known", func(t *testing.T) {
		if got := run(build("same.example.com"), build("same.example.com")); got.IsUnknown() || got.ValueInt64() != 100 {
			t.Fatalf("expected spec_revision to stay 100 on no-op, got %v", got)
		}
	})
}
