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
	hideRequestData := config.HideRequestData
	if err := sentrygo.Init(sentrygo.ClientOptions{
		Dsn: config.DSN,

		ServerName:  config.ServerName,
		Environment: config.Environment,
		Dist:        config.Distribution,
		Release:     config.Release,

		SampleRate:       *config.SampleRate,
		AttachStacktrace: !config.DisableStacktrace,
		BeforeSend: func(event *sentrygo.Event, hint *sentrygo.EventHint) *sentrygo.Event {
			if hideRequestData && event.Request != nil {
				event.Request.Data = ""
			}
			return event
		},
		IgnoreErrors: ignoreErrors,
		Debug:        false,
	}); err != nil {
		return nil, xerrors.Errorf("failed to initialize sentry: %w", err)
	}
	sentrygo.ConfigureScope(func(scope *sentrygo.Scope) {
		scope.SetTags(config.Tags)
		for key, value := range config.Contexts {
			scope.SetContext(key, value)
		}
	})
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
