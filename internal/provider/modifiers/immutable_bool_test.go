package modifiers

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestImmutableBool_PlanModifyBool(t *testing.T) {
	t.Parallel()

	nonNullPlan := tfsdk.Plan{
		Raw: tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{}}, map[string]tftypes.Value{}),
	}
	nullPlan := tfsdk.Plan{
		Raw: tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{}}, nil),
	}

	tests := map[string]struct {
		req             planmodifier.BoolRequest
		expectErr       bool
		expectPlanValue *types.Bool
	}{
		"create (state null) allows any value": {
			req: planmodifier.BoolRequest{
				StateValue:  types.BoolNull(),
				PlanValue:   types.BoolValue(true),
				ConfigValue: types.BoolValue(true),
				Plan:        nonNullPlan,
			},
			expectErr: false,
		},
		"destroy (plan null) allows deletion": {
			req: planmodifier.BoolRequest{
				StateValue:  types.BoolValue(true),
				PlanValue:   types.BoolNull(),
				ConfigValue: types.BoolNull(),
				Plan:        nullPlan,
			},
			expectErr: false,
		},
		"no change passes": {
			req: planmodifier.BoolRequest{
				StateValue:  types.BoolValue(true),
				PlanValue:   types.BoolValue(true),
				ConfigValue: types.BoolValue(true),
				Plan:        nonNullPlan,
			},
			expectErr: false,
		},
		"value changed errors": {
			req: planmodifier.BoolRequest{
				StateValue:  types.BoolValue(true),
				PlanValue:   types.BoolValue(false),
				ConfigValue: types.BoolValue(false),
				Plan:        nonNullPlan,
			},
			expectErr: true,
		},
		"config null with value change errors": {
			req: planmodifier.BoolRequest{
				StateValue:  types.BoolValue(true),
				PlanValue:   types.BoolValue(false),
				ConfigValue: types.BoolNull(),
				Plan:        nonNullPlan,
			},
			expectErr: true,
		},
		"plan unknown preserves state": {
			req: planmodifier.BoolRequest{
				StateValue:  types.BoolValue(true),
				PlanValue:   types.BoolUnknown(),
				ConfigValue: types.BoolNull(),
				Plan:        nonNullPlan,
			},
			expectErr:       false,
			expectPlanValue: boolPtr(types.BoolValue(true)),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			resp := &planmodifier.BoolResponse{PlanValue: tc.req.PlanValue}
			ImmutableBool("internal").PlanModifyBool(context.Background(), tc.req, resp)

			if tc.expectErr && !resp.Diagnostics.HasError() {
				t.Error("expected error diagnostic, got none")
			}
			if !tc.expectErr && resp.Diagnostics.HasError() {
				t.Errorf("unexpected error: %s", resp.Diagnostics.Errors())
			}
			if tc.expectPlanValue != nil && resp.PlanValue != *tc.expectPlanValue {
				t.Errorf("expected PlanValue %s, got %s", *tc.expectPlanValue, resp.PlanValue)
			}
		})
	}
}

func boolPtr(v types.Bool) *types.Bool { return &v }
