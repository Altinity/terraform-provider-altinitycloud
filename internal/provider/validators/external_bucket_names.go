package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type uniqueExternalBucketNamesValidator struct{}

func UniqueExternalBucketNames() validator.Set {
	return uniqueExternalBucketNamesValidator{}
}

func (v uniqueExternalBucketNamesValidator) Description(_ context.Context) string {
	return "external bucket names must be unique"
}

func (v uniqueExternalBucketNamesValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v uniqueExternalBucketNamesValidator) ValidateSet(_ context.Context, req validator.SetRequest, resp *validator.SetResponse) {
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

		nameAttr, ok := obj.Attributes()["name"]
		if !ok || nameAttr.IsNull() || nameAttr.IsUnknown() {
			continue
		}

		name, ok := nameAttr.(types.String)
		if !ok {
			continue
		}

		bucketName := name.ValueString()
		if _, exists := seen[bucketName]; exists {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Duplicate External Bucket Name",
				fmt.Sprintf("external_buckets contains more than one entry for bucket %q. Bucket names must be unique.", bucketName),
			)
			return
		}

		seen[bucketName] = struct{}{}
	}
}
