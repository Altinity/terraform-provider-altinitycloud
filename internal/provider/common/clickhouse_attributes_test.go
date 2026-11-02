package common

import (
	"context"
	"testing"

	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// A cluster added to an existing env has state for the resource but not for the entry: computed values must stay unknown.
func TestClickHouseComputedStayUnknownForNewEntries(t *testing.T) {
	ctx := context.Background()
	state := tfsdk.State{Raw: tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{})}

	var walk func(name string, attrs map[string]rschema.Attribute)
	walk = func(name string, attrs map[string]rschema.Attribute) {
		for n, a := range attrs {
			p := name + "." + n
			switch a := a.(type) {
			case rschema.StringAttribute:
				for _, m := range a.PlanModifiers {
					req := planmodifier.StringRequest{State: state, StateValue: types.StringNull(), PlanValue: types.StringUnknown(), ConfigValue: types.StringNull()}
					resp := &planmodifier.StringResponse{PlanValue: req.PlanValue}
					m.PlanModifyString(ctx, req, resp)
					if !resp.PlanValue.IsUnknown() {
						t.Errorf("%s: planned %s, want unknown", p, resp.PlanValue)
					}
				}
			case rschema.Int64Attribute:
				for _, m := range a.PlanModifiers {
					req := planmodifier.Int64Request{State: state, StateValue: types.Int64Null(), PlanValue: types.Int64Unknown(), ConfigValue: types.Int64Null()}
					resp := &planmodifier.Int64Response{PlanValue: req.PlanValue}
					m.PlanModifyInt64(ctx, req, resp)
					if !resp.PlanValue.IsUnknown() {
						t.Errorf("%s: planned %s, want unknown", p, resp.PlanValue)
					}
				}
			case rschema.BoolAttribute:
				for _, m := range a.PlanModifiers {
					req := planmodifier.BoolRequest{State: state, StateValue: types.BoolNull(), PlanValue: types.BoolUnknown(), ConfigValue: types.BoolNull()}
					resp := &planmodifier.BoolResponse{PlanValue: req.PlanValue}
					m.PlanModifyBool(ctx, req, resp)
					if !resp.PlanValue.IsUnknown() {
						t.Errorf("%s: planned %s, want unknown", p, resp.PlanValue)
					}
				}
			case rschema.ListAttribute:
				for _, m := range a.PlanModifiers {
					req := planmodifier.ListRequest{State: state, StateValue: types.ListNull(types.StringType), PlanValue: types.ListUnknown(types.StringType), ConfigValue: types.ListNull(types.StringType)}
					resp := &planmodifier.ListResponse{PlanValue: req.PlanValue}
					m.PlanModifyList(ctx, req, resp)
					if !resp.PlanValue.IsUnknown() {
						t.Errorf("%s: planned %s, want unknown", p, resp.PlanValue)
					}
				}
			case rschema.SingleNestedAttribute:
				walk(p, a.Attributes)
			case rschema.MapNestedAttribute:
				walk(p, a.NestedObject.Attributes)
			case rschema.ListNestedAttribute:
				walk(p, a.NestedObject.Attributes)
			}
		}
	}

	walk("clickhouse_clusters", GetClickHouseClustersAttribute(false, true, false).NestedObject.Attributes)
	walk("clickhouse_keepers", GetClickHouseKeepersAttribute(false, true, false).NestedObject.Attributes)
}
