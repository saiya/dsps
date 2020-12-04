package http

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/saiya/dsps/server/http/endpoints"
	"github.com/saiya/dsps/server/logger"
)

// InitEndpoints registers endpoints of the DSPS server
func InitEndpoints(mainCtx context.Context, router gin.IRouter, deps *ServerDependencies) {
	endpoints.InitProbeEndpoints(router, deps)

	channel := router.Group("/channel/:channelID")
	channel.Use(func(ctx *gin.Context) {
		logger.ModifyGinContext(ctx).WithStr("channelID", ctx.Param("channelID")).Build()
		ctx.Next()
	})
	endpoints.InitPublishEndpoints(channel, deps)
	endpoints.InitPollingEndpoints(channel, deps)
}
