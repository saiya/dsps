package middleware

import (
	"context"

	"github.com/saiya/dsps/server/http/router"
	"github.com/saiya/dsps/server/telemetry"
)

// TracingDependency is to inject required objects to the middleware
type TracingDependency interface {
	GetTelemetry() *telemetry.Telemetry
}

// TracingMiddleware traces incoming HTTP request/response
func TracingMiddleware(realIPDeps RealIPDependency, deps TracingDependency) router.MiddlewareFunc {
	telemetry := deps.GetTelemetry()
	return func(method, path string) router.Middleware {
		return func(ctx context.Context, args router.MiddlewareArgs, next func(context.Context, router.MiddlewareArgs)) {
			ctx, end := telemetry.StartHTTPSpan(ctx, true, args.R.Request)
			defer end()
			telemetry.SetHTTPServerAttributes(ctx, args.R.Request, path, GetRealIP(realIPDeps, args.R))
			next(ctx, args)
			telemetry.SetHTTPResponseAttributes(ctx, args.W.Written().StatusCode, int64(args.W.Written().BodyBytes))
		}
	}
}
