package telemetry

import (
	"context"
	"testing"

	otsdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/saiya/dsps/server/telemetry/opentelemetry"
	ottesting "github.com/saiya/dsps/server/telemetry/opentelemetry/testing"
)

// TraceResult contains stub tracing result.
type TraceResult struct {
	OT *ottesting.StubExporter
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
