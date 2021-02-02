package sentry

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/domain"
)

func TestNewSentry(t *testing.T) {
	sentry, err := NewSentry(nil)
	_, ok := sentry.(*emptySentry)
	assert.NoError(t, err)
	assert.True(t, ok)
	sentry.Shutdown(context.Background())

	sentry, err = NewSentry(&config.SentryConfig{
		DSN: "",
		Contexts: map[string]string{
			"test": "value",
		},
		HideRequestData: true,
	})
	_, ok = sentry.(*emptySentry)
	assert.NoError(t, err)
	assert.True(t, ok)
	sentry.Shutdown(context.Background())

	sampleRate := 1.0
	_, err = NewSentry(&config.SentryConfig{
		DSN:        "***invalid dsn***",
		SampleRate: &sampleRate,
	})
	assert.Regexp(t, `DsnParseError`, err)

	sentry = newSupressedSentry(t)
	_, ok = sentry.(*emptySentry)
	assert.False(t, ok)
	sentry.Shutdown(context.Background())
}

func newSupressedSentry(t *testing.T) Sentry {
	sampleRate := 0.0
	regex, err := domain.NewRegex(".*")
	assert.NoError(t, err)
	sentry, err := NewSentry(&config.SentryConfig{
		DSN:          "http://dummy@example.com/1",
		SampleRate:   &sampleRate,
		IgnoreErrors: []*domain.Regex{regex},
		FlushTimeout: &domain.Duration{Duration: 0 * time.Second},
	})
	assert.NoError(t, err)
	return sentry
}
