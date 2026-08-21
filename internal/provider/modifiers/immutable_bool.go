package modifiers

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

var _ planmodifier.Bool = immutableBoolModifier{}

func ImmutableBool() immutableBoolModifier {
	return immutableBoolModifier{}
}

type immutableBoolModifier struct{}

func (m immutableBoolModifier) Description(_ context.Context) string {
	return "Attribute is immutable after creation."
}

func (m immutableBoolModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m immutableBoolModifier) PlanModifyBool(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	if req.StateValue.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	if req.PlanValue.IsUnknown() {
		resp.PlanValue = req.StateValue
		return
	}

	if req.StateValue.ValueBool() != req.PlanValue.ValueBool() {
		resp.Diagnostics.AddAttributeError(req.Path, "Immutable Attribute", fmt.Sprintf("%s is immutable and cannot be modified after creation.", req.Path))
	}
}
