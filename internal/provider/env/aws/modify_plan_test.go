package env

import (
	"context"
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/testutil"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestAWSModifyPlanSpecRevision(t *testing.T) {
	testutil.AssertModifyPlanSpecRevision(t, &AWSEnvResource{})
}

func TestAWSModifyPlanClickHouseRemoval(t *testing.T) {
	cfg := awsSchema(t)

	entries := func(names ...string) func(tftypes.Type) tftypes.Value {
		return func(at tftypes.Type) tftypes.Value {
			mapType, ok := at.(tftypes.Map)
			if !ok {
				t.Fatalf("expected a map type, got %T", at)
			}
			objType, ok := mapType.ElementType.(tftypes.Object)
			if !ok {
				t.Fatalf("expected an object element type, got %T", mapType.ElementType)
			}
			elems := map[string]tftypes.Value{}
			for _, n := range names {
				elems[n] = buildObject(objType)
			}
			return tftypes.NewValue(mapType, elems)
		}
	}
	env := func(force bool, clusters, keepers func(tftypes.Type) tftypes.Value) tftypes.Value {
		return buildConfig(t, cfg, map[string]func(tftypes.Type) tftypes.Value{
			"name":                   func(at tftypes.Type) tftypes.Value { return tftypes.NewValue(at, "env") },
			"force_destroy_clusters": func(at tftypes.Type) tftypes.Value { return tftypes.NewValue(at, force) },
			"clickhouse_clusters":    clusters,
			"clickhouse_keepers":     keepers,
		})
	}
	run := func(state, plan tftypes.Value) *resource.ModifyPlanResponse {
		resp := &resource.ModifyPlanResponse{Plan: tfsdk.Plan{Raw: plan, Schema: cfg.Schema}}
		(&AWSEnvResource{}).ModifyPlan(context.Background(), resource.ModifyPlanRequest{
			Config: tfsdk.Config{Raw: plan, Schema: cfg.Schema},
			State:  tfsdk.State{Raw: state, Schema: cfg.Schema},
			Plan:   tfsdk.Plan{Raw: plan, Schema: cfg.Schema},
		}, resp)
		return resp
	}
	state := env(false, entries("ch", "other"), entries("keeper"))

	t.Run("removing a cluster or keeper without the flag fails", func(t *testing.T) {
		resp := run(state, env(false, entries("other"), entries()))
		if got := len(resp.Diagnostics.Errors()); got != 2 {
			t.Fatalf("expected 2 errors, got %v", resp.Diagnostics.Errors())
		}
	})

	t.Run("force_destroy_clusters set in the same plan allows it", func(t *testing.T) {
		if resp := run(state, env(true, entries("other"), entries())); resp.Diagnostics.HasError() {
			t.Fatalf("errored: %v", resp.Diagnostics.Errors())
		}
	})

	t.Run("keeping every entry passes without the flag", func(t *testing.T) {
		if resp := run(state, env(false, entries("ch", "other"), entries("keeper"))); resp.Diagnostics.HasError() {
			t.Fatalf("errored: %v", resp.Diagnostics.Errors())
		}
	})
}

func buildObject(objType tftypes.Object) tftypes.Value {
	vals := map[string]tftypes.Value{}
	for name, at := range objType.AttributeTypes {
		vals[name] = tftypes.NewValue(at, nil)
	}
	return tftypes.NewValue(objType, vals)
}
