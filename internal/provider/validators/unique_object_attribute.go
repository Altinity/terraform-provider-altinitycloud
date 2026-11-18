package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type uniqueObjectAttributeValidator struct {
	attribute string
	summary   string
	detail    func(value string) string
}

func UniqueObjectAttribute(attribute, summary string, detail func(value string) string) validator.Set {
	return uniqueObjectAttributeValidator{attribute: attribute, summary: summary, detail: detail}
}

func UniqueListObjectAttribute(attribute, summary string, detail func(value string) string) validator.List {
	return uniqueObjectAttributeValidator{attribute: attribute, summary: summary, detail: detail}
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
	v.validate(req.ConfigValue.Elements(), req.Path, &resp.Diagnostics)
}

func (v uniqueObjectAttributeValidator) ValidateList(_ context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	v.validate(req.ConfigValue.Elements(), req.Path, &resp.Diagnostics)
}

func (v uniqueObjectAttributeValidator) validate(elements []attr.Value, p path.Path, diags *diag.Diagnostics) {
	seen := make(map[string]struct{}, len(elements))

	for _, elem := range elements {
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
			diags.AddAttributeError(p, v.summary, v.detail(value))
			return
		}

		seen[value] = struct{}{}
	}
}
