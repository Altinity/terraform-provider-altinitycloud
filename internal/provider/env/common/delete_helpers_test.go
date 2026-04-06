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
