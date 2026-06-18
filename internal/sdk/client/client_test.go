package client

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/Yamashou/gqlgenc/clientv2"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// netErr builds the *clientv2.ErrorResponse the client returns for a non-2xx
// HTTP response (a transient, server-side failure).
func netErr(code int) error {
	return &clientv2.ErrorResponse{NetworkError: &clientv2.HTTPError{Code: code, Message: "boom"}}
}

// gqlErr builds the *clientv2.ErrorResponse the client returns for a 200 OK
// response that carries GraphQL business errors (no network error).
func gqlErr(message string) error {
	return &clientv2.ErrorResponse{GqlErrors: &gqlerror.List{{Message: message}}}
}

// queryInfo is a retryable (non-mutation) operation for exercising WithRetry.
func queryInfo() *clientv2.GQLRequestInfo {
	return &clientv2.GQLRequestInfo{Request: &clientv2.Request{Query: "query GetAWSEnv { awsEnv }"}}
}

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
			return netErr(503)
		}
		return nil
	}

	err := interceptor(context.Background(), &http.Request{}, queryInfo(), nil, next)
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
		return netErr(400)
	}

	err := interceptor(context.Background(), &http.Request{}, queryInfo(), nil, next)
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

	err := interceptor(context.Background(), &http.Request{}, queryInfo(), nil, next)
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
		return netErr(503)
	}

	err := interceptor(ctx, &http.Request{}, queryInfo(), nil, next)
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
			return netErr(429)
		}
		return nil
	}

	err := interceptor(context.Background(), &http.Request{}, queryInfo(), nil, next)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if calls != 2 {
		t.Errorf("expected 2 calls, got %d", calls)
	}
}

func TestWithRetry_ReplaysBodyOnRetry(t *testing.T) {
	interceptor := WithRetry(3, time.Millisecond)
	const payload = "graphql-payload"

	req, err := http.NewRequest(http.MethodPost, "http://example.test", bytes.NewReader([]byte(payload)))
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}

	calls := 0
	var reads []string
	next := func(ctx context.Context, req *http.Request, gqlInfo *clientv2.GQLRequestInfo, res interface{}) error {
		calls++
		body, _ := io.ReadAll(req.Body)
		reads = append(reads, string(body))
		if calls < 3 {
			return netErr(503)
		}
		return nil
	}

	if err := interceptor(context.Background(), req, queryInfo(), nil, next); err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
	for i, got := range reads {
		if got != payload {
			t.Errorf("attempt %d read body %q, want %q (body not replayed on retry)", i+1, got, payload)
		}
	}
}

func TestWithRetry_DoesNotRetryMutations(t *testing.T) {
	interceptor := WithRetry(3, time.Millisecond)
	calls := 0
	next := func(ctx context.Context, req *http.Request, gqlInfo *clientv2.GQLRequestInfo, res interface{}) error {
		calls++
		return netErr(503)
	}

	gqlInfo := &clientv2.GQLRequestInfo{Request: &clientv2.Request{Query: "mutation CreateAWSEnv ($input: CreateAWSEnvInput!) { createAWSEnv }"}}
	err := interceptor(context.Background(), &http.Request{}, gqlInfo, nil, next)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if calls != 1 {
		t.Errorf("expected 1 call (mutations are not retried), got %d", calls)
	}
}

func TestWithRetry_DoesNotRetryCommentedMutation(t *testing.T) {
	interceptor := WithRetry(3, time.Millisecond)
	calls := 0
	next := func(ctx context.Context, req *http.Request, gqlInfo *clientv2.GQLRequestInfo, res interface{}) error {
		calls++
		return netErr(503)
	}

	gqlInfo := &clientv2.GQLRequestInfo{Request: &clientv2.Request{Query: "# leading comment\n\nmutation CreateAWSEnv { createAWSEnv }"}}
	err := interceptor(context.Background(), &http.Request{}, gqlInfo, nil, next)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if calls != 1 {
		t.Errorf("expected 1 call (commented mutation must not be retried), got %d", calls)
	}
}

func TestWithRetry_RetriesQueries(t *testing.T) {
	interceptor := WithRetry(3, time.Millisecond)
	calls := 0
	next := func(ctx context.Context, req *http.Request, gqlInfo *clientv2.GQLRequestInfo, res interface{}) error {
		calls++
		if calls < 2 {
			return netErr(503)
		}
		return nil
	}

	gqlInfo := &clientv2.GQLRequestInfo{Request: &clientv2.Request{Query: "query GetAWSEnv ($name: String!) { awsEnv }"}}
	if err := interceptor(context.Background(), &http.Request{}, gqlInfo, nil, next); err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if calls != 2 {
		t.Errorf("expected 2 calls (queries are retried), got %d", calls)
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"429 network", netErr(429), true},
		{"502 network", netErr(502), true},
		{"503 network", netErr(503), true},
		{"504 network", netErr(504), true},
		{"400 network", netErr(400), false},
		{"401 network", netErr(401), false},
		{"404 network", netErr(404), false},
		{"connection reset", errors.New("connection reset by peer"), true},
		{"connection refused", errors.New("connection refused"), true},
		{"i/o timeout", errors.New("dial tcp: i/o timeout"), true},
		{"TLS handshake timeout", errors.New("net/http: TLS handshake timeout"), true},
		{"EOF transport", errors.New("request failed: EOF"), true},
		{"random transport", errors.New("some random error"), false},
		// GraphQL business errors must never be retried, even when the message
		// text happens to contain a transient code or "EOF".
		{"gql error mentioning 503", gqlErr("validation failed: value 503 invalid"), false},
		{"gql error mentioning EOF", gqlErr("unexpected EOF in user input"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRetryable(tt.err); got != tt.want {
				t.Errorf("isRetryable(%q) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestIsMutation(t *testing.T) {
	tests := []struct {
		name    string
		gqlInfo *clientv2.GQLRequestInfo
		want    bool
	}{
		{"plain mutation", &clientv2.GQLRequestInfo{Request: &clientv2.Request{Query: "mutation Foo { foo }"}}, true},
		{"leading whitespace mutation", &clientv2.GQLRequestInfo{Request: &clientv2.Request{Query: "\n\t  mutation Foo { foo }"}}, true},
		{"commented mutation", &clientv2.GQLRequestInfo{Request: &clientv2.Request{Query: "# c1\n# c2\nmutation Foo { foo }"}}, true},
		{"query", &clientv2.GQLRequestInfo{Request: &clientv2.Request{Query: "query Foo { foo }"}}, false},
		{"commented query", &clientv2.GQLRequestInfo{Request: &clientv2.Request{Query: "# note\nquery Foo { foo }"}}, false},
		{"nil info fails closed", nil, true},
		{"nil request fails closed", &clientv2.GQLRequestInfo{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isMutation(tt.gqlInfo); got != tt.want {
				t.Errorf("isMutation() = %v, want %v", got, tt.want)
			}
		})
	}
}
