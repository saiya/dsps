package http

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/http/endpoints"
	"github.com/saiya/dsps/server/http/middleware"
	"github.com/saiya/dsps/server/logger"
)

// InitEndpoints registers endpoints of the DSPS server
func InitEndpoints(mainCtx context.Context, router gin.IRouter, deps *ServerDependencies) {
	endpoints.InitProbeEndpoints(router, deps)

	adminRouter := router.Group("/admin")
	endpoints.InitAdminJwtEndpoints(adminRouter, deps)
	endpoints.InitAdminLoggingEndpoints(adminRouter, deps)

	channelRouter := router.Group(
		"/channel/:channelID",
		func(ctx *gin.Context) {
			logger.ModifyGinContext(ctx).WithStr("channelID", ctx.Param("channelID")).Build()
			ctx.Next()
		},
		middleware.NewNormalAuth(mainCtx, deps, func(ctx *gin.Context) (domain.Channel, error) {
			id, err := domain.ParseChannelID(ctx.Param("channelID"))
			if err != nil {
				return nil, err
			}
			return deps.ChannelProvider(id)
		}),
	)
	endpoints.InitPublishEndpoints(channelRouter, deps)
	endpoints.InitSubscriptionPollingEndpoints(channelRouter, deps)
}
