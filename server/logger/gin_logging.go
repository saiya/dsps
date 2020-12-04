package logger

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggingMiddleware is middleware for logging
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				Of(c).Error("Panic in HTTP endpoint", panicAsError(err))
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()

		ModifyGinContext(c).
			WithStr("method", c.Request.Method).
			WithStr("path", c.Request.URL.Path).
			WithStr("ip", c.ClientIP()).
			WithStr("ua", c.Request.UserAgent()).
			WithStr("referer", c.Request.Referer()).
			WithInt64("reqLength", c.Request.ContentLength).
			Build()

		startAt := time.Now()
		c.Next()
		elapsed := time.Since(startAt)

		ModifyGinContext(c).
			WithInt("status", c.Writer.Status()).
			WithInt64("elapsedMs", elapsed.Milliseconds()).
			WithInt("resLength", c.Writer.Size()).
			Build()
		Of(c).Infof("HTTP endpoint ended")
	}
}

func panicAsError(err interface{}) error {
	if e, ok := err.(error); ok {
		return e
	}
	return fmt.Errorf("%+v", err)
}
