package middleware

import (
	"context"

	"github.com/saiya/dsps/server/http/router"
)

// DefaultHeadersDependency is to inject required objects to the middleware
type DefaultHeadersDependency interface {
	GetDefaultHeaders() map[string]string
}

// DefaultHeadersMiddleware is middleware to set some headers by default
func DefaultHeadersMiddleware(deps DefaultHeadersDependency) router.Middleware {
	return func(ctx context.Context, args router.MiddlewareArgs, next func(context.Context)) {
		for name, value := range deps.GetDefaultHeaders() {
			if value != "" {
				args.W.Header().Add(name, value)
			}
		}
		next(ctx)
	}
}
