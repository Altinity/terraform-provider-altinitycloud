package env

import (
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDatadogToSDK(t *testing.T) {
	tests := []struct {
		name     string
		input    *DatadogModel
		expected *client.DatadogSpecInput
	}{
		{
			name:     "Nil input",
			input:    nil,
			expected: nil,
		},
		{
			name: "Full config",
			input: &DatadogModel{
				Enabled:        types.BoolValue(true),
				EncAPIKey:      types.StringValue("enc-secret"),
				Domain:         types.StringValue("us3.datadoghq.com"),
				LogsEnabled:    types.BoolValue(true),
				MetricsEnabled: types.BoolValue(true),
			},
			expected: &client.DatadogSpecInput{
				Enabled:        boolPtr(true),
				EncAPIKey:      strPtr("enc-secret"),
				Domain:         strPtr("us3.datadoghq.com"),
				LogsEnabled:    boolPtr(true),
				MetricsEnabled: boolPtr(true),
			},
		},
		{
			name: "Enc API key unset is omitted",
			input: &DatadogModel{
				Enabled:        types.BoolValue(true),
				EncAPIKey:      types.StringNull(),
				Domain:         types.StringValue("datadoghq.com"),
				LogsEnabled:    types.BoolValue(false),
				MetricsEnabled: types.BoolValue(false),
			},
			expected: &client.DatadogSpecInput{
				Enabled:        boolPtr(true),
				EncAPIKey:      nil,
				Domain:         strPtr("datadoghq.com"),
				LogsEnabled:    boolPtr(false),
				MetricsEnabled: boolPtr(false),
			},
		},
		{
			name: "Disabled",
			input: &DatadogModel{
				Enabled:        types.BoolValue(false),
				EncAPIKey:      types.StringNull(),
				Domain:         types.StringValue("datadoghq.com"),
				LogsEnabled:    types.BoolValue(false),
				MetricsEnabled: types.BoolValue(false),
			},
			expected: &client.DatadogSpecInput{
				Enabled:        boolPtr(false),
				EncAPIKey:      nil,
				Domain:         strPtr("datadoghq.com"),
				LogsEnabled:    boolPtr(false),
				MetricsEnabled: boolPtr(false),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DatadogToSDK(tt.input)

			if (tt.expected == nil) != (result == nil) {
				t.Fatalf("Expected nil: %v, got nil: %v", tt.expected == nil, result == nil)
			}
			if tt.expected == nil {
				return
			}

			assertBoolPtr(t, "Enabled", tt.expected.Enabled, result.Enabled)
			assertStrPtr(t, "EncAPIKey", tt.expected.EncAPIKey, result.EncAPIKey)
			assertStrPtr(t, "Domain", tt.expected.Domain, result.Domain)
			assertBoolPtr(t, "LogsEnabled", tt.expected.LogsEnabled, result.LogsEnabled)
			assertBoolPtr(t, "MetricsEnabled", tt.expected.MetricsEnabled, result.MetricsEnabled)
		})
	}
}

func boolPtr(b bool) *bool    { return &b }
func strPtr(s string) *string { return &s }

func assertBoolPtr(t *testing.T, field string, expected, got *bool) {
	t.Helper()
	if (expected == nil) != (got == nil) {
		t.Errorf("%s nil mismatch: expected nil %v, got nil %v", field, expected == nil, got == nil)
		return
	}
	if expected != nil && *expected != *got {
		t.Errorf("%s mismatch: expected %v, got %v", field, *expected, *got)
	}
}

func assertStrPtr(t *testing.T, field string, expected, got *string) {
	t.Helper()
	if (expected == nil) != (got == nil) {
		t.Errorf("%s nil mismatch: expected nil %v, got nil %v", field, expected == nil, got == nil)
		return
	}
	if expected != nil && *expected != *got {
		t.Errorf("%s mismatch: expected %q, got %q", field, *expected, *got)
	}
}
