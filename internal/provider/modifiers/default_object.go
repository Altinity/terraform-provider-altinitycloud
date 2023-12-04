package modifiers

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ planmodifier.Object = objectDefaultModifier{}

func DefaultObject(elements map[string]attr.Value) objectDefaultModifier {
	return objectDefaultModifier{Elements: elements}
}

type objectDefaultModifier struct {
	Elements map[string]attr.Value
}

func (m objectDefaultModifier) Description(_ context.Context) string {
	return fmt.Sprintf("If value is not configured, defaults to %s", m.Elements)
}

func (m objectDefaultModifier) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("If value is not configured, defaults to `%s`", m.Elements)
}

func (m objectDefaultModifier) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	if !req.ConfigValue.IsNull() {
		return
	}

	if !req.PlanValue.IsUnknown() && !req.PlanValue.IsNull() {
		return
	}
	resp.PlanValue, resp.Diagnostics = types.ObjectValue(req.PlanValue.AttributeTypes(ctx), m.Elements)
}
