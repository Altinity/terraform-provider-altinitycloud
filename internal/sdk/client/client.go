package client

import (
	"context"
	"fmt"
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

func WithRetry(maxRetries int, initialBackoff time.Duration) clientv2.RequestInterceptor {
	return func(ctx context.Context, req *http.Request, gqlInfo *clientv2.GQLRequestInfo, res interface{}, next clientv2.RequestInterceptorFunc) error {
		var err error
		backoff := initialBackoff

		for attempt := 0; attempt <= maxRetries; attempt++ {
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
