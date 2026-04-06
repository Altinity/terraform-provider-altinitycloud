package modifiers

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

var _ planmodifier.String = immutableStringModifier{}

func ImmutableString(attributeName string) immutableStringModifier {
	return immutableStringModifier{AttributeName: attributeName}
}

type immutableStringModifier struct {
	AttributeName string
}

func (m immutableStringModifier) Description(_ context.Context) string {
	return fmt.Sprintf("Attribute '%s' is immutable after creation.", m.AttributeName)
}

func (m immutableStringModifier) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("Attribute `%s` is immutable after creation.", m.AttributeName)
}

func (m immutableStringModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Only check when the attribute is being modified (i.e. not being created or destroyed).
	if req.StateValue.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	// Skip check if the property is not present in the current configuration
	if req.ConfigValue.IsNull() {
		return
	}

	if req.StateValue.ValueString() != req.PlanValue.ValueString() {
		resp.Diagnostics.AddAttributeError(path.Root(m.AttributeName), "Immutable Attribute", fmt.Sprintf("%s is immutable and cannot be modified after creation.", m.AttributeName))
		return
	}
}
