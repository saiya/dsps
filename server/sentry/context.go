package sentry

import (
	"context"

	sentrygo "github.com/getsentry/sentry-go"
)

type sentryContextKey int

const stubSentryKey = sentryContextKey(1)

func (s *sentry) WrapContext(ctx context.Context) context.Context {
	return sentrygo.SetHubOnContext(ctx, sentrygo.CurrentHub().Clone())
}
func (s *emptySentry) WrapContext(ctx context.Context) context.Context {
	return ctx
}

// WrapContext wraps context to activate Sentry.
func (s *StubSentry) WrapContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, stubSentryKey, s)
}

func withHub(ctx context.Context, f func(*sentrygo.Hub)) {
	hub := sentrygo.GetHubFromContext(ctx)
	if hub != nil {
		f(hub)
	}
}

func withStubLock(ctx context.Context, f func(*StubSentry)) {
	if stub, ok := ctx.Value(stubSentryKey).(*StubSentry); ok {
		stub.lock.Lock()
		defer stub.lock.Unlock()
		f(stub)
	}
}
