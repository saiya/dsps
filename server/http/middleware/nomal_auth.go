package middleware

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/http/utils"
	"github.com/saiya/dsps/server/logger"
)

// NormalAuthDependency is to inject required objects to the middleware
type NormalAuthDependency interface {
	DiscloseAuthRejectionDetail() bool
}

// NewNormalAuth creates middleware for authentication
func NewNormalAuth(mainCtx context.Context, deps NormalAuthDependency, channelOf func(*gin.Context) (domain.Channel, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		channel, err := channelOf(ctx)
		if err != nil {
			utils.SendInvalidParameter(ctx, "channelID", err)
			return
		}

		bearerToken := utils.GetBearerToken(ctx)
		if err := channel.ValidateJwt(ctx, bearerToken); err != nil {
			logger.Of(ctx).Infof(logger.CatAuth, `JWT verification failure: %v`, err)

			body := gin.H{
				"code":  ErrAuthRejection.Code(),
				"error": "Unauthorized",
			}
			if deps.DiscloseAuthRejectionDetail() {
				body["reason"] = fmt.Sprintf("JWT verification failure: %v", err)
			}
			ctx.AbortWithStatusJSON(403, body)
			return
		}

		ctx.Next()
	}
}
