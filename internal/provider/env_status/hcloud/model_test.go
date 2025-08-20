package env_status

import (
	"testing"

	sdk "github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestHCloudEnvStatusModel_toModel(t *testing.T) {
	tests := []struct {
		name     string
		input    sdk.GetHCloudEnvStatus_HcloudEnv
		expected HCloudEnvStatusModel
	}{
		{
			name: "basic hcloud env status",
			input: sdk.GetHCloudEnvStatus_HcloudEnv{
				Name:         "test-hcloud-env",
				SpecRevision: 1,
				Status: sdk.GetHCloudEnvStatus_HcloudEnv_Status{
					AppliedSpecRevision: 1,
					PendingDelete:       false,
				},
			},
			expected: HCloudEnvStatusModel{
				Name:                types.StringValue("test-hcloud-env"),
				AppliedSpecRevision: types.Int64Value(1),
				PendingDelete:       types.BoolValue(false),
			},
		},
		{
			name: "hcloud env status with pending delete",
			input: sdk.GetHCloudEnvStatus_HcloudEnv{
				Name:         "test-hcloud-env-delete",
				SpecRevision: 7,
				Status: sdk.GetHCloudEnvStatus_HcloudEnv_Status{
					AppliedSpecRevision: 7,
					PendingDelete:       true,
				},
			},
			expected: HCloudEnvStatusModel{
				Name:                types.StringValue("test-hcloud-env-delete"),
				AppliedSpecRevision: types.Int64Value(7),
				PendingDelete:       types.BoolValue(true),
			},
		},
		{
			name: "hcloud env status with large revision numbers",
			input: sdk.GetHCloudEnvStatus_HcloudEnv{
				Name:         "test-hcloud-env-large",
				SpecRevision: 54321,
				Status: sdk.GetHCloudEnvStatus_HcloudEnv_Status{
					AppliedSpecRevision: 54321,
					PendingDelete:       false,
				},
			},
			expected: HCloudEnvStatusModel{
				Name:                types.StringValue("test-hcloud-env-large"),
				AppliedSpecRevision: types.Int64Value(54321),
				PendingDelete:       types.BoolValue(false),
			},
		},
		{
			name: "hcloud env status with zero revision",
			input: sdk.GetHCloudEnvStatus_HcloudEnv{
				Name:         "test-hcloud-env-zero",
				SpecRevision: 0,
				Status: sdk.GetHCloudEnvStatus_HcloudEnv_Status{
					AppliedSpecRevision: 0,
					PendingDelete:       false,
				},
			},
			expected: HCloudEnvStatusModel{
				Name:                types.StringValue("test-hcloud-env-zero"),
				AppliedSpecRevision: types.Int64Value(0),
				PendingDelete:       types.BoolValue(false),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &HCloudEnvStatusModel{}
			model.toModel(tt.input)

			assert.Equal(t, tt.expected.Name, model.Name)
			assert.Equal(t, tt.expected.AppliedSpecRevision, model.AppliedSpecRevision)
			assert.Equal(t, tt.expected.PendingDelete, model.PendingDelete)
		})
	}
}
