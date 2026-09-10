package env

import (
	"context"
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestK8SEnvModel_toSDK_MFA(t *testing.T) {
	tests := []struct {
		name string
		mfa  types.Bool
		want *bool
	}{
		{name: "null leaves the input unset so the API applies its own default", mfa: types.BoolNull()},
		{name: "true is sent", mfa: types.BoolValue(true), want: boolPtr(true)},
		{name: "false is sent rather than dropped", mfa: types.BoolValue(false), want: boolPtr(false)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			create, update, diags := K8SEnvModel{MFA: tt.mfa}.toSDK(context.Background())
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}

			assertMFAInput(t, "create", create.Spec.Mfa, tt.want)
			assertMFAInput(t, "update", update.Spec.Mfa, tt.want)
		})
	}
}

func TestK8SEnvModel_toModel_MFA(t *testing.T) {
	for _, want := range []bool{true, false} {
		var model K8SEnvModel
		if diags := model.toModel("dummy", 1, client.K8SEnvSpecFragment{Mfa: want}); diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}

		if model.MFA.IsNull() || model.MFA.ValueBool() != want {
			t.Errorf("mfa = %v, want %v", model.MFA, want)
		}
	}
}

func assertMFAInput(t *testing.T, phase string, got, want *bool) {
	t.Helper()

	switch {
	case want == nil && got != nil:
		t.Errorf("%s mfa = %v, want unset", phase, *got)
	case want != nil && got == nil:
		t.Errorf("%s mfa unset, want %v", phase, *want)
	case want != nil && *got != *want:
		t.Errorf("%s mfa = %v, want %v", phase, *got, *want)
	}
}

func boolPtr(b bool) *bool { return &b }
