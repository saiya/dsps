package sentry

import (
	"context"
	"time"

	sentrygo "github.com/getsentry/sentry-go"
	"github.com/saiya/dsps/server/config"
	"golang.org/x/xerrors"
)

// Sentry integration interface.
type Sentry interface {
	// WrapContext makes dedicated Sentry context object("hub") for the context.
	// Should create separate "Hub" instances for each requests / background tasks.
	// https://docs.sentry.io/platforms/go/concurrency/
	WrapContext(ctx context.Context) context.Context

	Shutdown(ctx context.Context)
}

type sentry struct {
	flushTimeout time.Duration
}

type emptySentry struct{}

// NewSentry initialize Sentry integration.
func NewSentry(config *config.SentryConfig) (Sentry, error) {
	if config == nil || config.DSN == "" {
		return NewEmptySentry(), nil
	}

	ignoreErrors := make([]string, len(config.IgnoreErrors))
	for i, regex := range config.IgnoreErrors {
		ignoreErrors[i] = regex.String()
	}
	if err := sentrygo.Init(sentrygo.ClientOptions{
		Dsn: config.DSN,

		ServerName:  config.ServerName,
		Environment: config.Environment,
		Dist:        config.Distribution,
		Release:     config.Release,

		SampleRate:       *config.SampleRate,
		AttachStacktrace: !config.DisableStacktrace,
		IgnoreErrors:     ignoreErrors,
		Debug:            false,
	}); err != nil {
		return nil, xerrors.Errorf("failed to initialize sentry: %w", err)
	}
	return &sentry{
		flushTimeout: config.FlushTimeout.Duration,
	}, nil
}

// NewEmptySentry creates empty (no-op) Sentry
func NewEmptySentry() Sentry {
	return &emptySentry{}
}

func (s *sentry) Shutdown(ctx context.Context) {
	sentrygo.Flush(s.flushTimeout)
	sentrygo.AddBreadcrumb(&sentrygo.Breadcrumb{
		Level: sentrygo.LevelDebug,
	})
}
func (s *emptySentry) Shutdown(ctx context.Context) {}

// Shutdown stub implementation
func (s *StubSentry) Shutdown(ctx context.Context) {}
