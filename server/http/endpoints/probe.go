package endpoints

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/saiya/dsps/server/domain"
)

// ProbeEndpointDependency is to inject required objects to the endpoint
type ProbeEndpointDependency interface {
	GetStorage() domain.Storage
}

// InitProbeEndpoints registers endpoints
func InitProbeEndpoints(router gin.IRoutes, deps ProbeEndpointDependency) {
	router.GET("/probe/liveness", probeEndpointImpl(deps, func(ctx context.Context, storage domain.Storage) (interface{}, error) { return storage.Liveness(ctx) }))
	router.GET("/probe/readiness", probeEndpointImpl(deps, func(ctx context.Context, storage domain.Storage) (interface{}, error) { return storage.Readiness(ctx) }))
}

func probeEndpointImpl(deps ProbeEndpointDependency, storageHandler func(ctx context.Context, storage domain.Storage) (interface{}, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		status := http.StatusOK

		storage, storageErr := storageHandler(ctx, deps.GetStorage())
		if storageErr != nil {
			status = http.StatusInternalServerError
			storage = storageErr
		}

		ctx.JSON(status, gin.H{
			"storage": storage,
		})
	}
}
