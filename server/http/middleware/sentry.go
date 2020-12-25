package middleware

import (
	"context"

	sentryhttp "github.com/getsentry/sentry-go/http"

	"github.com/saiya/dsps/server/http/router"
	"github.com/saiya/dsps/server/sentry"
)

// SentryDependency is to inject required objects to the middleware
type SentryDependency interface {
	GetSentry() sentry.Sentry
}

// SentryMiddleware traces incoming HTTP request/response
func SentryMiddleware(deps SentryDependency) router.MiddlewareFunc {
	sentry := deps.GetSentry()
	m := router.WrapMiddleware(sentryhttp.New(sentryhttp.Options{
		Repanic:         true,
		WaitForDelivery: false,
	}).Handle)
	return router.AsMiddlewareFunc(func(ctx context.Context, args router.MiddlewareArgs, next func(context.Context, router.MiddlewareArgs)) {
		ctx = sentry.WrapContext(ctx)
		args.R = router.Request{Request: args.R.WithContext(ctx)}
		m(ctx, args, next)
	})
}
