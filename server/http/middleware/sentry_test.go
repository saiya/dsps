package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	sentrygo "github.com/getsentry/sentry-go"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"

	. "github.com/saiya/dsps/server/http"
	. "github.com/saiya/dsps/server/http/middleware"
	. "github.com/saiya/dsps/server/http/router"
	. "github.com/saiya/dsps/server/http/testing"
	"github.com/saiya/dsps/server/sentry"
)

func TestSentryMiddleware(t *testing.T) {
	sentry := sentry.NewStubSentry()
	WithServerDeps(t, ``, func(deps *ServerDependencies) {
		deps.Sentry = sentry
		r := httprouter.New()
		rt := NewRouter(func(r *http.Request, f func(context.Context)) {
			f(context.Background())
		}, r, "/", SentryMiddleware(deps))
		server := httptest.NewServer(r)
		defer server.Close()

		called := int32(0)
		rt.GET("/", func(ctx context.Context, args HandlerArgs) {
			atomic.AddInt32(&called, 1)

			assert.NotNil(t, sentrygo.GetHubFromContext(ctx))
			assert.Same(t, ctx, args.R.Context())
		})
	})
}
