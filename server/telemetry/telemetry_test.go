package telemetry

import (
	"context"
	"errors"
	"testing"

	"github.com/saiya/dsps/server/config"
	"github.com/stretchr/testify/assert"
)

func TestTelemetryInitFailure(t *testing.T) {
	cfg := defaultTelemetryConfig()
	cfg.OT.Exporters.Stdout.Enable = true
	cfg.OT.Exporters.Stdout.Quantiles = []float64{-1.0} // Invalid value

	_, err := InitTelemetry(cfg)
	assert.Regexp(t, `failed to initialize OpenTelemetry stdout trace exporter: the requested quantile is out of range`, err.Error())
}

func TestTelemetryShutdownFailure(t *testing.T) {
	telemetry := initEmptyTelemetry(t)
	telemetry.ot.TestShutdownError = errors.New("test error")
	telemetry.Shutdown(context.Background()) // Should not stop program
}

func defaultTelemetryConfig() *config.TelemetryConfig {
	return &config.TelemetryConfig{
		OT: &config.OpenTelemetryConfig{
			Tracing: &config.OpenTelemetryTracingConfig{},
		},
	}
}

func initEmptyTelemetry(t *testing.T) *Telemetry {
	telemetry, err := InitTelemetry(defaultTelemetryConfig())
	assert.NoError(t, err)
	return telemetry
}
