package http

import (
	"context"

	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/http/endpoints"
	"github.com/saiya/dsps/server/http/middleware"
	"github.com/saiya/dsps/server/http/router"
	"github.com/saiya/dsps/server/logger"
)

// InitEndpoints registers endpoints of the DSPS server
func InitEndpoints(mainCtx context.Context, rt *router.Router, deps *ServerDependencies) {
	endpoints.InitProbeEndpoints(rt, deps)

	adminRouter := rt.NewGroup("/admin", middleware.NewAdminAuth(mainCtx, deps))
	endpoints.InitAdminJwtEndpoints(adminRouter, deps)
	endpoints.InitAdminLoggingEndpoints(adminRouter, deps)

	channelRouter := rt.NewGroup(
		"/channel/:channelID",
		func(ctx context.Context, args router.MiddlewareArgs, next func(context.Context)) {
			next(logger.WithAttributes(ctx).WithStr("channelID", args.PS.ByName("channelID")).Build())
		},
		middleware.NewNormalAuth(mainCtx, deps, func(c context.Context, args router.MiddlewareArgs) (domain.Channel, error) {
			id, err := domain.ParseChannelID(args.PS.ByName("channelID"))
			if err != nil {
				return nil, err
			}
			return deps.ChannelProvider(id)
		}),
	)
	endpoints.InitPublishEndpoints(channelRouter, deps)
	endpoints.InitSubscriptionPollingEndpoints(channelRouter, deps)
}
