package common

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// shortPolls keeps multi-poll cases fast. The interval is a package var, so tests
// touching it cannot run in parallel.
func shortPolls(t *testing.T) {
	t.Helper()

	previous := MATCH_SPEC_POLL_INTERVAL
	MATCH_SPEC_POLL_INTERVAL = time.Millisecond
	t.Cleanup(func() { MATCH_SPEC_POLL_INTERVAL = previous })
}

// scriptedRefresh returns a refresh func that walks the given results, repeating
// the last one, and a pointer to the call count.
func scriptedRefresh(results ...*PollResult) (StatusRefreshFunc, *int) {
	calls := 0
	return func(context.Context) (*PollResult, error) {
		calls++
		if calls > len(results) {
			return results[len(results)-1], nil
		}
		return results[calls-1], nil
	}, &calls
}

func ready(revision int64) *PollResult {
	return &PollResult{AppliedSpecRevision: revision, Found: true}
}

func withErrors(revision int64, errs ...EnvError) *PollResult {
	return &PollResult{AppliedSpecRevision: revision, Errors: errs, Found: true}
}

func readEnvStatus(t *testing.T, target int64, refresh StatusRefreshFunc) (bool, diag.Diagnostics) {
	t.Helper()

	var diags diag.Diagnostics
	ok := ReadEnvStatus(context.Background(), "env", target, false, time.Minute, refresh, &diags)
	return ok, diags
}

func TestReadEnvStatusNoTargetFetchesOnce(t *testing.T) {
	refresh, calls := scriptedRefresh(ready(7))

	ok, diags := readEnvStatus(t, 0, refresh)

	if !ok || diags.HasError() {
		t.Fatalf("expected success, got ok=%v diags=%v", ok, diags.Errors())
	}
	if *calls != 1 {
		t.Errorf("expected 1 refresh, got %d", *calls)
	}
}

func TestReadEnvStatusTargetAlreadyMetFetchesOnce(t *testing.T) {
	refresh, calls := scriptedRefresh(ready(5))

	ok, diags := readEnvStatus(t, 5, refresh)

	if !ok || diags.HasError() {
		t.Fatalf("expected success, got ok=%v diags=%v", ok, diags.Errors())
	}
	if *calls != 1 {
		t.Errorf("expected 1 refresh, got %d", *calls)
	}
}

func TestReadEnvStatusRefreshError(t *testing.T) {
	var diags diag.Diagnostics
	refresh := func(context.Context) (*PollResult, error) {
		return nil, errors.New("boom")
	}

	if ReadEnvStatus(context.Background(), "env", 0, false, time.Minute, refresh, &diags) {
		t.Fatal("expected failure")
	}
	if !strings.Contains(diags.Errors()[0].Detail(), "Unable to read env status env") {
		t.Errorf("unexpected detail: %s", diags.Errors()[0].Detail())
	}
}

func TestReadEnvStatusNotFound(t *testing.T) {
	refresh, _ := scriptedRefresh(&PollResult{})

	ok, diags := readEnvStatus(t, 0, refresh)

	if ok {
		t.Fatal("expected failure")
	}
	if !strings.Contains(diags.Errors()[0].Detail(), "Environment env was not found") {
		t.Errorf("unexpected detail: %s", diags.Errors()[0].Detail())
	}
}

// The wait leaves the model holding whatever the last poll wrote, so the read
// refreshes once more before the caller sets state.
func TestReadEnvStatusRefreshesAfterWaiting(t *testing.T) {
	refresh, calls := scriptedRefresh(ready(1), ready(2))

	ok, diags := readEnvStatus(t, 2, refresh)

	if !ok || diags.HasError() {
		t.Fatalf("expected success, got ok=%v diags=%v", ok, diags.Errors())
	}
	if *calls != 3 {
		t.Errorf("expected 3 refreshes (initial, poll, settle), got %d", *calls)
	}
}

