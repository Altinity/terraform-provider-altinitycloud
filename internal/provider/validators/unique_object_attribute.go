package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type uniqueObjectAttributeValidator struct {
	attribute string
	summary   string
	detail    func(value string) string
}

func UniqueObjectAttribute(attribute, summary string, detail func(value string) string) validator.Set {
	return uniqueObjectAttributeValidator{
		attribute: attribute,
		summary:   summary,
		detail:    detail,
	}
}

func (v uniqueObjectAttributeValidator) Description(_ context.Context) string {
	return fmt.Sprintf("%s values must be unique", v.attribute)
}

func (v uniqueObjectAttributeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v uniqueObjectAttributeValidator) ValidateSet(_ context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	seen := make(map[string]struct{}, len(req.ConfigValue.Elements()))

	for _, elem := range req.ConfigValue.Elements() {
		if elem.IsNull() || elem.IsUnknown() {
			continue
		}

		obj, ok := elem.(types.Object)
		if !ok {
			continue
		}

		attribute, ok := obj.Attributes()[v.attribute]
		if !ok || attribute.IsNull() || attribute.IsUnknown() {
			continue
		}

		str, ok := attribute.(types.String)
		if !ok {
			continue
		}

		value := str.ValueString()
		if _, exists := seen[value]; exists {
			resp.Diagnostics.AddAttributeError(req.Path, v.summary, v.detail(value))
			return
		}

		seen[value] = struct{}{}
	}
}
