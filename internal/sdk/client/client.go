package client

import (
	"bytes"
	"context"
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

// retryablePatterns contains error substrings that indicate a transient failure.
var retryablePatterns = []string{
	"429",
	"502",
	"503",
	"504",
	"connection reset",
	"connection refused",
	"connect: connection refused",
	"i/o timeout",
	"TLS handshake timeout",
	"EOF",
}

func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	for _, pattern := range retryablePatterns {
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
	if gqlInfo == nil || gqlInfo.Request == nil {
		return false
	}
	return strings.HasPrefix(strings.TrimSpace(gqlInfo.Request.Query), "mutation")
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
