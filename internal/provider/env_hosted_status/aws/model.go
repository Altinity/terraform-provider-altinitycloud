package hosted_env_status

import (
	sdk "github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AWSEnvHostedStatusModel struct {
	Id                         types.String              `tfsdk:"id"`
	Name                       types.String              `tfsdk:"name"`
	WaitForAppliedSpecRevision types.Int64               `tfsdk:"wait_for_applied_spec_revision"`
	AppliedSpecRevision        types.Int64               `tfsdk:"applied_spec_revision"`
	Verbose                    types.Bool                `tfsdk:"verbose"`
	LoadBalancers              *LoadBalancersStatusModel `tfsdk:"load_balancers"`
	PendingDelete              types.Bool                `tfsdk:"pending_delete"`
	Timeouts                   timeouts.Value            `tfsdk:"timeouts"`
}

type LoadBalancersStatusModel struct {
	Internal *InternalLoadBalancerStatusModel `tfsdk:"internal"`
}

type InternalLoadBalancerStatusModel struct {
	EndpointServiceName types.String `tfsdk:"endpoint_service_name"`
}

func (model *AWSEnvHostedStatusModel) toModel(env sdk.GetAWSEnvHostedStatus_AWSEnvHosted) {
	model.Name = types.StringValue(env.Name)
	model.AppliedSpecRevision = types.Int64Value(env.Status.AppliedSpecRevision)
	model.LoadBalancers = &LoadBalancersStatusModel{
		Internal: &InternalLoadBalancerStatusModel{
			EndpointServiceName: types.StringPointerValue(env.Status.LoadBalancers.Internal.EndpointServiceName),
		},
	}
	model.PendingDelete = types.BoolValue(env.Status.PendingDelete)
}
