package client

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/Yamashou/gqlgenc/clientv2"
)

func TestWithRetry_Success(t *testing.T) {
	interceptor := WithRetry(3, time.Millisecond)
	calls := 0
	next := func(ctx context.Context, req *http.Request, gqlInfo *clientv2.GQLRequestInfo, res interface{}) error {
		calls++
		return nil
	}

	err := interceptor(context.Background(), &http.Request{}, &clientv2.GQLRequestInfo{}, nil, next)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestWithRetry_RetryableSucceedsOnThird(t *testing.T) {
	interceptor := WithRetry(3, time.Millisecond)
	calls := 0
	next := func(ctx context.Context, req *http.Request, gqlInfo *clientv2.GQLRequestInfo, res interface{}) error {
		calls++
		if calls < 3 {
			return errors.New("status 503")
		}
		return nil
	}

	err := interceptor(context.Background(), &http.Request{}, &clientv2.GQLRequestInfo{}, nil, next)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestWithRetry_NonRetryableError(t *testing.T) {
	interceptor := WithRetry(3, time.Millisecond)
	calls := 0
	next := func(ctx context.Context, req *http.Request, gqlInfo *clientv2.GQLRequestInfo, res interface{}) error {
		calls++
		return errors.New("status 400 bad request")
	}

	err := interceptor(context.Background(), &http.Request{}, &clientv2.GQLRequestInfo{}, nil, next)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if calls != 1 {
		t.Errorf("expected 1 call (no retry for 400), got %d", calls)
	}
}

func TestWithRetry_ExhaustsRetries(t *testing.T) {
	interceptor := WithRetry(2, time.Millisecond)
	calls := 0
	next := func(ctx context.Context, req *http.Request, gqlInfo *clientv2.GQLRequestInfo, res interface{}) error {
		calls++
		return errors.New("connection refused")
	}

	err := interceptor(context.Background(), &http.Request{}, &clientv2.GQLRequestInfo{}, nil, next)
	if err == nil {
		t.Fatal("expected error after retries exhausted")
	}
	// 1 initial + 2 retries = 3
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestWithRetry_RespectsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	interceptor := WithRetry(3, 100*time.Millisecond)
	calls := 0
	next := func(ctx context.Context, req *http.Request, gqlInfo *clientv2.GQLRequestInfo, res interface{}) error {
		calls++
		cancel()
		return errors.New("status 503")
	}

	err := interceptor(ctx, &http.Request{}, &clientv2.GQLRequestInfo{}, nil, next)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got: %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call before context cancel, got %d", calls)
	}
}

func TestWithRetry_429IsRetryable(t *testing.T) {
	interceptor := WithRetry(1, time.Millisecond)
	calls := 0
	next := func(ctx context.Context, req *http.Request, gqlInfo *clientv2.GQLRequestInfo, res interface{}) error {
		calls++
		if calls == 1 {
			return errors.New("status 429 too many requests")
		}
		return nil
	}

	err := interceptor(context.Background(), &http.Request{}, &clientv2.GQLRequestInfo{}, nil, next)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if calls != 2 {
		t.Errorf("expected 2 calls, got %d", calls)
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		err  string
		want bool
	}{
		{"status 429", true},
		{"status 502", true},
		{"status 503", true},
		{"status 504", true},
		{"connection reset", true},
		{"connection refused", true},
		{"i/o timeout", true},
		{"TLS handshake timeout", true},
		{"EOF", true},
		{"status 400", false},
		{"status 401", false},
		{"status 403", false},
		{"status 404", false},
		{"some random error", false},
	}

	for _, tt := range tests {
		got := isRetryable(errors.New(tt.err))
		if got != tt.want {
			t.Errorf("isRetryable(%q) = %v, want %v", tt.err, got, tt.want)
		}
	}
}
