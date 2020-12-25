package sentry_test

import (
	"context"
	"errors"
	"testing"

	sentrygo "github.com/getsentry/sentry-go"
	"github.com/stretchr/testify/assert"

	. "github.com/saiya/dsps/server/sentry"
)

func TestStub(t *testing.T) {
	sentry := NewStubSentry()
	assert.Nil(t, sentry.GetLastError())

	ctx := sentry.WrapContext(context.Background())

	RecordError(ctx, errors.New("test error"))
	assert.Equal(t, "test error", sentry.GetLastError().Error())

	breadcrumb := &sentrygo.Breadcrumb{}
	AddBreadcrumb(ctx, breadcrumb)
	assert.Same(t, breadcrumb, sentry.GetBreadcrumbs()[0])

	AddTag(ctx, "tag_key", "tag value")
	assert.Equal(t, "tag value", sentry.GetTags()["tag_key"])

	AddContext(ctx, "context_key", "context value")
	assert.Equal(t, "context value", sentry.GetContext()["context_key"])
}
