package env

import (
	"errors"
	"strings"
	"testing"
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

func TestValidateDisconnected(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		errorCode               string
		appliedSpecRevision     int64
		skipDeprovision         bool
		allowDeleteDisconnected bool
		expectErr               bool
		containsMsg             string
	}{
		"DISCONNECTED rev=0 blocks with never-provisioned message": {
			errorCode:           "DISCONNECTED",
			appliedSpecRevision: 0,
			expectErr:           true,
			containsMsg:         "never fully provisioned",
		},
		"DISCONNECTED rev>0 blocks with cloudconnect message": {
			errorCode:           "DISCONNECTED",
			appliedSpecRevision: 5,
			expectErr:           true,
			containsMsg:         "cloudconnect",
		},
		"K8S_DISCONNECTED rev=0 blocks with never-provisioned message": {
			errorCode:           "K8S_DISCONNECTED",
			appliedSpecRevision: 0,
			expectErr:           true,
			containsMsg:         "never fully provisioned",
		},
		"K8S_DISCONNECTED rev>0 blocks with cloudconnect message": {
			errorCode:           "K8S_DISCONNECTED",
			appliedSpecRevision: 3,
			expectErr:           true,
			containsMsg:         "cloudconnect",
		},
		"DISCONNECTED with allow_delete_while_disconnected passes": {
			errorCode:               "DISCONNECTED",
			appliedSpecRevision:     0,
			allowDeleteDisconnected: true,
			expectErr:               false,
		},
		"DISCONNECTED with skip_deprovision passes": {
			errorCode:           "DISCONNECTED",
			appliedSpecRevision: 5,
			skipDeprovision:     true,
			expectErr:           false,
		},
		"non-DISCONNECTED error code is ignored": {
			errorCode:           "SOME_OTHER_ERROR",
			appliedSpecRevision: 0,
			expectErr:           false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			diags := ValidateDisconnected("test-env", tc.errorCode, tc.appliedSpecRevision, tc.skipDeprovision, tc.allowDeleteDisconnected)
			if tc.expectErr && !diags.HasError() {
				t.Error("expected error, got none")
			}
			if !tc.expectErr && diags.HasError() {
				t.Errorf("unexpected error: %s", diags.Errors())
			}
			if tc.expectErr && tc.containsMsg != "" {
				errMsg := diags.Errors()[0].Detail()
				if !strings.Contains(errMsg, tc.containsMsg) {
					t.Errorf("error message %q should contain %q", errMsg, tc.containsMsg)
				}
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
