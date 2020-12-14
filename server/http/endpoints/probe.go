package endpoints

import (
	"context"
	"net/http"

	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/http/router"
	"github.com/saiya/dsps/server/http/utils"
)

// ProbeEndpointDependency is to inject required objects to the endpoint
type ProbeEndpointDependency interface {
	GetStorage() domain.Storage
}

// InitProbeEndpoints registers endpoints
func InitProbeEndpoints(rt *router.Router, deps ProbeEndpointDependency) {
	rt.GET("/probe/liveness", probeEndpointImpl(deps, func(ctx context.Context, storage domain.Storage) (interface{}, error) { return storage.Liveness(ctx) }))
	rt.GET("/probe/readiness", probeEndpointImpl(deps, func(ctx context.Context, storage domain.Storage) (interface{}, error) { return storage.Readiness(ctx) }))
}

func probeEndpointImpl(deps ProbeEndpointDependency, storageHandler func(ctx context.Context, storage domain.Storage) (interface{}, error)) router.Handler {
	return func(ctx context.Context, args router.HandlerArgs) {
		status := http.StatusOK

		storage, storageErr := storageHandler(ctx, deps.GetStorage())
		if storageErr != nil {
			status = http.StatusInternalServerError
			storage = storageErr
		}

		utils.SendJSON(ctx, args.W, status, map[string]interface{}{
			"storage": storage,
		})
	}
}
