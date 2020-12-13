package middleware

import (
	"context"
	"errors"
	"fmt"

	"github.com/saiya/dsps/server/jwt"

	"github.com/gin-gonic/gin"
	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/http/utils"
	"github.com/saiya/dsps/server/logger"
)

// NormalAuthDependency is to inject required objects to the middleware
type NormalAuthDependency interface {
	GetStorage() domain.Storage
	DiscloseAuthRejectionDetail() bool
}

// NewNormalAuth creates middleware for authentication
func NewNormalAuth(mainCtx context.Context, deps NormalAuthDependency, channelOf func(*gin.Context) (domain.Channel, error)) gin.HandlerFunc {
	jwtStorage := deps.GetStorage().AsJwtStorage()
	return func(ctx *gin.Context) {
		channel, err := channelOf(ctx)
		if err != nil {
			utils.SendInvalidParameter(ctx, "channelID", err)
			return
		}

		bearerToken := utils.GetBearerToken(ctx)
		authErr := channel.ValidateJwt(ctx, bearerToken)
		if authErr == nil && jwtStorage != nil {
			// If bearerToken is not JWT, channel.ValidateJwt() rejects it if JWT validation configured.
			// If JWT validation not configured, it is okay to pass non-JWT or empty bearerToken.
			jti, jwtParseError := jwt.ExtractJti(bearerToken)
			if authErr == nil && jwtParseError == nil && jti != nil {
				var revoked bool
				revoked, authErr = jwtStorage.IsRevokedJwt(ctx, *jti)
				if authErr == nil && revoked {
					authErr = errors.New(`presented JWT has been revoked`)
				}
			}
		}
		if authErr != nil {
			logger.Of(ctx).Infof(logger.CatAuth, `JWT verification failure: %v`, authErr)

			body := gin.H{
				"code":  ErrAuthRejection.Code(),
				"error": "Unauthorized",
			}
			if deps.DiscloseAuthRejectionDetail() {
				body["reason"] = fmt.Sprintf("JWT verification failure: %v", authErr)
			}
			ctx.AbortWithStatusJSON(403, body)
			return
		}

		ctx.Next()
	}
}
