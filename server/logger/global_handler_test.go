package logger_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	. "github.com/saiya/dsps/server/logger"
	"github.com/stretchr/testify/assert"
)

func TestGlobalErrorHandler(t *testing.T) {
	var ctx context.Context
	var msg string
	var err error
	handlerCalled := 0
	setVolatileGlobalErrorHandler(t, func(ctxActual context.Context, msgActual string, errActual error) {
		handlerCalled++
		assert.Same(t, ctx, ctxActual)
		assert.Equal(t, msg, msgActual)
		if err == nil {
			assert.Nil(t, errActual)
		} else {
			assert.Same(t, err, errActual)
		}
	})

	ctx = context.Background()
	msg = "test error message"
	err = nil
	Of(ctx).Error(msg, err)
	assert.Equal(t, handlerCalled, 1)

	ctx = context.Background()
	msg = "test error message"
	err = errors.New("test error")
	Of(ctx).Error(msg, err)
	assert.Equal(t, handlerCalled, 2)

	ctx = WithAttributes(context.Background()).WithStr("attr", "value").Build()
	msg = "test error message"
	err = errors.New("test error")
	Of(ctx).Error(msg, err)
	assert.Equal(t, handlerCalled, 3)
}

func TestGlobalErrorHandlerPanic(t *testing.T) {
	setVolatileGlobalErrorHandler(t, func(context.Context, string, error) {
		panic("test panic") // Should not stop program
	})

	Of(context.Background()).Error("test error log", nil)
}

func setVolatileGlobalErrorHandler(t *testing.T, f ErrorHandler) {
	enabled := int32(1)
	t.Cleanup(func() {
		atomic.StoreInt32(&enabled, 0)
	})
	SetGlobalErrorHandler(func(ctx context.Context, msg string, err error) {
		if atomic.LoadInt32(&enabled) != 1 {
			return
		}
		f(ctx, msg, err)
	})
}
