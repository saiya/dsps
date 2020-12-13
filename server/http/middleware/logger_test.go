package middleware_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"

	. "github.com/saiya/dsps/server/http"
	. "github.com/saiya/dsps/server/http/middleware"
	. "github.com/saiya/dsps/server/http/testing"
	"github.com/saiya/dsps/server/logger"
)

func TestPanicHandling(t *testing.T) {
	WithServerDeps(t, `logging: { category: "*": ERROR }`, func(deps *ServerDependencies) {
		rec := httptest.NewRecorder()
		ctx, engine := gin.CreateTestContext(rec)
		engine.Use(LoggingMiddleware())

		panicErr := errors.New("test panic error")
		engine.GET("/panic-error", func(c *gin.Context) {
			panic(panicErr)
		})
		panicString := "test panic string"
		engine.GET("/panic-string", func(c *gin.Context) {
			panic(panicString)
		})

		logger.WithTestLogger(t, nil, func(lc *logger.LogCapture) {
			req, _ := http.NewRequestWithContext(ctx, "GET", "/panic-error", nil)
			engine.ServeHTTP(rec, req)
			assert.Equal(t, 500, rec.Code)

			errorLog := lc.LastLog(0)
			assert.Equal(t, zapcore.ErrorLevel, errorLog.Level)
			assert.Equal(t, "Panic in HTTP endpoint", errorLog.Message)
			assert.Equal(t, panicErr, errorLog.FindErrorField())
		})

		logger.WithTestLogger(t, nil, func(lc *logger.LogCapture) {
			req, _ := http.NewRequestWithContext(ctx, "GET", "/panic-string", nil)
			engine.ServeHTTP(rec, req)
			assert.Equal(t, 500, rec.Code)

			errorLog := lc.LastLog(0)
			assert.Equal(t, zapcore.ErrorLevel, errorLog.Level)
			assert.Equal(t, "Panic in HTTP endpoint", errorLog.Message)
			assert.Equal(t, fmt.Errorf("%+v", panicString), errorLog.FindErrorField())
		})
	})
}
