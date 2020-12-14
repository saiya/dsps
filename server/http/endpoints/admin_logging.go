package endpoints

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/saiya/dsps/server/http/utils"
	"github.com/saiya/dsps/server/logger"
)

// AdminLoggingEndpointDependency is to inject required objects to the endpoint
type AdminLoggingEndpointDependency interface {
	GetLogFilter() *logger.Filter
}

// InitAdminLoggingEndpoints registers endpoints
func InitAdminLoggingEndpoints(adminRouter gin.IRoutes, deps AdminLoggingEndpointDependency) {
	adminRouter.PUT("/log/level", func(ctx *gin.Context) {
		category := ctx.Query("category")
		if category == "" {
			utils.SendMissingParameter(ctx, "category")
			return
		}

		level, err := logger.ParseLevel(ctx.Query("level"))
		if err != nil {
			utils.SendInvalidParameter(ctx, "level", err)
			return
		}

		logger.Of(ctx).Infof(logger.CatLogger, `set logging threshold of "%s" category to %s`, category, level)
		deps.GetLogFilter().SetThreshold(logger.ParseCategory(category), level)

		ctx.Status(http.StatusNoContent)
	})
}
