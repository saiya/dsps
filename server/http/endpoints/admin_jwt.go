package endpoints

import (
	"github.com/gin-gonic/gin"

	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/http/utils"
)

// AdminJwtEndpointDependency is to inject required objects to the endpoint
type AdminJwtEndpointDependency interface {
	GetStorage() domain.Storage
}

// InitAdminJwtEndpoints registers endpoints
func InitAdminJwtEndpoints(adminRouter gin.IRoutes, deps AdminJwtEndpointDependency) {
	storage := deps.GetStorage().AsJwtStorage()
	adminRouter.PUT("/jwt/revoke", func(ctx *gin.Context) {
		if storage == nil {
			utils.SendJwtUnsupportedError(ctx)
			return
		}

		jti := ctx.Query("jti")
		if jti == "" {
			utils.SendMissingParameter(ctx, "jti")
			return
		}

		exp, err := domain.ParseJwtExp(ctx.Query("exp"))
		if err != nil {
			utils.SendInvalidParameter(ctx, "exp", err)
			return
		}

		if err := storage.RevokeJwt(ctx, exp, domain.JwtJti(jti)); err != nil {
			utils.SentInternalServerError(ctx, err)
			return
		}

		ctx.JSON(200, gin.H{
			"jti": jti,
			"exp": exp.Int64(),
		})
	})
}
