package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/saiya/dsps/server/http/router"
	"github.com/saiya/dsps/server/http/utils"
	"github.com/saiya/dsps/server/logger"
)

// LoggingMiddleware is middleware for logging
func LoggingMiddleware(realIPDeps RealIPDependency) router.MiddlewareFunc {
	return router.AsMiddlewareFunc(func(ctx context.Context, args router.MiddlewareArgs, next func(context.Context, router.MiddlewareArgs)) {
		defer func() {
			if err := recover(); err != nil {
				utils.SendInternalServerError(ctx, args.W, panicAsError(err))
			}
		}()

		ctx = logger.WithAttributes(ctx).
			WithStr("method", args.R.Method).
			WithStr("path", args.R.URL.Path).
			WithStr("ip", GetRealIP(realIPDeps, args.R)).
			WithStr("ua", args.R.UserAgent()).
			WithStr("referer", args.R.Referer()).
			WithInt64("reqLength", args.R.ContentLength).
			Build()

		startAt := time.Now()
		next(ctx, args)
		elapsed := time.Since(startAt)

		ctx = logger.WithAttributes(ctx).
			WithInt("status", args.W.Written().StatusCode).
			WithInt64("elapsedMs", elapsed.Milliseconds()).
			WithInt("resLength", args.W.Written().BodyBytes).
			Build()
		logger.Of(ctx).Infof(logger.CatHTTP, "HTTP endpoint served")
	})
}

func panicAsError(err interface{}) error {
	if e, ok := err.(error); ok {
		return e
	}
	return fmt.Errorf("%+v", err)
}
