package modifiers

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDefaultObject(t *testing.T) {
	t.Parallel()

	attrTypes := map[string]attr.Type{
		"enabled": types.BoolType,
	}
	defaults := map[string]attr.Value{
		"enabled": types.BoolValue(true),
	}

	tests := map[string]struct {
		configValue types.Object
		planValue   types.Object
		expectSet   bool
	}{
		"config not null (skip)": {
			configValue: types.ObjectValueMust(attrTypes, map[string]attr.Value{"enabled": types.BoolValue(false)}),
			planValue:   types.ObjectUnknown(attrTypes),
			expectSet:   false,
		},
		"plan already set (skip)": {
			configValue: types.ObjectNull(attrTypes),
			planValue:   types.ObjectValueMust(attrTypes, map[string]attr.Value{"enabled": types.BoolValue(false)}),
			expectSet:   false,
		},
		"config null and plan unknown (set default)": {
			configValue: types.ObjectNull(attrTypes),
			planValue:   types.ObjectUnknown(attrTypes),
			expectSet:   true,
		},
		"config null and plan null (set default)": {
			configValue: types.ObjectNull(attrTypes),
			planValue:   types.ObjectNull(attrTypes),
			expectSet:   true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req := planmodifier.ObjectRequest{
				ConfigValue: tc.configValue,
				PlanValue:   tc.planValue,
			}
			resp := &planmodifier.ObjectResponse{
				PlanValue: tc.planValue,
			}

			DefaultObject(defaults).PlanModifyObject(context.Background(), req, resp)

			if resp.Diagnostics.HasError() {
				t.Fatalf("unexpected error: %s", resp.Diagnostics.Errors())
			}

			if tc.expectSet {
				expected := types.ObjectValueMust(attrTypes, defaults)
				if !resp.PlanValue.Equal(expected) {
					t.Errorf("expected PlanValue %s, got %s", expected, resp.PlanValue)
				}
			} else {
				if !resp.PlanValue.Equal(tc.planValue) {
					t.Errorf("expected PlanValue to remain %s, got %s", tc.planValue, resp.PlanValue)
				}
			}
		})
	}
}
