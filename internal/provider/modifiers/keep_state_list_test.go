package modifiers

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestKeepStateList(t *testing.T) {
	t.Parallel()

	elemType := types.StringType
	state := types.ListValueMust(elemType, []attr.Value{types.StringValue("gp3")})
	config := types.ListValueMust(elemType, []attr.Value{types.StringValue("standard")})

	tests := map[string]struct {
		configValue types.List
		stateValue  types.List
		planValue   types.List
		expected    types.List
	}{
		"config set keeps plan": {
			configValue: config,
			stateValue:  state,
			planValue:   config,
			expected:    config,
		},
		"config null on update copies state": {
			configValue: types.ListNull(elemType),
			stateValue:  state,
			planValue:   types.ListUnknown(elemType),
			expected:    state,
		},
		"config null on create plans null": {
			configValue: types.ListNull(elemType),
			stateValue:  types.ListNull(elemType),
			planValue:   types.ListUnknown(elemType),
			expected:    types.ListNull(elemType),
		},
		"config unknown stays unknown": {
			configValue: types.ListUnknown(elemType),
			stateValue:  state,
			planValue:   types.ListUnknown(elemType),
			expected:    types.ListUnknown(elemType),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req := planmodifier.ListRequest{
				ConfigValue: tc.configValue,
				StateValue:  tc.stateValue,
				PlanValue:   tc.planValue,
			}
			resp := &planmodifier.ListResponse{PlanValue: tc.planValue}

			KeepStateList().PlanModifyList(context.Background(), req, resp)

			if !resp.PlanValue.Equal(tc.expected) {
				t.Errorf("expected %s, got %s", tc.expected, resp.PlanValue)
			}
		})
	}
}
