package logger

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggingMiddleware is middleware for logging
func LoggingMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				Of(ctx).Error("Panic in HTTP endpoint", panicAsError(err))
				ctx.AbortWithStatus(http.StatusInternalServerError)
			}
		}()

		ModifyGinContext(ctx).
			WithStr("method", ctx.Request.Method).
			WithStr("path", ctx.Request.URL.Path).
			WithStr("ip", ctx.ClientIP()).
			WithStr("ua", ctx.Request.UserAgent()).
			WithStr("referer", ctx.Request.Referer()).
			WithInt64("reqLength", ctx.Request.ContentLength).
			Build()

		startAt := time.Now()
		ctx.Next()
		elapsed := time.Since(startAt)

		ModifyGinContext(ctx).
			WithInt("status", ctx.Writer.Status()).
			WithInt64("elapsedMs", elapsed.Milliseconds()).
			WithInt("resLength", ctx.Writer.Size()).
			Build()
		Of(ctx).Infof("HTTP endpoint ended")
	}
}

func panicAsError(err interface{}) error {
	if e, ok := err.(error); ok {
		return e
	}
	return fmt.Errorf("%+v", err)
}
