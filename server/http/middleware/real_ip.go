package middleware

import (
	"context"
	"net"
	"strings"

	"github.com/natureglobal/realip"

	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/http/router"
	"github.com/saiya/dsps/server/sentry"
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
		return getRemoteAddrHostname(r)
	}
	return r.Header.Get(deps.GetIPHeaderName()) // RealIPMiddleware overwritten this
}

func getRemoteAddrHostname(r router.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// RealIPMiddleware initialize middleware for real IP handling.
// Because "github.com/natureglobal/realip" is http.Handler middleware, this method wraps http.Handler
func RealIPMiddleware(deps RealIPDependency) router.MiddlewareFunc {
	ipHeaderName := deps.GetIPHeaderName()
	sentryTagName := strings.ToLower(strings.ReplaceAll(ipHeaderName, "-", "_"))
	return func(method, path string) router.Middleware {
		if ipHeaderName == "" {
			return func(ctx context.Context, args router.MiddlewareArgs, next func(context.Context, router.MiddlewareArgs)) {
				remoteAddr := getRemoteAddrHostname(args.R)
				sentry.SetIPAddress(ctx, remoteAddr)
				sentry.AddTag(ctx, "remote_addr", remoteAddr)
				next(ctx, args)
			}
		}

		realIPFrom := make([]*net.IPNet, 0, len(deps.GetTrustedProxyRanges()))
		for _, cidr := range deps.GetTrustedProxyRanges() {
			realIPFrom = append(realIPFrom, cidr.IPNet())
		}
		m := router.WrapMiddleware(realip.MustMiddleware(&realip.Config{
			RealIPHeader:    ipHeaderName,
			SetHeader:       ipHeaderName,
			RealIPFrom:      realIPFrom,
			RealIPRecursive: true,
		}))
		return func(ctx context.Context, args router.MiddlewareArgs, next func(context.Context, router.MiddlewareArgs)) {
			m(ctx, args, func(ctx context.Context, args router.MiddlewareArgs) {
				remoteAddr := args.R.Header.Get(ipHeaderName)
				sentry.SetIPAddress(ctx, remoteAddr)
				sentry.AddTag(ctx, "remote_addr", remoteAddr)
				sentry.AddTag(ctx, sentryTagName, remoteAddr)
				next(ctx, args)
			})
		}
	}
}
