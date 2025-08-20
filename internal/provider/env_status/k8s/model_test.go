package env_status

import (
	"testing"

	sdk "github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestK8SEnvStatusModel_toModel(t *testing.T) {
	tests := []struct {
		name     string
		input    sdk.GetK8SEnvStatus_K8sEnv
		expected K8SEnvStatusModel
	}{
		{
			name: "basic k8s env status",
			input: sdk.GetK8SEnvStatus_K8sEnv{
				Name:         "test-k8s-env",
				SpecRevision: 1,
				Status: sdk.GetK8SEnvStatus_K8sEnv_Status{
					AppliedSpecRevision: 1,
					PendingDelete:       false,
				},
			},
			expected: K8SEnvStatusModel{
				Name:                types.StringValue("test-k8s-env"),
				AppliedSpecRevision: types.Int64Value(1),
				PendingDelete:       types.BoolValue(false),
			},
		},
		{
			name: "k8s env status with pending delete",
			input: sdk.GetK8SEnvStatus_K8sEnv{
				Name:         "test-k8s-env-delete",
				SpecRevision: 42,
				Status: sdk.GetK8SEnvStatus_K8sEnv_Status{
					AppliedSpecRevision: 42,
					PendingDelete:       true,
				},
			},
			expected: K8SEnvStatusModel{
				Name:                types.StringValue("test-k8s-env-delete"),
				AppliedSpecRevision: types.Int64Value(42),
				PendingDelete:       types.BoolValue(true),
			},
		},
		{
			name: "k8s env status with large revision numbers",
			input: sdk.GetK8SEnvStatus_K8sEnv{
				Name:         "test-k8s-env-large",
				SpecRevision: 999999,
				Status: sdk.GetK8SEnvStatus_K8sEnv_Status{
					AppliedSpecRevision: 999999,
					PendingDelete:       false,
				},
			},
			expected: K8SEnvStatusModel{
				Name:                types.StringValue("test-k8s-env-large"),
				AppliedSpecRevision: types.Int64Value(999999),
				PendingDelete:       types.BoolValue(false),
			},
		},
		{
			name: "k8s env status with zero revision",
			input: sdk.GetK8SEnvStatus_K8sEnv{
				Name:         "test-k8s-env-zero",
				SpecRevision: 0,
				Status: sdk.GetK8SEnvStatus_K8sEnv_Status{
					AppliedSpecRevision: 0,
					PendingDelete:       false,
				},
			},
			expected: K8SEnvStatusModel{
				Name:                types.StringValue("test-k8s-env-zero"),
				AppliedSpecRevision: types.Int64Value(0),
				PendingDelete:       types.BoolValue(false),
			},
		},
		{
			name: "k8s env status with special characters in name",
			input: sdk.GetK8SEnvStatus_K8sEnv{
				Name:         "test-k8s-env-special_chars.123",
				SpecRevision: 15,
				Status: sdk.GetK8SEnvStatus_K8sEnv_Status{
					AppliedSpecRevision: 15,
					PendingDelete:       false,
				},
			},
			expected: K8SEnvStatusModel{
				Name:                types.StringValue("test-k8s-env-special_chars.123"),
				AppliedSpecRevision: types.Int64Value(15),
				PendingDelete:       types.BoolValue(false),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &K8SEnvStatusModel{}
			model.toModel(tt.input)

			assert.Equal(t, tt.expected.Name, model.Name)
			assert.Equal(t, tt.expected.AppliedSpecRevision, model.AppliedSpecRevision)
			assert.Equal(t, tt.expected.PendingDelete, model.PendingDelete)
		})
	}
}
