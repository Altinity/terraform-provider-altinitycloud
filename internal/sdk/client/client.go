package client

import (
	"context"
	"fmt"
	"net/http"

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
