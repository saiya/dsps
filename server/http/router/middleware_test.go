package router_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"

	. "github.com/saiya/dsps/server/http/router"
	. "github.com/saiya/dsps/server/http/testing"
	"github.com/saiya/dsps/server/http/utils"
)

func TestRouterMiddlewares(t *testing.T) {
	r := httprouter.New()
	rt := NewRouter(func(r *http.Request, f func(context.Context)) {
		f(context.Background())
	}, r, "/", AsMiddlewareFunc(func(ctx context.Context, args MiddlewareArgs, next func(context.Context, MiddlewareArgs)) {
		args.W.Header().Add("middleware", "1")
		next(ctx, args)
	}), AsMiddlewareFunc(func(ctx context.Context, args MiddlewareArgs, next func(context.Context, MiddlewareArgs)) {
		args.W.Header().Add("middleware", "2")
		next(ctx, args)
	}))
	rt.GET("/", func(ctx context.Context, args HandlerArgs) {
		utils.SendJSON(ctx, args.W, 200, map[string]interface{}{"ok": "/"})
	})
	server := httptest.NewServer(r)
	defer server.Close()

	res := DoHTTPRequest(t, "GET", server.URL+"/", ``)
	assert.Equal(t, []string{"1", "2"}, res.Header.Values("middleware"))
	AssertResponseJSON(t, res, 200, map[string]interface{}{"ok": "/"})
}

func TestNetHttpMiddlewareWrap(t *testing.T) {
	orgCtx := context.Background()
	orgR := httptest.NewRequest("GET", "/", bytes.NewBufferString("")).WithContext(orgCtx)
	expectedCtx, cancel := context.WithCancel(orgCtx)
	defer cancel()
	expectedR := orgR.WithContext(expectedCtx)

	h := WrapMiddleware(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Same(t, orgR, r)
			assert.Same(t, orgCtx, orgR.Context())
			h.ServeHTTP(w, expectedR)
		})
	})
	h(orgCtx, MiddlewareArgs{
		HandlerArgs: HandlerArgs{
			R: Request{Request: orgR},
			W: NewResponseWriter(httptest.NewRecorder()),
		},
	}, func(ctx context.Context, args MiddlewareArgs) {
		assert.Same(t, expectedCtx, ctx)
		assert.Same(t, expectedR, args.R.Request)
	})
}
