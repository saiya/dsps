package router

import (
	"github.com/julienschmidt/httprouter"
)

// HandlerArgs holds arguments of handler invoke
type HandlerArgs struct {
	W  ResponseWriter
	R  Request
	PS httprouter.Params
}
