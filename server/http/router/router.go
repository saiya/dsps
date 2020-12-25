package router

import (
	"context"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

// Router represents tree of endpoints
type Router struct {
	r  *httprouter.Router
	cp ContextProvider

	pathPrefix      string
	middlewareFuncs []MiddlewareFunc
}

// Handler is a implementation of a endpoint
type Handler func(ctx context.Context, args HandlerArgs)

// ContextProvider is a function to create Context for each request
type ContextProvider func(r *http.Request, f func(context.Context))

// NewRouter creates new router (root node of router tree)
func NewRouter(cp ContextProvider, r *httprouter.Router, pathPrefix string, middlewareFuncs ...MiddlewareFunc) *Router {
	return &Router{
		r:  r,
		cp: cp,

		pathPrefix:      concatPath(pathPrefix),
		middlewareFuncs: middlewareFuncs,
	}
}

// NewGroup creates new child node of the router tree
func (rt *Router) NewGroup(pathPrefix string, middlewareFuncs ...MiddlewareFunc) *Router {
	return &Router{
		r:  rt.r,
		cp: rt.cp,

		pathPrefix:      concatPath(rt.pathPrefix, pathPrefix),
		middlewareFuncs: append(rt.middlewareFuncs, middlewareFuncs...),
	}
}

// GET endpoint registration
func (rt *Router) GET(pathPrefix string, h Handler) {
	rt.register("GET", pathPrefix, h)
}

// PUT endpoint registration
func (rt *Router) PUT(pathPrefix string, h Handler) {
	rt.register("PUT", pathPrefix, h)
}

// POST endpoint registration
func (rt *Router) POST(pathPrefix string, h Handler) {
	rt.register("POST", pathPrefix, h)
}

// DELETE endpoint registration
func (rt *Router) DELETE(pathPrefix string, h Handler) {
	rt.register("DELETE", pathPrefix, h)
}

func (rt *Router) register(method string, pathPrefix string, h Handler) {
	path := concatPath(rt.pathPrefix, pathPrefix)
	rt.r.Handle(method, path, rt.wrap(method, path, h))
}

func (rt *Router) wrap(method, path string, h Handler) httprouter.Handle {
	middlewares := make([]Middleware, len(rt.middlewareFuncs))
	for i, mf := range rt.middlewareFuncs {
		middlewares[i] = mf(method, path)
	}

	var caller func(ctx context.Context, middlewareIndex int, args MiddlewareArgs)
	caller = func(ctx context.Context, middlewareIndex int, args MiddlewareArgs) {
		if middlewareIndex < len(middlewares) {
			middlewares[middlewareIndex](ctx, args, func(ctx context.Context, args MiddlewareArgs) {
				args.R = Request{
					Request: args.R.WithContext(ctx),
				}
				caller(ctx, middlewareIndex+1, args)
			})
			return
		}
		h(ctx, args.HandlerArgs)
	}
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		rt.cp(r, func(ctx context.Context) {
			caller(ctx, 0, MiddlewareArgs{HandlerArgs: HandlerArgs{
				W:  NewResponseWriter(w),
				R:  Request{Request: r},
				PS: ps,
			}})
		})
	}
}

func concatPath(components ...string) string {
	var result strings.Builder
	for _, c := range components {
		c = strings.TrimSuffix(strings.TrimPrefix(c, "/"), "/")
		if c == "" {
			continue
		}
		result.WriteString("/")
		result.WriteString(c)
	}
	if result.Len() == 0 {
		result.WriteString("/")
	}
	return result.String()
}
