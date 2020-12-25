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

	"github.com/saiya/dsps/server/domain"
	. "github.com/saiya/dsps/server/http"
	. "github.com/saiya/dsps/server/http/middleware"
	. "github.com/saiya/dsps/server/http/router"
	. "github.com/saiya/dsps/server/http/testing"
	"github.com/saiya/dsps/server/http/utils"
	"github.com/saiya/dsps/server/logger"
	"github.com/saiya/dsps/server/sentry"
)

func TestLoggerMiddleware(t *testing.T) {
	WithServer(t, `logging: { category: { "*": ERROR, http: INFO } }`, func(deps *ServerDependencies) {}, func(deps *ServerDependencies, baseURL string) {
		logger.WithTestLogger(t, nil, func(lc *logger.LogCapture) {
			url := baseURL + "/probe/liveness"

			// Without IP request header configuration
			res := DoHTTPRequestWithHeaders(t, "GET", url, map[string]string{}, ``)
			assert.NoError(t, res.Body.Close())
			assert.Equal(t, 200, res.StatusCode)

			log := lc.LastLog(0)
			assert.Equal(t, zapcore.InfoLevel, log.Level)
			assert.Equal(t, "HTTP endpoint served", log.Message)
			assert.Equal(t, "GET", log.FindStringField("method"))
			assert.Equal(t, "127.0.0.1", log.FindStringField("ip"))
		})
	})

	WithServer(t, `http: { realIpHeader: "X-Forwarded-For", trustedProxyRanges: [ "127.0.0.0/8", "10.0.0.0/8" ] }, logging: { category: { "*": ERROR, http: INFO } }`, func(deps *ServerDependencies) {}, func(deps *ServerDependencies, baseURL string) {
		logger.WithTestLogger(t, nil, func(lc *logger.LogCapture) {
			url := baseURL + "/probe/liveness"

			// Without IP request header
			res := DoHTTPRequestWithHeaders(t, "GET", url, map[string]string{}, ``)
			assert.NoError(t, res.Body.Close())
			assert.Equal(t, 200, res.StatusCode)

			log := lc.LastLog(0)
			assert.Equal(t, zapcore.InfoLevel, log.Level)
			assert.Equal(t, "HTTP endpoint served", log.Message)
			assert.Equal(t, "GET", log.FindStringField("method"))
			assert.Equal(t, "127.0.0.1", log.FindStringField("ip"))

			// With valid IP request header
			res = DoHTTPRequestWithHeaders(t, "GET", url, map[string]string{
				"X-Forwarded-For": "192.0.2.1, 10.0.0.1, 10.0.0.2, 10.0.0.3",
			}, ``)
			assert.NoError(t, res.Body.Close())
			assert.Equal(t, 200, res.StatusCode)

			log = lc.LastLog(0)
			assert.Equal(t, zapcore.InfoLevel, log.Level)
			assert.Equal(t, "HTTP endpoint served", log.Message)
			assert.Equal(t, "GET", log.FindStringField("method"))
			assert.Equal(t, "192.0.2.1", log.FindStringField("ip"))
		})
	})
}

func TestPanicHandling(t *testing.T) {
	realIPDeps := &stubRealIPDeps{
		ipHeaderName:       "X-Forwarded-For",
		trustedProxyRanges: domain.PrivateCIDRs,
	}
	WithServerDeps(t, `logging: { category: "*": ERROR }`, func(deps *ServerDependencies) {
		sentry := sentry.NewStubSentry()
		deps.Sentry = sentry

		r := httprouter.New()
		rt := NewRouter(func(r *http.Request, f func(context.Context)) {
			f(context.Background())
		}, r, "/", SentryMiddleware(deps), LoggingMiddleware(realIPDeps, deps))
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

			assert.Regexp(t, `test panic error`, sentry.GetLastError().Error())
			errorLog := lc.LastLog(0)
			assert.Equal(t, zapcore.ErrorLevel, errorLog.Level)
			assert.Equal(t, "internal server error on HTTP endpoint", errorLog.Message)
			assert.Equal(t, panicErr, errorLog.FindErrorField())
		})

		logger.WithTestLogger(t, nil, func(lc *logger.LogCapture) {
			res := DoHTTPRequest(t, "GET", server.URL+"/panic-string", ``)
			assert.NoError(t, res.Body.Close())
			assert.Equal(t, 500, res.StatusCode)

			assert.Regexp(t, `test panic string`, sentry.GetLastError().Error())
			errorLog := lc.LastLog(0)
			assert.Equal(t, zapcore.ErrorLevel, errorLog.Level)
			assert.Equal(t, "internal server error on HTTP endpoint", errorLog.Message)
			assert.Equal(t, fmt.Errorf("%+v", panicString), errorLog.FindErrorField())
		})

		logger.WithTestLogger(t, nil, func(lc *logger.LogCapture) {
			res := DoHTTPRequest(t, "GET", server.URL+"/panic-after-200", ``)
			assert.Equal(t, 200, res.StatusCode) // After response sent.
			body, err := ioutil.ReadAll(res.Body)
			assert.NoError(t, err)
			assert.Equal(t, `{"hi":"hello"}`+"\n", string(body))
			assert.NoError(t, res.Body.Close())

			assert.Regexp(t, `test panic string`, sentry.GetLastError().Error())
			errorLog := lc.LastLog(0)
			assert.Equal(t, zapcore.ErrorLevel, errorLog.Level)
			assert.Equal(t, "internal server error on HTTP endpoint", errorLog.Message)
			assert.Equal(t, fmt.Errorf("%+v", panicString), errorLog.FindErrorField())
		})

		logger.WithTestLogger(t, nil, func(lc *logger.LogCapture) {
			res := DoHTTPRequest(t, "GET", server.URL+"/panic-after-204", ``)
			assert.NoError(t, res.Body.Close())
			assert.Equal(t, 204, res.StatusCode) // After response sent.

			assert.Regexp(t, `test panic string`, sentry.GetLastError().Error())
			errorLog := lc.LastLog(0)
			assert.Equal(t, zapcore.ErrorLevel, errorLog.Level)
			assert.Equal(t, "internal server error on HTTP endpoint", errorLog.Message)
			assert.Equal(t, fmt.Errorf("%+v", panicString), errorLog.FindErrorField())
		})
	})
}
