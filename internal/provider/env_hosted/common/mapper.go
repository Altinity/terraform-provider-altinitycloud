package hosted_env

import (
	common "github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/common"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func MetricsEndpointToSDK(endpoint *MetricsEndpointModel) *client.MetricsEndpointSpecInput {
	if endpoint == nil {
		return nil
	}

	return &client.MetricsEndpointSpecInput{
		Enabled:        endpoint.Enabled.ValueBoolPointer(),
		SourceIPRanges: common.ListStringToSDK(endpoint.SourceIPRanges),
	}
}

// The API always returns a metrics endpoint block. Keep state null when the user
// never configured it and it's disabled, to avoid a perpetual diff.
func MetricsEndpointToModel(existing *MetricsEndpointModel, enabled bool, sourceIPRanges []string) *MetricsEndpointModel {
	if existing == nil && !enabled {
		return nil
	}

	return &MetricsEndpointModel{
		Enabled:        types.BoolValue(enabled),
		SourceIPRanges: common.ListStringToModel(sourceIPRanges),
	}
}

// The API always returns a datadog block (DatadogSpec!), so the same null-preserving
// rule as the metrics endpoint applies. enc_api_key is write-only and never returned.
func DatadogToModel(existing *common.DatadogModel, enabled bool, domain string, logsEnabled, metricsEnabled bool) *common.DatadogModel {
	if existing == nil && !enabled {
		return nil
	}

	model := &common.DatadogModel{
		Enabled:        types.BoolValue(enabled),
		Domain:         types.StringValue(domain),
		LogsEnabled:    types.BoolValue(logsEnabled),
		MetricsEnabled: types.BoolValue(metricsEnabled),
	}

	if existing != nil {
		model.EncAPIKey = existing.EncAPIKey
	}

	return model
}

// Mirrors CustomDomainsToModel: a resource that never configured the attribute must
// keep it null, otherwise the API's empty list turns into a permanent diff.
func CustomDomainsToModel(prior types.List, specCustomDomains []string) (types.List, diag.Diagnostics) {
	if prior.IsNull() {
		return types.ListNull(types.StringType), nil
	}

	return common.ListToModel(specCustomDomains)
}

// The API only guarantees uniqueness on name; fall back to node_type while the computed name is still unknown.
func NodeGroupKey(ng NodeGroupsModel) string {
	if ng.Name.IsNull() || ng.Name.IsUnknown() || ng.Name.ValueString() == "" {
		return ng.NodeType.ValueString()
	}

	return ng.Name.ValueString()
}
