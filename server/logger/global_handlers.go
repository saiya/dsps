package logger

import (
	"context"
	"fmt"
	"os"
	"sync"
)

var globalHandlers struct {
	errorHandlersLock sync.RWMutex
	errorHandlers     []ErrorHandler
}

// ErrorHandler function to receive error logs
type ErrorHandler func(ctx context.Context, msg string, err error)

// SetGlobalErrorHandler to set callback function for each error log
func SetGlobalErrorHandler(f ErrorHandler) {
	globalHandlers.errorHandlersLock.Lock()
	defer globalHandlers.errorHandlersLock.Unlock()

	globalHandlers.errorHandlers = append(globalHandlers.errorHandlers, f)
}

func invokeGlobalErrorHandlers(ctx context.Context, msg string, err error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(os.Stderr, "Panic in logging GlobalErrorHandler", r)
		}
	}()

	var handlers []ErrorHandler
	func() {
		globalHandlers.errorHandlersLock.RLock()
		defer globalHandlers.errorHandlersLock.RUnlock()
		handlers = globalHandlers.errorHandlers
	}()
	for _, f := range handlers {
		f(ctx, msg, err)
	}
}
