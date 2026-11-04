package modifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

var _ planmodifier.List = keepStateListModifier{}

// For API fields where omitting the input keeps the current value: plan stays null on create (no unknown to decode).
func KeepStateList() keepStateListModifier {
	return keepStateListModifier{}
}

type keepStateListModifier struct{}

func (m keepStateListModifier) Description(_ context.Context) string {
	return "If value is not configured, keeps the current value."
}

func (m keepStateListModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m keepStateListModifier) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	if !req.ConfigValue.IsNull() {
		return
	}

	resp.PlanValue = req.StateValue
}
