package modifiers

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestRequiredIfSiblingValue(t *testing.T) {
	t.Parallel()

	testSchema := rschema.Schema{
		Attributes: map[string]rschema.Attribute{
			"type":       rschema.StringAttribute{},
			"bucket_arn": rschema.StringAttribute{},
		},
	}

	objType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"type":       tftypes.String,
			"bucket_arn": tftypes.String,
		},
	}

	makeConfig := func(typeVal, bucketVal tftypes.Value) tfsdk.Config {
		return tfsdk.Config{
			Raw: tftypes.NewValue(objType, map[string]tftypes.Value{
				"type":       typeVal,
				"bucket_arn": bucketVal,
			}),
			Schema: testSchema,
		}
	}

	tests := map[string]struct {
		req       validator.StringRequest
		expectErr bool
	}{
		"config value null (skip)": {
			req: validator.StringRequest{
				Path:        path.Root("bucket_arn"),
				ConfigValue: types.StringNull(),
				Config: makeConfig(
					tftypes.NewValue(tftypes.String, "S3_TABLE"),
					tftypes.NewValue(tftypes.String, nil),
				),
			},
			expectErr: false,
		},
		"config value unknown (skip)": {
			req: validator.StringRequest{
				Path:        path.Root("bucket_arn"),
				ConfigValue: types.StringUnknown(),
				Config: makeConfig(
					tftypes.NewValue(tftypes.String, "S3_TABLE"),
					tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				),
			},
			expectErr: false,
		},
		"sibling matches expected (pass)": {
			req: validator.StringRequest{
				Path:        path.Root("bucket_arn"),
				ConfigValue: types.StringValue("arn:aws:s3:::bucket"),
				Config: makeConfig(
					tftypes.NewValue(tftypes.String, "S3_TABLE"),
					tftypes.NewValue(tftypes.String, "arn:aws:s3:::bucket"),
				),
			},
			expectErr: false,
		},
		"sibling does not match (error)": {
			req: validator.StringRequest{
				Path:        path.Root("bucket_arn"),
				ConfigValue: types.StringValue("arn:aws:s3:::bucket"),
				Config: makeConfig(
					tftypes.NewValue(tftypes.String, "S3"),
					tftypes.NewValue(tftypes.String, "arn:aws:s3:::bucket"),
				),
			},
			expectErr: true,
		},
		"sibling null (skip)": {
			req: validator.StringRequest{
				Path:        path.Root("bucket_arn"),
				ConfigValue: types.StringValue("arn:aws:s3:::bucket"),
				Config: makeConfig(
					tftypes.NewValue(tftypes.String, nil),
					tftypes.NewValue(tftypes.String, "arn:aws:s3:::bucket"),
				),
			},
			expectErr: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			resp := &validator.StringResponse{}
			RequiredIfSiblingValue("type", "S3_TABLE").ValidateString(context.Background(), tc.req, resp)

			if tc.expectErr && !resp.Diagnostics.HasError() {
				t.Error("expected error diagnostic, got none")
			}
			if !tc.expectErr && resp.Diagnostics.HasError() {
				t.Errorf("unexpected error: %s", resp.Diagnostics.Errors())
			}
		})
	}
}
