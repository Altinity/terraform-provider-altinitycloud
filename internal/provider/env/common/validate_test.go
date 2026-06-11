package env

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestValidateDatadog(t *testing.T) {
	tests := []struct {
		name      string
		datadog   *DatadogModel
		expectErr bool
	}{
		{
			name:      "Nil datadog",
			datadog:   nil,
			expectErr: false,
		},
		{
			name:      "Disabled without key",
			datadog:   &DatadogModel{Enabled: types.BoolValue(false)},
			expectErr: false,
		},
		{
			name:      "Enabled null treated as disabled",
			datadog:   &DatadogModel{Enabled: types.BoolNull()},
			expectErr: false,
		},
		{
			name:      "Enabled with key",
			datadog:   &DatadogModel{Enabled: types.BoolValue(true), EncAPIKey: types.StringValue("enc-secret")},
			expectErr: false,
		},
		{
			name:      "Enabled with unknown key is skipped",
			datadog:   &DatadogModel{Enabled: types.BoolValue(true), EncAPIKey: types.StringUnknown()},
			expectErr: false,
		},
		{
			name:      "Enabled with null key errors",
			datadog:   &DatadogModel{Enabled: types.BoolValue(true), EncAPIKey: types.StringNull()},
			expectErr: true,
		},
		{
			name:      "Enabled with empty key errors",
			datadog:   &DatadogModel{Enabled: types.BoolValue(true), EncAPIKey: types.StringValue("")},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := ValidateDatadog(tt.datadog)
			if diags.HasError() != tt.expectErr {
				t.Errorf("expected error: %v, got diagnostics: %v", tt.expectErr, diags)
			}
		})
	}
}
