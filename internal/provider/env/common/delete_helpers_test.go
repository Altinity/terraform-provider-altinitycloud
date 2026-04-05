package env

import (
	"errors"
	"strings"
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
)

func TestValidateForceDestroy(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		forceDestroy bool
		expectErr    bool
	}{
		"force_destroy=false blocks deletion": {
			forceDestroy: false,
			expectErr:    true,
		},
		"force_destroy=true allows deletion": {
			forceDestroy: true,
			expectErr:    false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			diags := ValidateForceDestroy("test-env", tc.forceDestroy)
			if tc.expectErr && !diags.HasError() {
				t.Error("expected error, got none")
			}
			if !tc.expectErr && diags.HasError() {
				t.Errorf("unexpected error: %s", diags.Errors())
			}
		})
	}
}

func TestHasBlockingDisconnectedError(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		errorCodes                   []client.EnvStatusErrorCode
		skipDeprovision              bool
		allowDeleteWhileDisconnected bool
		expected                     bool
	}{
		"no errors": {
			errorCodes: nil,
			expected:   false,
		},
		"disconnected + both flags false (blocked)": {
			errorCodes:                   []client.EnvStatusErrorCode{client.EnvStatusErrorCodeDisconnected},
			skipDeprovision:              false,
			allowDeleteWhileDisconnected: false,
			expected:                     true,
		},
		"disconnected + skip_deprovision=true (allowed)": {
			errorCodes:                   []client.EnvStatusErrorCode{client.EnvStatusErrorCodeDisconnected},
			skipDeprovision:              true,
			allowDeleteWhileDisconnected: false,
			expected:                     false,
		},
		"disconnected + allow_delete=true (allowed)": {
			errorCodes:                   []client.EnvStatusErrorCode{client.EnvStatusErrorCodeDisconnected},
			skipDeprovision:              false,
			allowDeleteWhileDisconnected: true,
			expected:                     false,
		},
		"disconnected + both flags true (allowed)": {
			errorCodes:                   []client.EnvStatusErrorCode{client.EnvStatusErrorCodeDisconnected},
			skipDeprovision:              true,
			allowDeleteWhileDisconnected: true,
			expected:                     false,
		},
		"k8s_disconnected + both flags false (blocked)": {
			errorCodes:                   []client.EnvStatusErrorCode{client.EnvStatusErrorCodeK8sDisconnected},
			skipDeprovision:              false,
			allowDeleteWhileDisconnected: false,
			expected:                     true,
		},
		"k8s_disconnected + skip_deprovision=true (allowed)": {
			errorCodes:                   []client.EnvStatusErrorCode{client.EnvStatusErrorCodeK8sDisconnected},
			skipDeprovision:              true,
			allowDeleteWhileDisconnected: false,
			expected:                     false,
		},
		"k8s_disconnected + allow_delete=true (allowed)": {
			errorCodes:                   []client.EnvStatusErrorCode{client.EnvStatusErrorCodeK8sDisconnected},
			skipDeprovision:              false,
			allowDeleteWhileDisconnected: true,
			expected:                     false,
		},
		"non-disconnected error (allowed)": {
			errorCodes:                   []client.EnvStatusErrorCode{client.EnvStatusErrorCodeInternal},
			skipDeprovision:              false,
			allowDeleteWhileDisconnected: false,
			expected:                     false,
		},
		"multiple errors with disconnected (blocked)": {
			errorCodes:                   []client.EnvStatusErrorCode{client.EnvStatusErrorCodeInternal, client.EnvStatusErrorCodeDisconnected},
			skipDeprovision:              false,
			allowDeleteWhileDisconnected: false,
			expected:                     true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := HasBlockingDisconnectedError(tc.errorCodes, tc.skipDeprovision, tc.allowDeleteWhileDisconnected)
			if got != tc.expected {
				t.Errorf("got %v, want %v", got, tc.expected)
			}
		})
	}
}

func TestFormatDeleteError(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		err      error
		contains string
	}{
		"active clusters error": {
			err:      errors.New(`{"networkErrors":null,"graphqlErrors":[{"message":"env has active clusters, use forceDestroyClusters=true","path":["deleteAWSEnv"],"extensions":{"code":"CONFLICT"}}]}`),
			contains: "force_destroy_clusters=true",
		},
		"generic error": {
			err:      errors.New("connection refused"),
			contains: "connection refused",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := FormatDeleteError("test-env", tc.err)
			if !strings.Contains(got, tc.contains) {
				t.Errorf("got %q, want to contain %q", got, tc.contains)
			}
		})
	}
}
