package env_status

import (
	"testing"

	sdk "github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestAzureEnvStatusModel_toModel(t *testing.T) {
	tests := []struct {
		name     string
		input    sdk.GetAzureEnvStatus_AzureEnv
		expected AzureEnvStatusModel
	}{
		{
			name: "basic azure env status",
			input: sdk.GetAzureEnvStatus_AzureEnv{
				Name:         "test-azure-env",
				SpecRevision: 1,
				Status: sdk.GetAzureEnvStatus_AzureEnv_Status{
					AppliedSpecRevision: 1,
					PendingDelete:       false,
					LoadBalancers: sdk.GetAzureEnvStatus_AzureEnv_Status_LoadBalancers{
						Internal: sdk.GetAzureEnvStatus_AzureEnv_Status_LoadBalancers_Internal{
							PrivateLinkServiceAlias: stringPtr("test.privatelink.alias"),
						},
					},
				},
			},
			expected: AzureEnvStatusModel{
				Name:                types.StringValue("test-azure-env"),
				AppliedSpecRevision: types.Int64Value(1),
				PendingDelete:       types.BoolValue(false),
				LoadBalancers: &AzureEnvLoadBalancersStatus{
					Internal: &AzureEnvLoadBalancerInternalStatus{
						PrivateLinkServiceAlias: types.StringValue("test.privatelink.alias"),
					},
				},
			},
		},
		{
			name: "azure env status with nil private link service alias",
			input: sdk.GetAzureEnvStatus_AzureEnv{
				Name:         "test-azure-env-nil",
				SpecRevision: 2,
				Status: sdk.GetAzureEnvStatus_AzureEnv_Status{
					AppliedSpecRevision: 2,
					PendingDelete:       true,
					LoadBalancers: sdk.GetAzureEnvStatus_AzureEnv_Status_LoadBalancers{
						Internal: sdk.GetAzureEnvStatus_AzureEnv_Status_LoadBalancers_Internal{
							PrivateLinkServiceAlias: nil,
						},
					},
				},
			},
			expected: AzureEnvStatusModel{
				Name:                types.StringValue("test-azure-env-nil"),
				AppliedSpecRevision: types.Int64Value(2),
				PendingDelete:       types.BoolValue(true),
				LoadBalancers: &AzureEnvLoadBalancersStatus{
					Internal: &AzureEnvLoadBalancerInternalStatus{
						PrivateLinkServiceAlias: types.StringNull(),
					},
				},
			},
		},
		{
			name: "azure env status with large revision numbers",
			input: sdk.GetAzureEnvStatus_AzureEnv{
				Name:         "test-azure-env-large",
				SpecRevision: 9999,
				Status: sdk.GetAzureEnvStatus_AzureEnv_Status{
					AppliedSpecRevision: 9999,
					PendingDelete:       false,
					LoadBalancers: sdk.GetAzureEnvStatus_AzureEnv_Status_LoadBalancers{
						Internal: sdk.GetAzureEnvStatus_AzureEnv_Status_LoadBalancers_Internal{
							PrivateLinkServiceAlias: stringPtr("large.revision.test"),
						},
					},
				},
			},
			expected: AzureEnvStatusModel{
				Name:                types.StringValue("test-azure-env-large"),
				AppliedSpecRevision: types.Int64Value(9999),
				PendingDelete:       types.BoolValue(false),
				LoadBalancers: &AzureEnvLoadBalancersStatus{
					Internal: &AzureEnvLoadBalancerInternalStatus{
						PrivateLinkServiceAlias: types.StringValue("large.revision.test"),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &AzureEnvStatusModel{}
			model.toModel(tt.input)

			assert.Equal(t, tt.expected.Name, model.Name)
			assert.Equal(t, tt.expected.AppliedSpecRevision, model.AppliedSpecRevision)
			assert.Equal(t, tt.expected.PendingDelete, model.PendingDelete)

			// Test LoadBalancers
			if tt.expected.LoadBalancers != nil {
				assert.NotNil(t, model.LoadBalancers)
				if tt.expected.LoadBalancers.Internal != nil {
					assert.NotNil(t, model.LoadBalancers.Internal)
					assert.Equal(t, tt.expected.LoadBalancers.Internal.PrivateLinkServiceAlias, model.LoadBalancers.Internal.PrivateLinkServiceAlias)
				}
			}
		})
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
