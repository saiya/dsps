package endpoints

import (
	"context"

	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/http/router"
	"github.com/saiya/dsps/server/http/utils"
)

// AdminJwtEndpointDependency is to inject required objects to the endpoint
type AdminJwtEndpointDependency interface {
	GetStorage() domain.Storage
}

// InitAdminJwtEndpoints registers endpoints
func InitAdminJwtEndpoints(adminRouter *router.Router, deps AdminJwtEndpointDependency) {
	storage := deps.GetStorage().AsJwtStorage()
	adminRouter.PUT("/jwt/revoke", func(ctx context.Context, args router.HandlerArgs) {
		if storage == nil {
			utils.SendJwtUnsupportedError(ctx, args.W)
			return
		}

		jti := args.R.GetQueryParam("jti")
		if jti == "" {
			utils.SendMissingParameter(ctx, args.W, "jti")
			return
		}

		exp, err := domain.ParseJwtExp(args.R.GetQueryParam("exp"))
		if err != nil {
			utils.SendInvalidParameter(ctx, args.W, "exp", err)
			return
		}

		if err := storage.RevokeJwt(ctx, exp, domain.JwtJti(jti)); err != nil {
			utils.SendInternalServerError(ctx, args.W, err)
			return
		}

		utils.SendJSON(ctx, args.W, 200, map[string]interface{}{
			"jti": jti,
			"exp": exp.Int64(),
		})
	})
}
