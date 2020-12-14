package endpoints

import (
	"context"

	"github.com/saiya/dsps/server/http/router"
	"github.com/saiya/dsps/server/http/utils"
	"github.com/saiya/dsps/server/logger"
)

// AdminLoggingEndpointDependency is to inject required objects to the endpoint
type AdminLoggingEndpointDependency interface {
	GetLogFilter() *logger.Filter
}

// InitAdminLoggingEndpoints registers endpoints
func InitAdminLoggingEndpoints(adminRouter *router.Router, deps AdminLoggingEndpointDependency) {
	adminRouter.PUT("/log/level", func(ctx context.Context, args router.HandlerArgs) {
		category := args.R.GetQueryParam("category")
		if category == "" {
			utils.SendMissingParameter(ctx, args.W, "category")
			return
		}

		level, err := logger.ParseLevel(args.R.GetQueryParam("level"))
		if err != nil {
			utils.SendInvalidParameter(ctx, args.W, "level", err)
			return
		}

		logger.Of(ctx).Infof(logger.CatLogger, `set logging threshold of "%s" category to %s`, category, level)
		deps.GetLogFilter().SetThreshold(logger.ParseCategory(category), level)

		utils.SendNoContent(ctx, args.W)
	})
}
