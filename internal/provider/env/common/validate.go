package env

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

// ValidateDatadog enforces that enc_api_key is provided when the Datadog agent
// is enabled. Unknown values are skipped: enc_api_key commonly references a
// secret resource whose value is not known until apply, so it can only be
// checked once resolved.
func ValidateDatadog(datadog *DatadogModel) diag.Diagnostics {
	var diags diag.Diagnostics
	if datadog == nil {
		return diags
	}

	if datadog.Enabled.IsNull() || datadog.Enabled.IsUnknown() || !datadog.Enabled.ValueBool() {
		return diags
	}

	if datadog.EncAPIKey.IsUnknown() {
		return diags
	}

	if datadog.EncAPIKey.IsNull() || datadog.EncAPIKey.ValueString() == "" {
		diags.AddAttributeError(
			path.Root("datadog").AtName("enc_api_key"),
			"Missing Datadog API key",
			"enc_api_key must be set when datadog.enabled is true.",
		)
	}

	return diags
}
