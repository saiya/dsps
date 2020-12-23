package middleware

import (
	"context"
	"fmt"

	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/http/router"
	"github.com/saiya/dsps/server/http/utils"
)

// AdminAuthDependency is to inject required objects to the middleware
type AdminAuthDependency interface {
	RealIPDependency
	GetAdminAuthConfig() *config.AdminAuthConfig
}

// NewAdminAuth creates middleware for authentication
func NewAdminAuth(mainCtx context.Context, deps AdminAuthDependency) router.MiddlewareFunc {
	allowedIP := func(args router.MiddlewareArgs) bool {
		clientIP := GetRealIP(deps, args.R)
		for _, allowed := range deps.GetAdminAuthConfig().Networks {
			if allowed.Contains(clientIP) {
				return true
			}
		}
		return false
	}

	allowedToken := func(ctx context.Context, args router.MiddlewareArgs) bool {
		token := utils.GetBearerToken(ctx, args)
		for _, allowed := range deps.GetAdminAuthConfig().BearerTokens {
			if allowed == token {
				return true
			}
		}
		return false
	}

	return router.AsMiddlewareFunc(func(ctx context.Context, args router.MiddlewareArgs, next func(context.Context, router.MiddlewareArgs)) {
		if !allowedIP(args) {
			utils.SendError(ctx, args.W, 403, ``, fmt.Errorf(`IP address is not in allow list: %s`, GetRealIP(deps, args.R)))
			return
		}
		if !allowedToken(ctx, args) {
			utils.SendError(ctx, args.W, 403, ``, fmt.Errorf(`Invalid authentication token or no token`))
			return
		}
		next(ctx, args)
	})
}
