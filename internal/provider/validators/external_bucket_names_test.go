package validators

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var externalBucketAttrTypes = map[string]attr.Type{
	"name":        types.StringType,
	"kms_key_arn": types.StringType,
}

func TestUniqueExternalBucketNames(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		value        types.Set
		expectErr    bool
		errSubstring string
	}{
		"unique names": {
			value: externalBucketsSet(
				externalBucketObject("bucket-a", "arn:aws:kms:us-east-1:123456789012:key/one"),
				externalBucketObject("bucket-b", "arn:aws:kms:us-east-1:123456789012:key/two"),
			),
		},
		"duplicate names with different kms keys": {
			value: externalBucketsSet(
				externalBucketObject("bucket-a", "arn:aws:kms:us-east-1:123456789012:key/one"),
				externalBucketObject("bucket-a", "arn:aws:kms:us-east-1:123456789012:key/two"),
			),
			expectErr:    true,
			errSubstring: `bucket "bucket-a"`,
		},
		"duplicate names with null kms key": {
			value: externalBucketsSet(
				externalBucketObject("bucket-a", ""),
				types.ObjectValueMust(externalBucketAttrTypes, map[string]attr.Value{
					"name":        types.StringValue("bucket-a"),
					"kms_key_arn": types.StringNull(),
				}),
			),
			expectErr:    true,
			errSubstring: `bucket "bucket-a"`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req := validator.SetRequest{
				Path:        path.Root("external_buckets"),
				ConfigValue: tc.value,
			}
			resp := &validator.SetResponse{}

			UniqueExternalBucketNames().ValidateSet(context.Background(), req, resp)

			if tc.expectErr && !resp.Diagnostics.HasError() {
				t.Fatal("expected error, got none")
			}
			if !tc.expectErr && resp.Diagnostics.HasError() {
				t.Fatalf("unexpected error: %s", resp.Diagnostics.Errors())
			}
			if tc.errSubstring != "" && !strings.Contains(resp.Diagnostics.Errors()[0].Detail(), tc.errSubstring) {
				t.Fatalf("expected error to contain %q, got %q", tc.errSubstring, resp.Diagnostics.Errors()[0].Detail())
			}
		})
	}
}

func TestUniqueExternalBucketNames_NullAndUnknown(t *testing.T) {
	t.Parallel()

	for name, val := range map[string]types.Set{
		"null":    types.SetNull(types.ObjectType{AttrTypes: externalBucketAttrTypes}),
		"unknown": types.SetUnknown(types.ObjectType{AttrTypes: externalBucketAttrTypes}),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req := validator.SetRequest{
				Path:        path.Root("external_buckets"),
				ConfigValue: val,
			}
			resp := &validator.SetResponse{}

			UniqueExternalBucketNames().ValidateSet(context.Background(), req, resp)

			if resp.Diagnostics.HasError() {
				t.Fatalf("unexpected error: %s", resp.Diagnostics.Errors())
			}
		})
	}
}

func externalBucketsSet(elements ...attr.Value) types.Set {
	return types.SetValueMust(types.ObjectType{AttrTypes: externalBucketAttrTypes}, elements)
}

func externalBucketObject(name, kmsKeyARN string) types.Object {
	return types.ObjectValueMust(externalBucketAttrTypes, map[string]attr.Value{
		"name":        types.StringValue(name),
		"kms_key_arn": types.StringValue(kmsKeyARN),
	})
}
