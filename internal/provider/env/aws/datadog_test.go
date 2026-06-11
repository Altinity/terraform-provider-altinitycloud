package env

import (
	"testing"

	common "github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/common"
	sdk "github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDatadogToModel(t *testing.T) {
	tests := []struct {
		name     string
		existing *common.DatadogModel
		input    *sdk.AWSEnvSpecFragment_Datadog
		expected *common.DatadogModel
	}{
		{
			name:     "Nil fragment returns existing",
			existing: nil,
			input:    nil,
			expected: nil,
		},
		{
			// Existing envs upgrading without a datadog block must not drift:
			// the API always returns the block (disabled by default) but state stays null.
			name:     "Unconfigured and disabled stays nil",
			existing: nil,
			input:    &sdk.AWSEnvSpecFragment_Datadog{Enabled: false, Domain: "datadoghq.com"},
			expected: nil,
		},
		{
			name:     "Unconfigured but enabled out-of-band populates without api key",
			existing: nil,
			input:    &sdk.AWSEnvSpecFragment_Datadog{Enabled: true, Domain: "us3.datadoghq.com", LogsEnabled: true, MetricsEnabled: false},
			expected: &common.DatadogModel{
				Enabled:        types.BoolValue(true),
				EncAPIKey:      types.StringNull(),
				Domain:         types.StringValue("us3.datadoghq.com"),
				LogsEnabled:    types.BoolValue(true),
				MetricsEnabled: types.BoolValue(false),
			},
		},
		{
			name:     "Configured preserves write-only enc_api_key",
			existing: &common.DatadogModel{EncAPIKey: types.StringValue("enc-secret")},
			input:    &sdk.AWSEnvSpecFragment_Datadog{Enabled: true, Domain: "datadoghq.com", LogsEnabled: false, MetricsEnabled: true},
			expected: &common.DatadogModel{
				Enabled:        types.BoolValue(true),
				EncAPIKey:      types.StringValue("enc-secret"),
				Domain:         types.StringValue("datadoghq.com"),
				LogsEnabled:    types.BoolValue(false),
				MetricsEnabled: types.BoolValue(true),
			},
		},
		{
			name: "Configured but disabled stays populated",
			existing: &common.DatadogModel{
				Enabled:   types.BoolValue(true),
				EncAPIKey: types.StringValue("enc-secret"),
			},
			input: &sdk.AWSEnvSpecFragment_Datadog{Enabled: false, Domain: "datadoghq.com"},
			expected: &common.DatadogModel{
				Enabled:        types.BoolValue(false),
				EncAPIKey:      types.StringValue("enc-secret"),
				Domain:         types.StringValue("datadoghq.com"),
				LogsEnabled:    types.BoolValue(false),
				MetricsEnabled: types.BoolValue(false),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertDatadogToModel(t, datadogToModel(tt.existing, tt.input), tt.expected)
		})
	}
}

func assertDatadogToModel(t *testing.T, result, expected *common.DatadogModel) {
	t.Helper()
	if (expected == nil) != (result == nil) {
		t.Fatalf("Expected nil: %v, got nil: %v", expected == nil, result == nil)
	}
	if expected == nil {
		return
	}
	if expected.Enabled.ValueBool() != result.Enabled.ValueBool() {
		t.Errorf("Enabled: expected %v, got %v", expected.Enabled.ValueBool(), result.Enabled.ValueBool())
	}
	if expected.EncAPIKey.IsNull() != result.EncAPIKey.IsNull() {
		t.Errorf("EncAPIKey null: expected %v, got %v", expected.EncAPIKey.IsNull(), result.EncAPIKey.IsNull())
	}
	if expected.EncAPIKey.ValueString() != result.EncAPIKey.ValueString() {
		t.Errorf("EncAPIKey: expected %q, got %q", expected.EncAPIKey.ValueString(), result.EncAPIKey.ValueString())
	}
	if expected.Domain.ValueString() != result.Domain.ValueString() {
		t.Errorf("Domain: expected %q, got %q", expected.Domain.ValueString(), result.Domain.ValueString())
	}
	if expected.LogsEnabled.ValueBool() != result.LogsEnabled.ValueBool() {
		t.Errorf("LogsEnabled: expected %v, got %v", expected.LogsEnabled.ValueBool(), result.LogsEnabled.ValueBool())
	}
	if expected.MetricsEnabled.ValueBool() != result.MetricsEnabled.ValueBool() {
		t.Errorf("MetricsEnabled: expected %v, got %v", expected.MetricsEnabled.ValueBool(), result.MetricsEnabled.ValueBool())
	}
}
