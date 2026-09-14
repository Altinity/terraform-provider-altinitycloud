package env

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// enc_api_key often references a secret resource, so unknown values are skipped until apply resolves them.
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

// Names must be unique per env and an omitted name defaults to node_type server-side, so both collide the same way.
func ValidateNodeGroupNames(ctx context.Context, config tfsdk.Config) diag.Diagnostics {
	var diags diag.Diagnostics

	schemaType, ok := config.Schema.Type().(types.ObjectType)
	if !ok {
		return diags
	}
	attributeType, ok := schemaType.AttrTypes["node_groups"]
	if !ok {
		return diags
	}

	// node_groups is a list on most envs and a set on hcloud, which also changes the path step.
	_, isSet := attributeType.(types.SetType)
	elements, diags := nodeGroupElements(ctx, config, isSet)
	if diags.HasError() {
		return diags
	}

	seen := make(map[string]bool, len(elements))
	for i, element := range elements {
		group, ok := element.(types.Object)
		if !ok || group.IsNull() || group.IsUnknown() {
			continue
		}

		attributes := group.Attributes()
		name, _ := attributes["name"].(types.String)
		nodeType, _ := attributes["node_type"].(types.String)
		if name.IsUnknown() || nodeType.IsUnknown() {
			continue
		}

		key := name.ValueString()
		if key == "" {
			key = nodeType.ValueString()
		}
		if key == "" {
			continue
		}

		if seen[key] {
			groupPath := path.Root("node_groups").AtListIndex(i)
			if isSet {
				groupPath = path.Root("node_groups").AtSetValue(element)
			}
			diags.AddAttributeError(
				groupPath.AtName("name"),
				"Duplicate node group name",
				fmt.Sprintf("More than one node group resolves to the name %q, but node group names must be unique within an environment. An omitted name defaults to node_type, so set a distinct name on each group that reuses a node_type.", key),
			)
			continue
		}
		seen[key] = true
	}

	return diags
}

func nodeGroupElements(ctx context.Context, config tfsdk.Config, isSet bool) ([]attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	if isSet {
		var nodeGroups types.Set
		diags.Append(config.GetAttribute(ctx, path.Root("node_groups"), &nodeGroups)...)
		if diags.HasError() || nodeGroups.IsNull() || nodeGroups.IsUnknown() {
			return nil, diags
		}
		return nodeGroups.Elements(), diags
	}

	var nodeGroups types.List
	diags.Append(config.GetAttribute(ctx, path.Root("node_groups"), &nodeGroups)...)
	if diags.HasError() || nodeGroups.IsNull() || nodeGroups.IsUnknown() {
		return nil, diags
	}

	return nodeGroups.Elements(), diags
}
