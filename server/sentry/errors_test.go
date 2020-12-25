package sentry

import (
	"context"
	"errors"
	"testing"
)

func TestErrorsWithRealSentry(t *testing.T) {
	emptySentry := NewEmptySentry()
	emptyCtx := emptySentry.WrapContext(context.Background())
	defer emptySentry.Shutdown(context.Background())

	surpressedSentry := newSupressedSentry(t)
	defer surpressedSentry.Shutdown(context.Background())
	surpressedCtx := surpressedSentry.WrapContext(context.Background())

	stubSentry := NewStubSentry()
	defer stubSentry.Shutdown(context.Background())
	stubCtx := stubSentry.WrapContext(context.Background())

	for _, pattern := range []struct {
		sentry Sentry
		ctx    context.Context
	}{
		{sentry: emptySentry, ctx: emptyCtx},
		{sentry: surpressedSentry, ctx: surpressedCtx},
		{sentry: stubSentry, ctx: stubCtx},
	} {
		ctx := pattern.ctx
		RecordError(ctx, errors.New("test error"))
	}
}
