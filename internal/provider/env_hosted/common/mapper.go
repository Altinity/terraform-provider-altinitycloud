package hosted_env

import (
	common "github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/common"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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
