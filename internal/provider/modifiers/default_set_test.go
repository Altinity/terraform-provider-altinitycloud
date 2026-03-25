package modifiers

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDefaultSet(t *testing.T) {
	t.Parallel()

	elemType := types.StringType
	defaults := []attr.Value{types.StringValue("default")}

	tests := map[string]struct {
		configValue types.Set
		planValue   types.Set
		expectSet   bool
	}{
		"config not null (skip)": {
			configValue: types.SetValueMust(elemType, []attr.Value{types.StringValue("custom")}),
			planValue:   types.SetUnknown(elemType),
			expectSet:   false,
		},
		"plan already set (skip)": {
			configValue: types.SetNull(elemType),
			planValue:   types.SetValueMust(elemType, []attr.Value{types.StringValue("existing")}),
			expectSet:   false,
		},
		"config null and plan unknown (set default)": {
			configValue: types.SetNull(elemType),
			planValue:   types.SetUnknown(elemType),
			expectSet:   true,
		},
		"config null and plan null (set default)": {
			configValue: types.SetNull(elemType),
			planValue:   types.SetNull(elemType),
			expectSet:   true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req := planmodifier.SetRequest{
				ConfigValue: tc.configValue,
				PlanValue:   tc.planValue,
			}
			resp := &planmodifier.SetResponse{
				PlanValue: tc.planValue,
			}

			DefaultSet(elemType, defaults).PlanModifySet(context.Background(), req, resp)

			if resp.Diagnostics.HasError() {
				t.Fatalf("unexpected error: %s", resp.Diagnostics.Errors())
			}

			if tc.expectSet {
				expected := types.SetValueMust(elemType, defaults)
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