func TestReadEnvStatusWaitTimesOut(t *testing.T) {
	shortPolls(t)

	refresh, _ := scriptedRefresh(ready(1))

	var diags diag.Diagnostics
	ok := ReadEnvStatus(context.Background(), "env", 2, false, 20*time.Millisecond, refresh, &diags)

	if ok {
		t.Fatal("expected failure")
	}
	if !strings.Contains(diags.Errors()[0].Summary(), "Status Error") {
		t.Errorf("unexpected summary: %s", diags.Errors()[0].Summary())
	}
}

// An env that never provisioned reports DISCONNECTED until it first connects.
func TestReadEnvStatusDisconnectedBeforeFirstProvisionKeepsWaiting(t *testing.T) {
	shortPolls(t)

	refresh, calls := scriptedRefresh(
		ready(0),
		withErrors(0, EnvError{Code: "DISCONNECTED", Message: "not connected"}),
		ready(2),
	)

	ok, diags := readEnvStatus(t, 2, refresh)

	if !ok || diags.HasError() {
		t.Fatalf("expected success, got ok=%v diags=%v", ok, diags.Errors())
	}
	if *calls < 3 {
		t.Errorf("expected the disconnected poll to be retried, got %d refreshes", *calls)
	}
}

func TestReadEnvStatusDisconnectedAfterProvisionFails(t *testing.T) {
	shortPolls(t)

	refresh, _ := scriptedRefresh(
		ready(1),
		withErrors(1, EnvError{Code: "DISCONNECTED", Message: "lost the tunnel"}),
	)

	ok, diags := readEnvStatus(t, 2, refresh)

	if ok {
		t.Fatal("expected failure")
	}
	if !strings.Contains(diags.Errors()[0].Detail(), "lost the tunnel") {
		t.Errorf("unexpected detail: %s", diags.Errors()[0].Detail())
	}
}

func TestReadEnvStatusK8SDisconnectedBeforeFirstProvisionKeepsWaiting(t *testing.T) {
	shortPolls(t)

	refresh, calls := scriptedRefresh(
		ready(0),
		withErrors(0, EnvError{Code: "K8S_DISCONNECTED", Message: "not connected"}),
		ready(2),
	)

	ok, diags := readEnvStatus(t, 2, refresh)

	if !ok || diags.HasError() {
		t.Fatalf("expected success, got ok=%v diags=%v", ok, diags.Errors())
	}
	if *calls < 3 {
		t.Errorf("expected the disconnected poll to be retried, got %d refreshes", *calls)
	}
}

func TestReadEnvStatusProvisioningErrorFails(t *testing.T) {
	shortPolls(t)

	refresh, _ := scriptedRefresh(
		ready(1),
		withErrors(0, EnvError{Code: "QUOTA_EXCEEDED", Message: "no capacity"}),
	)

	ok, diags := readEnvStatus(t, 2, refresh)

	if ok {
		t.Fatal("expected failure")
	}
	if !strings.Contains(diags.Errors()[0].Detail(), "QUOTA_EXCEEDED: no capacity") {
		t.Errorf("unexpected detail: %s", diags.Errors()[0].Detail())
	}
}

// A disconnected env that has provisioned before must not be masked by a
// non-blocking sibling error in the same response.
func TestReadEnvStatusMixedErrorsBlock(t *testing.T) {
	shortPolls(t)

	refresh, _ := scriptedRefresh(
		ready(1),
		withErrors(0,
			EnvError{Code: "DISCONNECTED", Message: "not connected"},
			EnvError{Code: "QUOTA_EXCEEDED", Message: "no capacity"},
		),
	)

	ok, diags := readEnvStatus(t, 2, refresh)

	if ok {
		t.Fatal("expected failure")
	}
	detail := diags.Errors()[0].Detail()
	if !strings.Contains(detail, "QUOTA_EXCEEDED") {
		t.Errorf("unexpected detail: %s", detail)
	}
	if strings.Contains(detail, "DISCONNECTED") {
		t.Errorf("the non-blocking disconnect should not be reported: %s", detail)
	}
}
