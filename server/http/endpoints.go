package http

import (
	"github.com/gin-gonic/gin"

	"github.com/saiya/dsps/server/http/endpoints"
)

// InitEndpoints registers endpoints of the DSPS server
func InitEndpoints(router gin.IRoutes, deps *ServerDependencies) {
	endpoints.InitProbeEndpoints(router, deps)
	endpoints.InitPublishEndpoints(router, deps)
	endpoints.InitPollingEndpoints(router, deps)
}
