package opentelemetry

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/label"
	exporttrace "go.opentelemetry.io/otel/sdk/export/trace"
	"go.opentelemetry.io/otel/trace"
	ottrace "go.opentelemetry.io/otel/trace"

	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/domain"
	ottesting "github.com/saiya/dsps/server/telemetry/opentelemetry/testing"
)

func TestEmptyTracing(t *testing.T) {
	ot, err := NewOTFacility(&config.OpenTelemetryConfig{Tracing: &config.OpenTelemetryTracingConfig{}})
	assert.NoError(t, err)
	defer func() { assert.NoError(t, ot.Shutdown(context.Background())) }()

	_, span := ot.Tracing.Tracer.Start(context.Background(), "test span", trace.WithAttributes(
		label.String("attr1", "value1"),
	))
	span.SetAttributes(label.String("attr2", "value2"))
	span.End()
}

func TestTracing(t *testing.T) {
	stubExporter := ottesting.NewStubExporter(t)

	tracingConfig := tracingConfigDefault()
	tracingConfig.Attributes["global.attr1"] = "value1"
	ot, err := newOTFacility(&config.OpenTelemetryConfig{
		Tracing: tracingConfig,
		Exporters: config.OpenTelemetryExportersConfig{
			Stdout: config.OpenTelemetryExporterStdoutConfig{
				Enable: true,
			},
		},
	}, injection{
		tracing: tracingInjection{
			exporters: []exporttrace.SpanExporter{stubExporter},
		},
	})
	assert.NoError(t, err)
	tracing := ot.Tracing
	defer func() { assert.NoError(t, tracing.Shutdown(context.Background())) }()

	ctx, span := tracing.Tracer.Start(context.Background(), "test span", trace.WithAttributes(
		label.String("attr1", "value1"),
	))
	span.SetAttributes(label.String("attr2", "value2"))
	assert.Same(t, span, ottrace.SpanFromContext(ctx))
	span.End()
	assert.NoError(t, tracing.Shutdown(context.Background())) // Flush

	spanData := stubExporter.AssertSpan(0, ottrace.SpanKindInternal, "test span", map[string]interface{}{
		"attr1": "value1",
		"attr2": "value2",
	})
	assert.EqualValues(t, map[string]interface{}{"global.attr1": "value1"}, ottesting.KVPairsToMap(spanData.Resource.Attributes()))
}

func tracingConfigDefault() *config.OpenTelemetryTracingConfig {
	maxQueueSize := 1
	batchTimeout := domain.Duration{Duration: 1 * time.Microsecond}
	batchSize := 1
	return &config.OpenTelemetryTracingConfig{
		Enable:     true,
		Attributes: map[string]interface{}{},
		Batch: config.OpenTelemetryTracingBatchConfig{
			MaxQueueSize: &maxQueueSize,
			Timeout:      &batchTimeout,
			BatchSize:    &batchSize,
		},
	}
}
