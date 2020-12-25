package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"

	"github.com/saiya/dsps/server/domain"
	. "github.com/saiya/dsps/server/http"
	. "github.com/saiya/dsps/server/http/middleware"
	. "github.com/saiya/dsps/server/http/router"
	. "github.com/saiya/dsps/server/http/testing"
	"github.com/saiya/dsps/server/sentry"
)

type stubRealIPDeps struct {
	ipHeaderName       string
	trustedProxyRanges []domain.CIDR
}

func (deps *stubRealIPDeps) GetIPHeaderName() string {
	return deps.ipHeaderName
}
func (deps *stubRealIPDeps) GetTrustedProxyRanges() []domain.CIDR {
	return deps.trustedProxyRanges
}

func TestRealIPMiddlewareXFF(t *testing.T) {
	sentry := sentry.NewStubSentry()
	WithServerDeps(t, `http: { realIpHeader: "X-Forwarded-For" }, logging: { category: "*": ERROR }`, func(deps *ServerDependencies) {
		deps.Sentry = sentry
		r := httprouter.New()
		rt := NewRouter(func(r *http.Request, f func(context.Context)) {
			f(context.Background())
		}, r, "/", SentryMiddleware(deps), RealIPMiddleware(deps))
		server := httptest.NewServer(r)
		defer server.Close()

		var lastRealIP string
		rt.GET("/", func(ctx context.Context, args HandlerArgs) {
			lastRealIP = GetRealIP(deps, args.R)
		})

		assert.Equal(t, "X-Forwarded-For", deps.GetIPHeaderName())

		// Without header
		DoHTTPRequestWithHeaders(t, "GET", server.URL+"/", map[string]string{}, ``)
		assert.Equal(t, "127.0.0.1", lastRealIP)
		assert.Equal(t, "127.0.0.1", sentry.GetTags()["remote_addr"])
		assert.Equal(t, "127.0.0.1", sentry.GetTags()["x_forwarded_for"])

		// With valid XFF, without chaining
		DoHTTPRequestWithHeaders(t, "GET", server.URL+"/", map[string]string{
			"X-Forwarded-For": `192.0.2.1`,
		}, ``)
		assert.Equal(t, "192.0.2.1", lastRealIP)
		assert.Equal(t, "192.0.2.1", sentry.GetTags()["remote_addr"])
		assert.Equal(t, "192.0.2.1", sentry.GetTags()["x_forwarded_for"])

		// With valid XFF, with chaining
		DoHTTPRequestWithHeaders(t, "GET", server.URL+"/", map[string]string{
			"X-Forwarded-For": `192.0.2.1, 192.168.0.1`,
		}, ``)
		assert.Equal(t, "192.0.2.1", lastRealIP)
		assert.Equal(t, "192.0.2.1", sentry.GetTags()["remote_addr"])
		assert.Equal(t, "192.0.2.1", sentry.GetTags()["x_forwarded_for"])

		// With XFF, contains untrusted IP in the chain
		DoHTTPRequestWithHeaders(t, "GET", server.URL+"/", map[string]string{
			"X-Forwarded-For": `192.0.2.1, 192.168.0.1, 192.0.2.2, 192.168.0.1`,
		}, ``)
		assert.Equal(t, "192.0.2.2", lastRealIP)
		assert.Equal(t, "192.0.2.2", sentry.GetTags()["remote_addr"])
		assert.Equal(t, "192.0.2.2", sentry.GetTags()["x_forwarded_for"])
	})
}

func TestRealIPMiddlewareWithoutHeader(t *testing.T) {
	sentry := sentry.NewStubSentry()
	WithServerDeps(t, `logging: { category: "*": ERROR }`, func(deps *ServerDependencies) {
		deps.Sentry = sentry
		r := httprouter.New()
		rt := NewRouter(func(r *http.Request, f func(context.Context)) {
			f(context.Background())
		}, r, "/", SentryMiddleware(deps), RealIPMiddleware(deps))
		server := httptest.NewServer(r)
		defer server.Close()

		var lastRealIP string
		rt.GET("/", func(ctx context.Context, args HandlerArgs) {
			lastRealIP = GetRealIP(deps, args.R)
		})

		assert.Equal(t, "", deps.GetIPHeaderName())

		// Without header
		DoHTTPRequestWithHeaders(t, "GET", server.URL+"/", map[string]string{}, ``)
		assert.Equal(t, "127.0.0.1", lastRealIP)
		assert.Equal(t, "127.0.0.1", sentry.GetTags()["remote_addr"])

		// With XFF (should be ignored)
		DoHTTPRequestWithHeaders(t, "GET", server.URL+"/", map[string]string{
			"X-Forwarded-For": `192.0.2.1`,
		}, ``)
		assert.Equal(t, "127.0.0.1", lastRealIP)
		assert.Equal(t, "127.0.0.1", sentry.GetTags()["remote_addr"])
	})
}
