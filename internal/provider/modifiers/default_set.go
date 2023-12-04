package modifiers

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ planmodifier.Set = setDefaultModifier{}

func DefaultSet(elementsType attr.Type, elements []attr.Value) setDefaultModifier {
	return setDefaultModifier{ElementsType: elementsType, Elements: elements}
}

type setDefaultModifier struct {
	Elements     []attr.Value
	ElementsType attr.Type
}

func (m setDefaultModifier) Description(_ context.Context) string {
	return fmt.Sprintf("If value is not configured, defaults to %s", m.Elements)
}

func (m setDefaultModifier) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("If value is not configured, defaults to `%s`", m.Elements)
}

func (m setDefaultModifier) PlanModifySet(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	if !req.ConfigValue.IsNull() {
		return
	}

	if !req.PlanValue.IsUnknown() && !req.PlanValue.IsNull() {
		return
	}

	resp.PlanValue, resp.Diagnostics = types.SetValue(m.ElementsType, m.Elements)
}
