package http

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/saiya/dsps/server/http/endpoints"
)

// InitEndpoints registers endpoints of the DSPS server
func InitEndpoints(mainCtx context.Context, router gin.IRoutes, deps *ServerDependencies) {
	endpoints.InitProbeEndpoints(router, deps)
	endpoints.InitPublishEndpoints(router, deps)
	endpoints.InitPollingEndpoints(router, deps)
}
