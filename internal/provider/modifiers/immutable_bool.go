package modifiers

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

var _ planmodifier.Bool = immutableBoolModifier{}

func ImmutableBool(attributeName string) immutableBoolModifier {
	return immutableBoolModifier{AttributeName: attributeName}
}

type immutableBoolModifier struct {
	AttributeName string
}

func (m immutableBoolModifier) Description(_ context.Context) string {
	return fmt.Sprintf("Attribute '%s' is immutable after creation.", m.AttributeName)
}

func (m immutableBoolModifier) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("Attribute `%s` is immutable after creation.", m.AttributeName)
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
		resp.Diagnostics.AddAttributeError(path.Root(m.AttributeName), "Immutable Attribute", fmt.Sprintf("%s is immutable and cannot be modified after creation.", m.AttributeName))
		return
	}
}
