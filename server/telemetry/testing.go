package telemetry

import (
	"context"
	"testing"

	otsdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/telemetry/opentelemetry"
	ottesting "github.com/saiya/dsps/server/telemetry/opentelemetry/testing"
	"github.com/stretchr/testify/assert"
)

// TraceResult contains stub tracing result.
type TraceResult struct {
	OT *ottesting.StubExporter
}

// NewEmptyTelemetry creates no-op telemetry system
func NewEmptyTelemetry(t *testing.T) *Telemetry {
	telemetry, err := InitTelemetry(&config.TelemetryConfig{
		OT: &config.OpenTelemetryConfig{
			Tracing: &config.OpenTelemetryTracingConfig{},
		},
	})
	assert.NoError(t, err)
	return telemetry
}

// WithStubTracing is testing utility to create stub Telemetry facility and returns captured spans.
func WithStubTracing(t *testing.T, f func(*Telemetry)) *TraceResult {
	tr := &TraceResult{
		OT: ottesting.NewStubExporter(t),
	}
	telemetry := &Telemetry{
		ot: &opentelemetry.OTFacility{
			Tracing: &opentelemetry.Tracing{
				Tracer: otsdktrace.NewTracerProvider(otsdktrace.WithSyncer(tr.OT)).Tracer("test"),
			},
		},
	}
	defer telemetry.Shutdown(context.Background())
	f(telemetry)
	return tr
}
