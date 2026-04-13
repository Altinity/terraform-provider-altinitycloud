package env

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func notFoundErr() error {
	return fmt.Errorf(`{"networkErrors":null,"graphqlErrors":[{"message":"not found","path":["env"],"extensions":{"code":"NOT_FOUND"}}]}`)
}

func TestWaitForDeletion_NotFoundImmediate(t *testing.T) {
	t.Parallel()
	origInterval := DeletePollInterval
	DeletePollInterval = 50 * time.Millisecond
	defer func() { DeletePollInterval = origInterval }()

	resp := &resource.DeleteResponse{}
	check := func(ctx context.Context, name string) (bool, error) {
		return false, notFoundErr()
	}

	WaitForDeletion(context.Background(), resp, "test-env", false, check, 5*time.Second, 1*time.Second)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %s", resp.Diagnostics.Errors())
	}
}

func TestWaitForDeletion_DeletingThenNotFound(t *testing.T) {
	t.Parallel()
	origInterval := DeletePollInterval
	DeletePollInterval = 50 * time.Millisecond
	defer func() { DeletePollInterval = origInterval }()

	var calls atomic.Int32
	resp := &resource.DeleteResponse{}
	check := func(ctx context.Context, name string) (bool, error) {
		n := calls.Add(1)
		if n <= 2 {
			return true, nil // pendingDelete=true
		}
		return false, notFoundErr()
	}

	WaitForDeletion(context.Background(), resp, "test-env", false, check, 5*time.Second, 1*time.Second)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %s", resp.Diagnostics.Errors())
	}
}

func TestWaitForDeletion_NoMfaNoPendingDelete_ReturnsDeleted(t *testing.T) {
	t.Parallel()
	origInterval := DeletePollInterval
	DeletePollInterval = 50 * time.Millisecond
	defer func() { DeletePollInterval = origInterval }()

	resp := &resource.DeleteResponse{}
	check := func(ctx context.Context, name string) (bool, error) {
		return false, nil // pendingDelete=false, no error (not 404)
	}

	start := time.Now()
	WaitForDeletion(context.Background(), resp, "test-env", false, check, 5*time.Second, 1*time.Second)
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
	origInterval := DeletePollInterval
	DeletePollInterval = 50 * time.Millisecond
	defer func() { DeletePollInterval = origInterval }()

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

	WaitForDeletion(context.Background(), resp, "test-env", true, check, 5*time.Second, 5*time.Second)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %s", resp.Diagnostics.Errors())
	}
}

func TestWaitForDeletion_MfaTimeout(t *testing.T) {
	t.Parallel()
	origInterval := DeletePollInterval
	DeletePollInterval = 50 * time.Millisecond
	defer func() { DeletePollInterval = origInterval }()

	resp := &resource.DeleteResponse{}
	check := func(ctx context.Context, name string) (bool, error) {
		return false, nil // pendingDelete stays false forever
	}

	WaitForDeletion(context.Background(), resp, "test-env", true, check, 5*time.Second, 200*time.Millisecond)

	if !resp.Diagnostics.HasError() {
		t.Error("expected MFA timeout error, got none")
	}
}

func TestWaitForDeletion_NonNotFoundError(t *testing.T) {
	t.Parallel()
	origInterval := DeletePollInterval
	DeletePollInterval = 50 * time.Millisecond
	defer func() { DeletePollInterval = origInterval }()

	resp := &resource.DeleteResponse{}
	check := func(ctx context.Context, name string) (bool, error) {
		return false, fmt.Errorf("connection refused")
	}

	WaitForDeletion(context.Background(), resp, "test-env", false, check, 5*time.Second, 1*time.Second)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error, got none")
	}
}

func TestWaitForDeletion_TransientErrorThenDeleted(t *testing.T) {
	t.Parallel()
	origInterval := DeletePollInterval
	DeletePollInterval = 50 * time.Millisecond
	defer func() { DeletePollInterval = origInterval }()

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

	WaitForDeletion(context.Background(), resp, "test-env", false, check, 5*time.Second, 1*time.Second)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %s", resp.Diagnostics.Errors())
	}
	if calls.Load() < 3 {
		t.Errorf("expected status check to run until not-found after 503, got %d calls", calls.Load())
	}
}
