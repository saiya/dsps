package sentry

import (
	"context"
	"testing"

	sentrygo "github.com/getsentry/sentry-go"
)

func TestEnrich(t *testing.T) {
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
		AddBreadcrumb(ctx, &sentrygo.Breadcrumb{
			Type:    "unknown",
			Level:   sentrygo.LevelDebug,
			Message: "test data",
		})
		AddTag(ctx, "key", "value")
		AddContext(ctx, "key", "value")
		SetIPAddress(ctx, "127.0.0.1")
	}
}
