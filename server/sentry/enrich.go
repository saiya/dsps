package sentry

import (
	"context"

	sentrygo "github.com/getsentry/sentry-go"
)

// AddBreadcrumb adds breadcrumb (trail of operations).
// https://docs.sentry.io/platforms/go/enriching-events/breadcrumbs/
func AddBreadcrumb(ctx context.Context, breadcrumb *sentrygo.Breadcrumb) {
	withHub(ctx, func(hub *sentrygo.Hub) {
		hub.AddBreadcrumb(breadcrumb, nil)
	})
	withStubLock(ctx, func(s *StubSentry) {
		s.breadcrumbs = append(s.breadcrumbs, breadcrumb)
	})
}

// AddTag adds tag (searchable value).
// https://docs.sentry.io/platforms/go/enriching-events/tags/
func AddTag(ctx context.Context, key, value string) {
	withHub(ctx, func(hub *sentrygo.Hub) {
		hub.Scope().SetTag(key, value)
	})
	withStubLock(ctx, func(s *StubSentry) {
		s.tags[key] = value
	})
}

// AddContext add "context" (sentry term, un-searchable attribute) value.
// https://docs.sentry.io/platforms/go/enriching-events/context/
func AddContext(ctx context.Context, key string, value interface{}) {
	withHub(ctx, func(hub *sentrygo.Hub) {
		hub.Scope().SetContext(key, value)
	})
	withStubLock(ctx, func(s *StubSentry) {
		s.context[key] = value
	})
}

// SetIPAddress set IP address information of the client
func SetIPAddress(ctx context.Context, value string) {
	withHub(ctx, func(hub *sentrygo.Hub) {
		hub.Scope().SetUser(sentrygo.User{
			IPAddress: value,
		})
	})
}
