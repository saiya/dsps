package config

import (
	"fmt"
	"os"

	"github.com/saiya/dsps/server/domain"
)

// SentryConfig contains Sentry configuration
type SentryConfig struct {
	DSN string `json:"-"` // Given by environment variable

	ServerName   string `json:"serverName"`
	Environment  string `json:"environment"`
	Release      string `json:"-"` // Given by compilation time constant value
	Distribution string `json:"-"` // Given by compilation time constant value

	Tags     map[string]string `json:"tags"`
	Contexts map[string]string `json:"contexts"`

	SampleRate        *float64        `json:"sampleRate"`
	IgnoreErrors      []*domain.Regex `json:"ignoreErrors"`
	DisableStacktrace bool            `json:"disableStacktrace"`
	HideRequestData   bool            `json:"hideRequestData"`

	FlushTimeout *domain.Duration `json:"flushTimeout"`
}

// DefaultSentryConfig returns empty object
func DefaultSentryConfig() *SentryConfig {
	return &SentryConfig{}
}

// PostprocessSentryConfig fixup given configurations
func PostprocessSentryConfig(config *SentryConfig, overrides Overrides) error {
	config.DSN = os.Getenv("SENTRY_DSN")
	if config.ServerName == "" {
		if hostname, err := os.Hostname(); err == nil {
			config.ServerName = hostname
		}
	}
	config.Release = overrides.BuildVersion
	config.Distribution = overrides.BuildDist
	if config.SampleRate == nil {
		config.SampleRate = makeFloat64Ptr(1.0)
	}
	if config.FlushTimeout == nil {
		config.FlushTimeout = makeDurationPtr("15s")
	}

	if *config.SampleRate < 0 || 1 < *config.SampleRate {
		return fmt.Errorf(`sample ratio must be within [0.0, 1.0]`)
	}
	if config.FlushTimeout.Duration <= 0 {
		return fmt.Errorf(`flushTimeout must be larger than zero`)
	}

	return nil
}
