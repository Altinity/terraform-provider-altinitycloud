package modifiers

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
		req             planmodifier.StringRequest
		expectErr       bool
		expectPlanValue *types.String
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
		"plan unknown preserves state": {
			req: planmodifier.StringRequest{
				StateValue:  types.StringValue("us-east-1"),
				PlanValue:   types.StringUnknown(),
				ConfigValue: types.StringNull(),
				Plan:        nonNullPlan,
			},
			expectErr:       false,
			expectPlanValue: ptr(types.StringValue("us-east-1")),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			resp := &planmodifier.StringResponse{PlanValue: tc.req.PlanValue}
			ImmutableString().PlanModifyString(context.Background(), tc.req, resp)

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

func ptr[T any](v T) *T { return &v }

// The modifier used to take the attribute name as a string, so a shared schema
// helper reused for a differently named attribute reported the wrong path.
func TestImmutableString_DiagnosticUsesRequestPath(t *testing.T) {
	t.Parallel()

	req := planmodifier.StringRequest{
		Path:        path.Root("network_zone"),
		StateValue:  types.StringValue("eu-central"),
		PlanValue:   types.StringValue("us-east"),
		ConfigValue: types.StringValue("us-east"),
		Plan: tfsdk.Plan{
			Raw: tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{}}, map[string]tftypes.Value{}),
		},
	}

	resp := &planmodifier.StringResponse{PlanValue: req.PlanValue}
	ImmutableString().PlanModifyString(context.Background(), req, resp)

	errs := resp.Diagnostics.Errors()
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}

	withPath, ok := errs[0].(diag.DiagnosticWithPath)
	if !ok {
		t.Fatal("expected a diagnostic carrying a path")
	}
	if got := withPath.Path().String(); got != "network_zone" {
		t.Errorf("diagnostic path: got %q, want %q", got, "network_zone")
	}
	if !strings.Contains(errs[0].Detail(), "network_zone") {
		t.Errorf("detail should name the attribute, got: %s", errs[0].Detail())
	}
}
