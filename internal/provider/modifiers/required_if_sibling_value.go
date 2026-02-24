package modifiers

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.String = requiredIfSiblingValue{}

// RequiredIfSiblingValue returns a validator that checks a string attribute is only
// set when a sibling attribute has a specific value.
// For example: custom_s3_table_bucket_arn can only be set when type == "S3_TABLE".
func RequiredIfSiblingValue(siblingAttr, expectedValue string) requiredIfSiblingValue {
	return requiredIfSiblingValue{
		SiblingAttr:   siblingAttr,
		ExpectedValue: expectedValue,
	}
}

type requiredIfSiblingValue struct {
	SiblingAttr   string
	ExpectedValue string
}

func (v requiredIfSiblingValue) Description(_ context.Context) string {
	return fmt.Sprintf("Attribute can only be set when '%s' is '%s'.", v.SiblingAttr, v.ExpectedValue)
}

func (v requiredIfSiblingValue) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("Attribute can only be set when `%s` is `%s`.", v.SiblingAttr, v.ExpectedValue)
}

func (v requiredIfSiblingValue) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// Skip if attribute is not configured.
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	// Get the sibling attribute value.
	var siblingValue types.String
	diags := req.Config.GetAttribute(ctx, req.Path.ParentPath().AtName(v.SiblingAttr), &siblingValue)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if siblingValue.IsNull() || siblingValue.IsUnknown() {
		return
	}

	if siblingValue.ValueString() != v.ExpectedValue {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Attribute Combination",
			fmt.Sprintf("Attribute can only be set when '%s' is '%s', got '%s'.",
				v.SiblingAttr,
				v.ExpectedValue,
				siblingValue.ValueString(),
			),
		)
	}
}
