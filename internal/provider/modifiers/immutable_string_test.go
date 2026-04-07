package modifiers

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestImmutableString_PlanModifyString(t *testing.T) {
	t.Parallel()

	nonNullPlan := tfsdk.Plan{
		Raw: tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{}}, map[string]tftypes.Value{}),
	}
	nullPlan := tfsdk.Plan{
		Raw: tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{}}, nil),
	}

	tests := map[string]struct {
		req       planmodifier.StringRequest
		expectErr bool
	}{
		"create (state null) allows any value": {
			req: planmodifier.StringRequest{
				StateValue:  types.StringNull(),
				PlanValue:   types.StringValue("us-east-1"),
				ConfigValue: types.StringValue("us-east-1"),
				Plan:        nonNullPlan,
			},
			expectErr: false,
		},
		"destroy (plan null) allows deletion": {
			req: planmodifier.StringRequest{
				StateValue:  types.StringValue("us-east-1"),
				PlanValue:   types.StringNull(),
				ConfigValue: types.StringNull(),
				Plan:        nullPlan,
			},
			expectErr: false,
		},
		"no change passes": {
			req: planmodifier.StringRequest{
				StateValue:  types.StringValue("us-east-1"),
				PlanValue:   types.StringValue("us-east-1"),
				ConfigValue: types.StringValue("us-east-1"),
				Plan:        nonNullPlan,
			},
			expectErr: false,
		},
		"value changed errors": {
			req: planmodifier.StringRequest{
				StateValue:  types.StringValue("us-east-1"),
				PlanValue:   types.StringValue("us-west-2"),
				ConfigValue: types.StringValue("us-west-2"),
				Plan:        nonNullPlan,
			},
			expectErr: true,
		},
		"config null with value change errors": {
			req: planmodifier.StringRequest{
				StateValue:  types.StringValue("us-east-1"),
				PlanValue:   types.StringValue("different-value"),
				ConfigValue: types.StringNull(),
				Plan:        nonNullPlan,
			},
			expectErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			resp := &planmodifier.StringResponse{}
			ImmutableString("region").PlanModifyString(context.Background(), tc.req, resp)

			if tc.expectErr && !resp.Diagnostics.HasError() {
				t.Error("expected error diagnostic, got none")
			}
			if !tc.expectErr && resp.Diagnostics.HasError() {
				t.Errorf("unexpected error: %s", resp.Diagnostics.Errors())
			}
		})
	}
}
