package hosted_env_status

import (
	"testing"

	sdk "github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func stringPtr(s string) *string { return &s }

func TestAWSEnvHostedStatusModel_toModel(t *testing.T) {
	tests := []struct {
		name     string
		input    sdk.GetAWSEnvHostedStatus_AWSEnvHosted
		expected AWSEnvHostedStatusModel
	}{
		{
			name: "basic hosted aws env status",
			input: sdk.GetAWSEnvHostedStatus_AWSEnvHosted{
				Name:         "test-env",
				SpecRevision: 1,
				Status: sdk.GetAWSEnvHostedStatus_AWSEnvHosted_Status{
					AppliedSpecRevision: 1,
					PendingDelete:       false,
					LoadBalancers: sdk.GetAWSEnvHostedStatus_AWSEnvHosted_Status_LoadBalancers{
						Internal: sdk.GetAWSEnvHostedStatus_AWSEnvHosted_Status_LoadBalancers_Internal{
							EndpointServiceName: stringPtr("test.service.name"),
						},
					},
				},
			},
			expected: AWSEnvHostedStatusModel{
				Name:                types.StringValue("test-env"),
				AppliedSpecRevision: types.Int64Value(1),
				PendingDelete:       types.BoolValue(false),
				LoadBalancers: &LoadBalancersStatusModel{
					Internal: &InternalLoadBalancerStatusModel{
						EndpointServiceName: types.StringValue("test.service.name"),
					},
				},
			},
		},
		{
			name: "env pending delete without an endpoint service",
			input: sdk.GetAWSEnvHostedStatus_AWSEnvHosted{
				Name:         "test-env",
				SpecRevision: 4,
				Status: sdk.GetAWSEnvHostedStatus_AWSEnvHosted_Status{
					AppliedSpecRevision: 3,
					PendingDelete:       true,
				},
			},
			expected: AWSEnvHostedStatusModel{
				Name:                types.StringValue("test-env"),
				AppliedSpecRevision: types.Int64Value(3),
				PendingDelete:       types.BoolValue(true),
				LoadBalancers: &LoadBalancersStatusModel{
					Internal: &InternalLoadBalancerStatusModel{
						EndpointServiceName: types.StringNull(),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var model AWSEnvHostedStatusModel
			model.toModel(tt.input)

			assert.Equal(t, tt.expected.Name, model.Name)
			assert.Equal(t, tt.expected.AppliedSpecRevision, model.AppliedSpecRevision)
			assert.Equal(t, tt.expected.PendingDelete, model.PendingDelete)
			assert.Equal(t, tt.expected.LoadBalancers, model.LoadBalancers)
		})
	}
}
