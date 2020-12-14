package middleware_test

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"

	. "github.com/saiya/dsps/server/http"
	. "github.com/saiya/dsps/server/http/middleware"
	. "github.com/saiya/dsps/server/http/router"
	. "github.com/saiya/dsps/server/http/testing"
	"github.com/saiya/dsps/server/http/utils"
	"github.com/saiya/dsps/server/logger"
)

func TestPanicHandling(t *testing.T) {
	WithServerDeps(t, `logging: { category: "*": ERROR }`, func(deps *ServerDependencies) {
		r := httprouter.New()
		rt := NewRouter(func(r *http.Request, f func(context.Context)) {
			f(context.Background())
		}, r, "/", LoggingMiddleware())
		server := httptest.NewServer(r)
		defer server.Close()

		panicErr := errors.New("test panic error")
		rt.GET("/panic-error", func(ctx context.Context, args HandlerArgs) {
			panic(panicErr)
		})
		panicString := "test panic string"
		rt.GET("/panic-string", func(ctx context.Context, args HandlerArgs) {
			panic(panicString)
		})
		rt.GET("/panic-after-200", func(ctx context.Context, args HandlerArgs) {
			utils.SendJSON(ctx, args.W, 200, map[string]string{"hi": "hello"})
			panic(panicString)
		})
		rt.GET("/panic-after-204", func(ctx context.Context, args HandlerArgs) {
			utils.SendNoContent(ctx, args.W)
			panic(panicString)
		})

		logger.WithTestLogger(t, nil, func(lc *logger.LogCapture) {
			res := DoHTTPRequest(t, "GET", server.URL+"/panic-error", ``)
			assert.NoError(t, res.Body.Close())
			assert.Equal(t, 500, res.StatusCode)

			errorLog := lc.LastLog(0)
			assert.Equal(t, zapcore.ErrorLevel, errorLog.Level)
			assert.Equal(t, "internal server error caught on HTTP endpoint", errorLog.Message)
			assert.Equal(t, panicErr, errorLog.FindErrorField())
		})

		logger.WithTestLogger(t, nil, func(lc *logger.LogCapture) {
			res := DoHTTPRequest(t, "GET", server.URL+"/panic-string", ``)
			assert.NoError(t, res.Body.Close())
			assert.Equal(t, 500, res.StatusCode)

			errorLog := lc.LastLog(0)
			assert.Equal(t, zapcore.ErrorLevel, errorLog.Level)
			assert.Equal(t, "internal server error caught on HTTP endpoint", errorLog.Message)
			assert.Equal(t, fmt.Errorf("%+v", panicString), errorLog.FindErrorField())
		})

		logger.WithTestLogger(t, nil, func(lc *logger.LogCapture) {
			res := DoHTTPRequest(t, "GET", server.URL+"/panic-after-200", ``)
			assert.Equal(t, 200, res.StatusCode) // After response sent.
			body, err := ioutil.ReadAll(res.Body)
			assert.NoError(t, err)
			assert.Equal(t, `{"hi":"hello"}`+"\n", string(body))
			assert.NoError(t, res.Body.Close())

			errorLog := lc.LastLog(0)
			assert.Equal(t, zapcore.ErrorLevel, errorLog.Level)
			assert.Equal(t, "internal server error caught on HTTP endpoint", errorLog.Message)
			assert.Equal(t, fmt.Errorf("%+v", panicString), errorLog.FindErrorField())
		})

		logger.WithTestLogger(t, nil, func(lc *logger.LogCapture) {
			res := DoHTTPRequest(t, "GET", server.URL+"/panic-after-204", ``)
			assert.NoError(t, res.Body.Close())
			assert.Equal(t, 204, res.StatusCode) // After response sent.

			errorLog := lc.LastLog(0)
			assert.Equal(t, zapcore.ErrorLevel, errorLog.Level)
			assert.Equal(t, "internal server error caught on HTTP endpoint", errorLog.Message)
			assert.Equal(t, fmt.Errorf("%+v", panicString), errorLog.FindErrorField())
		})
	})
}
