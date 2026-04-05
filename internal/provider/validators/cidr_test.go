package validators

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestCIDRValidator(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		value     string
		expectErr bool
	}{
		"valid /16": {
			value: "10.0.0.0/16", expectErr: false,
		},
		"valid /21": {
			value: "10.0.0.0/21", expectErr: false,
		},
		"valid /32": {
			value: "192.168.1.1/32", expectErr: false,
		},
		"invalid octets": {
			value: "999.999.999.999/16", expectErr: true,
		},
		"invalid prefix": {
			value: "10.0.0.0/99", expectErr: true,
		},
		"not CIDR": {
			value: "not-a-cidr", expectErr: true,
		},
		"empty": {
			value: "", expectErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			req := validator.StringRequest{
				Path:        path.Root("test"),
				ConfigValue: types.StringValue(tc.value),
			}
			resp := &validator.StringResponse{}
			CIDR().ValidateString(context.Background(), req, resp)
			if tc.expectErr && !resp.Diagnostics.HasError() {
				t.Error("expected error, got none")
			}
			if !tc.expectErr && resp.Diagnostics.HasError() {
				t.Errorf("unexpected error: %s", resp.Diagnostics.Errors())
			}
		})
	}
}

func TestCIDRWithMaxPrefixValidator(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		value     string
		maxPrefix int
		expectErr bool
	}{
		"/16 with max /21 (ok)": {
			value: "10.0.0.0/16", maxPrefix: 21, expectErr: false,
		},
		"/21 with max /21 (ok)": {
			value: "10.0.0.0/21", maxPrefix: 21, expectErr: false,
		},
		"/24 with max /21 (rejected)": {
			value: "10.0.0.0/24", maxPrefix: 21, expectErr: true,
		},
		"/28 with max /21 (rejected)": {
			value: "10.0.0.0/28", maxPrefix: 21, expectErr: true,
		},
		"invalid CIDR with max prefix": {
			value: "999.0.0.0/16", maxPrefix: 21, expectErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			req := validator.StringRequest{
				Path:        path.Root("test"),
				ConfigValue: types.StringValue(tc.value),
			}
			resp := &validator.StringResponse{}
			CIDRWithMaxPrefix(tc.maxPrefix).ValidateString(context.Background(), req, resp)
			if tc.expectErr && !resp.Diagnostics.HasError() {
				t.Error("expected error, got none")
			}
			if !tc.expectErr && resp.Diagnostics.HasError() {
				t.Errorf("unexpected error: %s", resp.Diagnostics.Errors())
			}
		})
	}
}

func TestCIDRValidator_NullAndUnknown(t *testing.T) {
	t.Parallel()

	for name, val := range map[string]types.String{
		"null":    types.StringNull(),
		"unknown": types.StringUnknown(),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			req := validator.StringRequest{
				Path:        path.Root("test"),
				ConfigValue: val,
			}
			resp := &validator.StringResponse{}
			CIDRWithMaxPrefix(21).ValidateString(context.Background(), req, resp)
			if resp.Diagnostics.HasError() {
				t.Errorf("unexpected error for %s value: %s", name, resp.Diagnostics.Errors())
			}
		})
	}
}
