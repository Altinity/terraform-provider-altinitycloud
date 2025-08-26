package env_status

import (
	"testing"

	sdk "github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestAWSEnvStatusModel_toModel(t *testing.T) {
	tests := []struct {
		name     string
		input    sdk.GetAWSEnvStatus_AWSEnv
		expected AWSEnvStatusModel
	}{
		{
			name: "basic aws env status",
			input: sdk.GetAWSEnvStatus_AWSEnv{
				Name:         "test-env",
				SpecRevision: 1,
				Status: sdk.GetAWSEnvStatus_AWSEnv_Status{
					AppliedSpecRevision: 1,
					PendingDelete:       false,
					LoadBalancers: sdk.GetAWSEnvStatus_AWSEnv_Status_LoadBalancers{
						Internal: sdk.GetAWSEnvStatus_AWSEnv_Status_LoadBalancers_Internal{
							EndpointServiceName: stringPtr("test.service.name"),
						},
					},
					PeeringConnections: []*sdk.GetAWSEnvStatus_AWSEnv_Status_PeeringConnections{
						{
							ID:    stringPtr("pcx-12345"),
							VpcID: "vpc-12345",
						},
					},
				},
			},
			expected: AWSEnvStatusModel{
				Name:                types.StringValue("test-env"),
				AppliedSpecRevision: types.Int64Value(1),
				PendingDelete:       types.BoolValue(false),
				LoadBalancers: &AWSEnvLoadBalancersStatus{
					Internal: &AWSEnvLoadBalancerInternalStatus{
						EndpointServiceName: types.StringValue("test.service.name"),
					},
				},
				PeeringConnections: []AWSEnvPeeringConnectionStatus{
					{
						Id:    types.StringValue("pcx-12345"),
						VpcId: types.StringValue("vpc-12345"),
					},
				},
			},
		},
		{
			name: "env status with nil endpoint service name",
			input: sdk.GetAWSEnvStatus_AWSEnv{
				Name:         "test-env-nil",
				SpecRevision: 2,
				Status: sdk.GetAWSEnvStatus_AWSEnv_Status{
					AppliedSpecRevision: 2,
					PendingDelete:       true,
					LoadBalancers: sdk.GetAWSEnvStatus_AWSEnv_Status_LoadBalancers{
						Internal: sdk.GetAWSEnvStatus_AWSEnv_Status_LoadBalancers_Internal{
							EndpointServiceName: nil,
						},
					},
					PeeringConnections: []*sdk.GetAWSEnvStatus_AWSEnv_Status_PeeringConnections{},
				},
			},
			expected: AWSEnvStatusModel{
				Name:                types.StringValue("test-env-nil"),
				AppliedSpecRevision: types.Int64Value(2),
				PendingDelete:       types.BoolValue(true),
				LoadBalancers: &AWSEnvLoadBalancersStatus{
					Internal: &AWSEnvLoadBalancerInternalStatus{
						EndpointServiceName: types.StringNull(),
					},
				},
				PeeringConnections: []AWSEnvPeeringConnectionStatus{},
			},
		},
		{
			name: "env status with multiple peering connections",
			input: sdk.GetAWSEnvStatus_AWSEnv{
				Name:         "test-env-multi",
				SpecRevision: 3,
				Status: sdk.GetAWSEnvStatus_AWSEnv_Status{
					AppliedSpecRevision: 3,
					PendingDelete:       false,
					LoadBalancers: sdk.GetAWSEnvStatus_AWSEnv_Status_LoadBalancers{
						Internal: sdk.GetAWSEnvStatus_AWSEnv_Status_LoadBalancers_Internal{
							EndpointServiceName: stringPtr("multi.service.name"),
						},
					},
					PeeringConnections: []*sdk.GetAWSEnvStatus_AWSEnv_Status_PeeringConnections{
						{
							ID:    stringPtr("pcx-11111"),
							VpcID: "vpc-11111",
						},
						{
							ID:    stringPtr("pcx-22222"),
							VpcID: "vpc-22222",
						},
						{
							ID:    nil,
							VpcID: "vpc-33333",
						},
					},
				},
			},
			expected: AWSEnvStatusModel{
				Name:                types.StringValue("test-env-multi"),
				AppliedSpecRevision: types.Int64Value(3),
				PendingDelete:       types.BoolValue(false),
				LoadBalancers: &AWSEnvLoadBalancersStatus{
					Internal: &AWSEnvLoadBalancerInternalStatus{
						EndpointServiceName: types.StringValue("multi.service.name"),
					},
				},
				PeeringConnections: []AWSEnvPeeringConnectionStatus{
					{
						Id:    types.StringValue("pcx-11111"),
						VpcId: types.StringValue("vpc-11111"),
					},
					{
						Id:    types.StringValue("pcx-22222"),
						VpcId: types.StringValue("vpc-22222"),
					},
					{
						Id:    types.StringNull(),
						VpcId: types.StringValue("vpc-33333"),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &AWSEnvStatusModel{}
			model.toModel(tt.input)

			assert.Equal(t, tt.expected.Name, model.Name)
			assert.Equal(t, tt.expected.AppliedSpecRevision, model.AppliedSpecRevision)
			assert.Equal(t, tt.expected.PendingDelete, model.PendingDelete)

			if tt.expected.LoadBalancers != nil {
				assert.NotNil(t, model.LoadBalancers)
				if tt.expected.LoadBalancers.Internal != nil {
					assert.NotNil(t, model.LoadBalancers.Internal)
					assert.Equal(t, tt.expected.LoadBalancers.Internal.EndpointServiceName, model.LoadBalancers.Internal.EndpointServiceName)
				}
			}

			assert.Equal(t, len(tt.expected.PeeringConnections), len(model.PeeringConnections))
			for i, expectedConn := range tt.expected.PeeringConnections {
				if i < len(model.PeeringConnections) {
					assert.Equal(t, expectedConn.Id, model.PeeringConnections[i].Id)
					assert.Equal(t, expectedConn.VpcId, model.PeeringConnections[i].VpcId)
				}
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
