package modifiers

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

var _ planmodifier.String = immutableStringModifier{}

func ImmutableString() immutableStringModifier {
	return immutableStringModifier{}
}

type immutableStringModifier struct{}

func (m immutableStringModifier) Description(_ context.Context) string {
	return "Attribute is immutable after creation."
}

func (m immutableStringModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m immutableStringModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Only check when the attribute is being modified (i.e. not being created or destroyed).
	if req.StateValue.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	if req.PlanValue.IsUnknown() {
		resp.PlanValue = req.StateValue
		return
	}

	if req.StateValue.ValueString() != req.PlanValue.ValueString() {
		resp.Diagnostics.AddAttributeError(req.Path, "Immutable Attribute", fmt.Sprintf("%s is immutable and cannot be modified after creation.", req.Path))
	}
}
