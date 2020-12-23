package middleware

import (
	"context"
	"net"
	"net/http"

	"github.com/natureglobal/realip"

	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/http/router"
)

// RealIPDependency is to inject required objects to the middleware
type RealIPDependency interface {
	// If empty, RealIPMiddleware does nothing.
	GetIPHeaderName() string
	GetTrustedProxyRanges() []domain.CIDR
}

// GetRealIP returns end-user IP address information if available.
// If not available, returns ""
func GetRealIP(deps RealIPDependency, r router.Request) string {
	if deps.GetIPHeaderName() == "" {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			return r.RemoteAddr
		}
		return ip
	}
	return r.Header.Get(deps.GetIPHeaderName()) // RealIPMiddleware overwritten this
}

// RealIPMiddleware initialize middleware for real IP handling.
// Because "github.com/natureglobal/realip" is http.Handler middleware, this method wraps http.Handler
func RealIPMiddleware(deps RealIPDependency) router.MiddlewareFunc {
	return func(method, path string) router.Middleware {
		if deps.GetIPHeaderName() == "" {
			return func(ctx context.Context, args router.MiddlewareArgs, next func(context.Context, router.MiddlewareArgs)) {
				next(ctx, args)
			}
		}

		realIPFrom := make([]*net.IPNet, 0, len(deps.GetTrustedProxyRanges()))
		for _, cidr := range deps.GetTrustedProxyRanges() {
			realIPFrom = append(realIPFrom, cidr.IPNet())
		}
		m := realip.MustMiddleware(&realip.Config{
			RealIPHeader:    deps.GetIPHeaderName(),
			SetHeader:       deps.GetIPHeaderName(),
			RealIPFrom:      realIPFrom,
			RealIPRecursive: true,
		})
		return func(ctx context.Context, args router.MiddlewareArgs, next func(context.Context, router.MiddlewareArgs)) {
			m(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				args.R = router.Request{Request: r}
				args.W = router.NewResponseWriter(rw)
				next(ctx, args)
			})).ServeHTTP(args.W, args.R.Request)
		}
	}

}
