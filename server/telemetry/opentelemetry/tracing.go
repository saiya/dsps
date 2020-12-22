package opentelemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/label"
	exporttrace "go.opentelemetry.io/otel/sdk/export/trace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	ottrace "go.opentelemetry.io/otel/trace"

	"github.com/saiya/dsps/server/config"
)

// Tracing contains OT tracing resources
type Tracing struct {
	Tracer ottrace.Tracer

	sampler       sdktrace.Sampler
	exporters     []exporttrace.SpanExporter
	resource      *resource.Resource
	traceProvider *sdktrace.TracerProvider
}

type tracingInjection struct {
	exporters []exporttrace.SpanExporter
}

func newTracing(config *config.OpenTelemetryTracingConfig, injection tracingInjection) (setup *Tracing, err error) {
	setup = &Tracing{}
	if !config.Enable {
		setup.Tracer = &emptyTracer{}
	} else {
		setup.sampler = sdktrace.ParentBased(sdktrace.TraceIDRatioBased(config.Sampling))
		setup.exporters = injection.exporters
		setup.resource = resource.NewWithAttributes(initTracingGlobalLabels(config)...)
		setup.traceProvider = sdktrace.NewTracerProvider(append(
			wrapTraceExporters(config, setup.exporters),
			sdktrace.WithResource(setup.resource),
		)...)
		setup.Tracer = setup.traceProvider.Tracer("github.com/saiya/dsps")

		// Set global settings
		otel.SetTracerProvider(setup.traceProvider)
	}
	return
}

// Shutdown OT tracer
func (setup *Tracing) Shutdown(ctx context.Context) error {
	if setup.traceProvider != nil {
		return setup.traceProvider.Shutdown(ctx)
	}
	return nil
}

func initTracingGlobalLabels(config *config.OpenTelemetryTracingConfig) []label.KeyValue {
	labels := make([]label.KeyValue, 0, len(config.Attributes))
	for key, value := range config.Attributes {
		labels = append(labels, label.Any(key, value))
	}
	return labels
}

func wrapTraceExporters(config *config.OpenTelemetryTracingConfig, exporters []exporttrace.SpanExporter) []sdktrace.TracerProviderOption {
	result := make([]sdktrace.TracerProviderOption, 0, len(exporters))
	for _, exporter := range exporters {
		result = append(result, sdktrace.WithBatcher(
			exporter,
			sdktrace.WithMaxExportBatchSize(*config.Batch.BatchSize),
			sdktrace.WithBatchTimeout(config.Batch.Timeout.Duration),
			sdktrace.WithMaxQueueSize(*config.Batch.MaxQueueSize),
		))
	}
	return result
}

type emptyTracer struct{}

var noopSpan = ottrace.SpanFromContext(context.Background())

// Implements ottrace.Tracer
func (t *emptyTracer) Start(ctx context.Context, spanName string, opts ...ottrace.SpanOption) (context.Context, ottrace.Span) {
	return ctx, noopSpan
}
