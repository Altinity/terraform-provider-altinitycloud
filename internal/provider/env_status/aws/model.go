package env_status

import (
	sdk "github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AWSEnvStatusModel struct {
	Id                         types.String                    `tfsdk:"id"`
	Name                       types.String                    `tfsdk:"name"`
	WaitForAppliedSpecRevision types.Int64                     `tfsdk:"wait_for_applied_spec_revision"`
	AppliedSpecRevision        types.Int64                     `tfsdk:"applied_spec_revision"`
	LoadBalancers              *AWSEnvLoadBalancersStatus      `tfsdk:"load_balancers"`
	PeeringConnections         []AWSEnvPeeringConnectionStatus `tfsdk:"peering_connections"`
	PendingDelete              types.Bool                      `tfsdk:"pending_delete"`
}

type AWSEnvLoadBalancersStatus struct {
	Internal *AWSEnvLoadBalancerInternalStatus `tfsdk:"internal"`
}

type AWSEnvLoadBalancerInternalStatus struct {
	EndpointServiceName types.String `tfsdk:"endpoint_service_name"`
}

type AWSEnvPeeringConnectionStatus struct {
	Id    types.String `tfsdk:"id"`
	VpcId types.String `tfsdk:"vpc_id"`
}

func (model *AWSEnvStatusModel) toModel(env sdk.GetAWSEnvStatus_AwsEnv) {
	model.Name = types.StringValue(env.Name)
	model.AppliedSpecRevision = types.Int64Value(env.Status.AppliedSpecRevision)
	model.LoadBalancers = &AWSEnvLoadBalancersStatus{
		Internal: &AWSEnvLoadBalancerInternalStatus{
			EndpointServiceName: types.StringPointerValue(env.Status.LoadBalancers.Internal.EndpointServiceName),
		},
	}
	var peeringConnections []AWSEnvPeeringConnectionStatus
	for _, p := range env.Status.PeeringConnections {
		peeringConnections = append(peeringConnections, AWSEnvPeeringConnectionStatus{
			Id:    types.StringPointerValue(p.ID),
			VpcId: types.StringValue(p.VpcID),
		})
	}
	model.PeeringConnections = peeringConnections
	model.PendingDelete = types.BoolValue(env.Status.PendingDelete)
}
