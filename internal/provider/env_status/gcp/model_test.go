package env_status

import (
	"testing"

	sdk "github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestGCPEnvStatusModel_toModel(t *testing.T) {
	tests := []struct {
		name     string
		input    sdk.GetGCPEnvStatus_GCPEnv
		expected GCPEnvStatusModel
	}{
		{
			name: "basic gcp env status",
			input: sdk.GetGCPEnvStatus_GCPEnv{
				Name:         "test-gcp-env",
				SpecRevision: 1,
				Status: sdk.GetGCPEnvStatus_GCPEnv_Status{
					AppliedSpecRevision: 1,
					PendingDelete:       false,
				},
			},
			expected: GCPEnvStatusModel{
				Name:                types.StringValue("test-gcp-env"),
				AppliedSpecRevision: types.Int64Value(1),
				PendingDelete:       types.BoolValue(false),
			},
		},
		{
			name: "gcp env status with pending delete",
			input: sdk.GetGCPEnvStatus_GCPEnv{
				Name:         "test-gcp-env-delete",
				SpecRevision: 5,
				Status: sdk.GetGCPEnvStatus_GCPEnv_Status{
					AppliedSpecRevision: 5,
					PendingDelete:       true,
				},
			},
			expected: GCPEnvStatusModel{
				Name:                types.StringValue("test-gcp-env-delete"),
				AppliedSpecRevision: types.Int64Value(5),
				PendingDelete:       types.BoolValue(true),
			},
		},
		{
			name: "gcp env status with large revision numbers",
			input: sdk.GetGCPEnvStatus_GCPEnv{
				Name:         "test-gcp-env-large",
				SpecRevision: 12345,
				Status: sdk.GetGCPEnvStatus_GCPEnv_Status{
					AppliedSpecRevision: 12345,
					PendingDelete:       false,
				},
			},
			expected: GCPEnvStatusModel{
				Name:                types.StringValue("test-gcp-env-large"),
				AppliedSpecRevision: types.Int64Value(12345),
				PendingDelete:       types.BoolValue(false),
			},
		},
		{
			name: "gcp env status with zero revision",
			input: sdk.GetGCPEnvStatus_GCPEnv{
				Name:         "test-gcp-env-zero",
				SpecRevision: 0,
				Status: sdk.GetGCPEnvStatus_GCPEnv_Status{
					AppliedSpecRevision: 0,
					PendingDelete:       false,
				},
			},
			expected: GCPEnvStatusModel{
				Name:                types.StringValue("test-gcp-env-zero"),
				AppliedSpecRevision: types.Int64Value(0),
				PendingDelete:       types.BoolValue(false),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &GCPEnvStatusModel{}
			model.toModel(tt.input)

			assert.Equal(t, tt.expected.Name, model.Name)
			assert.Equal(t, tt.expected.AppliedSpecRevision, model.AppliedSpecRevision)
			assert.Equal(t, tt.expected.PendingDelete, model.PendingDelete)
		})
	}
}
