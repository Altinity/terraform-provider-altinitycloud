package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Yamashou/gqlgenc/clientv2"
)

func WithBearerAuthorization(ctx context.Context, token string) clientv2.RequestInterceptor {
	return func(ctx context.Context, req *http.Request, gqlInfo *clientv2.GQLRequestInfo, res interface{}, next clientv2.RequestInterceptorFunc) error {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		return next(ctx, req, gqlInfo, res)
	}
}

func WithUserAgent(ctx context.Context, userAgent string) clientv2.RequestInterceptor {
	return func(ctx context.Context, req *http.Request, gqlInfo *clientv2.GQLRequestInfo, res interface{}, next clientv2.RequestInterceptorFunc) error {
		req.Header.Set("User-Agent", userAgent)
		return next(ctx, req, gqlInfo, res)
	}
}

// retryableNetworkCodes are HTTP status codes that indicate a transient
// server-side failure worth retrying.
var retryableNetworkCodes = map[int]bool{
	429: true,
	502: true,
	503: true,
	504: true,
}

// transportPatterns match transport-level errors (no HTTP response received),
// which clientv2 surfaces as a plain wrapped error rather than an
// *ErrorResponse, so there is no status code to inspect.
var transportPatterns = []string{
	"connection reset",
	"connection refused",
	"i/o timeout",
	"TLS handshake timeout",
	"EOF",
}

func isRetryable(err error) bool {
	if err == nil {
		return false
	}

	// HTTP response received: only retry on transient status codes. GraphQL
	// business errors (NetworkError nil) must never be retried, even if their
	// message text happens to contain a code or "EOF".
	var errResp *clientv2.ErrorResponse
	if errors.As(err, &errResp) {
		return errResp.NetworkError != nil && retryableNetworkCodes[errResp.NetworkError.Code]
	}

	// No HTTP response: classify the transport error by its message.
	msg := err.Error()
	for _, pattern := range transportPatterns {
		if strings.Contains(msg, pattern) {
			return true
		}
	}
	return false
}

// isMutation reports whether the GraphQL operation mutates server state. Such
// operations are not idempotent: retrying one that already succeeded (e.g. when
// a transient error happens after the server committed it) yields spurious
// failures like "environment already exists".
func isMutation(gqlInfo *clientv2.GQLRequestInfo) bool {
	// Fail closed: if the operation can't be determined, treat it as a mutation
	// so it is never retried.
	if gqlInfo == nil || gqlInfo.Request == nil {
		return true
	}
	return strings.HasPrefix(operationStart(gqlInfo.Request.Query), "mutation")
}

// operationStart returns the query with leading whitespace and '#' comment
// lines stripped, so the operation keyword can be matched even when the query
// begins with comments.
func operationStart(query string) string {
	for {
		query = strings.TrimLeft(query, " \t\r\n")
		if !strings.HasPrefix(query, "#") {
			return query
		}
		i := strings.IndexByte(query, '\n')
		if i < 0 {
			return ""
		}
		query = query[i+1:]
	}
}

func WithRetry(maxRetries int, initialBackoff time.Duration) clientv2.RequestInterceptor {
	return func(ctx context.Context, req *http.Request, gqlInfo *clientv2.GQLRequestInfo, res interface{}, next clientv2.RequestInterceptorFunc) error {
		// Only idempotent operations (queries) are safe to retry.
		if isMutation(gqlInfo) {
			return next(ctx, req, gqlInfo, res)
		}

		// Buffer the body so it can be replayed: the transport consumes req.Body on
		// each attempt, so without a fresh body a retry sends an empty payload
		// ("ContentLength=N with Body length 0").
		if req.Body != nil && req.GetBody == nil {
			body, readErr := io.ReadAll(req.Body)
			_ = req.Body.Close()
			if readErr != nil {
				return readErr
			}
			req.GetBody = func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader(body)), nil
			}
		}

		var err error
		backoff := initialBackoff

		for attempt := 0; attempt <= maxRetries; attempt++ {
			if req.GetBody != nil {
				body, bodyErr := req.GetBody()
				if bodyErr != nil {
					return bodyErr
				}
				req.Body = body
			}

			err = next(ctx, req, gqlInfo, res)
			if err == nil {
				return nil
			}

			if !isRetryable(err) {
				return err
			}

			if attempt == maxRetries {
				break
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}

			backoff *= 2
		}

		return err
	}
}
