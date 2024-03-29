package env_status

import (
	sdk "github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AzureEnvStatusModel struct {
	Id                         types.String                 `tfsdk:"id"`
	Name                       types.String                 `tfsdk:"name"`
	WaitForAppliedSpecRevision types.Int64                  `tfsdk:"wait_for_applied_spec_revision"`
	AppliedSpecRevision        types.Int64                  `tfsdk:"applied_spec_revision"`
	PendingDelete              types.Bool                   `tfsdk:"pending_delete"`
	LoadBalancers              *AzureEnvLoadBalancersStatus `tfsdk:"load_balancers"`
}

type AzureEnvLoadBalancersStatus struct {
	Internal *AzureEnvLoadBalancerInternalStatus `tfsdk:"internal"`
}

type AzureEnvLoadBalancerInternalStatus struct {
	PrivateLinkServiceAlias types.String `tfsdk:"private_link_service_alias"`
}

func (model *AzureEnvStatusModel) toModel(env sdk.GetAzureEnvStatus_AzureEnv) {
	model.Name = types.StringValue(env.Name)
	model.AppliedSpecRevision = types.Int64Value(env.Status.AppliedSpecRevision)
	model.PendingDelete = types.BoolValue(env.Status.PendingDelete)

	model.LoadBalancers = &AzureEnvLoadBalancersStatus{
		Internal: &AzureEnvLoadBalancerInternalStatus{
			PrivateLinkServiceAlias: types.StringPointerValue(env.Status.LoadBalancers.Internal.PrivateLinkServiceAlias),
		},
	}
}
