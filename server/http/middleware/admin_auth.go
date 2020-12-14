package middleware

import (
	"context"

	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/http/router"
)

// AdminAuthDependency is to inject required objects to the middleware
type AdminAuthDependency interface {
	GetStorage() domain.Storage
	DiscloseAuthRejectionDetail() bool
}

// NewAdminAuth creates middleware for authentication
func NewAdminAuth(mainCtx context.Context, deps AdminAuthDependency) router.Middleware {
	return func(ctx context.Context, args router.MiddlewareArgs, next func(context.Context)) {
		// TODO: Implement API key + IP address check
		next(ctx)
	}
}
