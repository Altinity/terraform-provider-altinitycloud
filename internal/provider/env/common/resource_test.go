package env

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Yamashou/gqlgenc/clientv2"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// Passed explicitly instead of shrinking DeletePollInterval, so these tests can
// stay parallel without racing on the package-level default.
const testPollInterval = 50 * time.Millisecond

// The real shape the SDK returns, so IsNotFoundError is exercised the way it is
// called in production.
func notFoundErr() error {
	list := gqlerror.List{{
		Message:    "not found",
		Extensions: map[string]interface{}{"code": "NOT_FOUND"},
	}}
	return &clientv2.ErrorResponse{GqlErrors: &list}
}

func TestWaitForDeletion_NotFoundImmediate(t *testing.T) {
	t.Parallel()
	resp := &resource.DeleteResponse{}
	check := func(ctx context.Context, name string) (bool, error) {
		return false, notFoundErr()
	}

	waitForDeletion(context.Background(), resp, "test-env", false, check, 5*time.Second, 1*time.Second, testPollInterval)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %s", resp.Diagnostics.Errors())
	}
}

func TestWaitForDeletion_DeletingThenNotFound(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	resp := &resource.DeleteResponse{}
	check := func(ctx context.Context, name string) (bool, error) {
		n := calls.Add(1)
		if n <= 2 {
			return true, nil // pendingDelete=true
		}
		return false, notFoundErr()
	}

	waitForDeletion(context.Background(), resp, "test-env", false, check, 5*time.Second, 1*time.Second, testPollInterval)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %s", resp.Diagnostics.Errors())
	}
}

func TestWaitForDeletion_NoMfaNoPendingDelete_ReturnsDeleted(t *testing.T) {
	t.Parallel()
	resp := &resource.DeleteResponse{}
	check := func(ctx context.Context, name string) (bool, error) {
		return false, nil // pendingDelete=false, no error (not 404)
	}

	start := time.Now()
	waitForDeletion(context.Background(), resp, "test-env", false, check, 5*time.Second, 1*time.Second, testPollInterval)
	elapsed := time.Since(start)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %s", resp.Diagnostics.Errors())
	}
	if elapsed > 2*time.Second {
		t.Errorf("expected quick return, took %s (would hang without fix)", elapsed)
	}
}

func TestWaitForDeletion_MfaPendingThenConfirmed(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	resp := &resource.DeleteResponse{}
	check := func(ctx context.Context, name string) (bool, error) {
		n := calls.Add(1)
		if n <= 2 {
			return false, nil // waiting for MFA
		}
		if n <= 4 {
			return true, nil // MFA confirmed, deleting
		}
		return false, notFoundErr()
	}

	waitForDeletion(context.Background(), resp, "test-env", true, check, 5*time.Second, 5*time.Second, testPollInterval)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %s", resp.Diagnostics.Errors())
	}
}

func TestWaitForDeletion_MfaTimeout(t *testing.T) {
	t.Parallel()
	resp := &resource.DeleteResponse{}
	check := func(ctx context.Context, name string) (bool, error) {
		return false, nil // pendingDelete stays false forever
	}

	waitForDeletion(context.Background(), resp, "test-env", true, check, 5*time.Second, 200*time.Millisecond, testPollInterval)

	if !resp.Diagnostics.HasError() {
		t.Error("expected MFA timeout error, got none")
	}
}

func TestWaitForDeletion_NonNotFoundError(t *testing.T) {
	t.Parallel()
	resp := &resource.DeleteResponse{}
	check := func(ctx context.Context, name string) (bool, error) {
		return false, fmt.Errorf("connection refused")
	}

	waitForDeletion(context.Background(), resp, "test-env", false, check, 5*time.Second, 1*time.Second, testPollInterval)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error, got none")
	}
}

func TestWaitForDeletion_TransientErrorThenDeleted(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	resp := &resource.DeleteResponse{}
	check := func(ctx context.Context, name string) (bool, error) {
		n := calls.Add(1)
		if n == 1 {
			return true, fmt.Errorf("HTTP 503 Service Unavailable")
		}
		if n == 2 {
			return true, nil
		}
		return false, notFoundErr()
	}

	waitForDeletion(context.Background(), resp, "test-env", false, check, 5*time.Second, 1*time.Second, testPollInterval)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %s", resp.Diagnostics.Errors())
	}
	if calls.Load() < 3 {
		t.Errorf("expected status check to run until not-found after 503, got %d calls", calls.Load())
	}
}

func TestWaitForDeletion_ErrEnvNotFoundIsDeleted(t *testing.T) {
	t.Parallel()
	resp := &resource.DeleteResponse{}
	check := func(ctx context.Context, name string) (bool, error) {
		return false, ErrEnvNotFound
	}

	waitForDeletion(context.Background(), resp, "test-env", false, check, 5*time.Second, 1*time.Second, testPollInterval)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %s", resp.Diagnostics.Errors())
	}
}

// A null env means gone, so a pendingMFA delete must finish instead of sitting in
// PENDING_MFA until the MFA timeout (which returning (false, nil) would cause).
func TestWaitForDeletion_ErrEnvNotFoundWithPendingMfaIsDeleted(t *testing.T) {
	t.Parallel()
	resp := &resource.DeleteResponse{}
	check := func(ctx context.Context, name string) (bool, error) {
		return false, ErrEnvNotFound
	}

	waitForDeletion(context.Background(), resp, "test-env", true, check, 5*time.Second, 200*time.Millisecond, testPollInterval)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %s", resp.Diagnostics.Errors())
	}
}

// Wrapped sentinels must still count as deleted.
func TestWaitForDeletion_WrappedErrEnvNotFoundIsDeleted(t *testing.T) {
	t.Parallel()
	resp := &resource.DeleteResponse{}
	check := func(ctx context.Context, name string) (bool, error) {
		return false, fmt.Errorf("polling env status: %w", ErrEnvNotFound)
	}

	waitForDeletion(context.Background(), resp, "test-env", true, check, 5*time.Second, 200*time.Millisecond, testPollInterval)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %s", resp.Diagnostics.Errors())
	}
}
