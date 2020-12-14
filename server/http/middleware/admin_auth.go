package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/saiya/dsps/server/domain"
)

// AdminAuthDependency is to inject required objects to the middleware
type AdminAuthDependency interface {
	GetStorage() domain.Storage
	DiscloseAuthRejectionDetail() bool
}

// NewAdminAuth creates middleware for authentication
func NewAdminAuth(mainCtx context.Context, deps AdminAuthDependency) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: Implement API key + IP address check
		ctx.Next()
	}
}
