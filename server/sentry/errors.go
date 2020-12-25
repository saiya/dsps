package sentry

import (
	"context"

	sentrygo "github.com/getsentry/sentry-go"
)

// RecordError send error event to sentry
func RecordError(ctx context.Context, err error) {
	withHub(ctx, func(hub *sentrygo.Hub) {
		hub.CaptureException(err)
	})
	withStubLock(ctx, func(s *StubSentry) {
		s.errors = append(s.errors, err)
	})
}
