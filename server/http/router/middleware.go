package router

import (
	"context"
	"net/http"
)

// Middleware is a filter that request passes
type Middleware func(ctx context.Context, args MiddlewareArgs, next func(context.Context, MiddlewareArgs))

// MiddlewareFunc is a function to dynamically create Middleware
type MiddlewareFunc func(method string, path string) Middleware

// AsMiddlewareFunc wraps Middleware as MiddlewareFunc
func AsMiddlewareFunc(m Middleware) MiddlewareFunc {
	return func(method, path string) Middleware { return m }
}

// WrapMiddleware wraps standard net/http middleware as Middleware
func WrapMiddleware(m func(http.Handler) http.Handler) Middleware {
	return func(ctx context.Context, args MiddlewareArgs, next func(context.Context, MiddlewareArgs)) {
		m(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			args.R = Request{Request: r}
			args.W = NewResponseWriter(rw)
			next(ctx, args)
		})).ServeHTTP(args.W, args.R.Request)
	}
}

// MiddlewareArgs holds arguments of middleware invoke
type MiddlewareArgs struct {
	HandlerArgs
}
