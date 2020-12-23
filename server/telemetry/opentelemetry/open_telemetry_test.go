package opentelemetry

import (
	"context"
	"errors"
	"testing"

	"github.com/saiya/dsps/server/config"
	ottesting "github.com/saiya/dsps/server/telemetry/opentelemetry/testing"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	otsdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"github.com/saiya/dsps/server/logger"
)

func TestOTInitFailure(t *testing.T) {
	_, err := newOTFacility(&config.OpenTelemetryConfig{
		Tracing: &config.OpenTelemetryTracingConfig{},
		Exporters: config.OpenTelemetryExportersConfig{
			Stdout: config.OpenTelemetryExporterStdoutConfig{
				Enable:    true,
				Quantiles: []float64{-1},
			},
		},
	}, injection{})
	assert.Regexp(t, `failed to initialize OpenTelemetry exporters: failed to initialize OpenTelemetry stdout trace exporter: the requested quantile is out of range`, err.Error())
}

func TestGCPExporter(t *testing.T) {
	_, err := newExporters(&config.OpenTelemetryConfig{
		Exporters: config.OpenTelemetryExportersConfig{
			GCP: config.OpenTelemetryExporterGCPConfig{
				EnableTrace: true,
				// Without GOOGLE_APPLICATION_CREDENTIALS, it fails to find default project ID due to lack of default credentials.
				// https://cloud.google.com/trace/docs/setup/go-ot
				// https://cloud.google.com/docs/authentication/production
				ProjectID: "",
			},
		},
	})
	assert.Regexp(t, `failed to initialize OpenTelemetry GCP Cloud Trace exporter.+stackdriver`, err.Error())
}

func TestGlobalErrorHandler(t *testing.T) {
	exporter := ottesting.NewStubExporter(t)
	tracerProvider := otsdktrace.NewTracerProvider(otsdktrace.WithSyncer(exporter))
	tracer := tracerProvider.Tracer("test")

	ctx, span := tracer.Start(context.Background(), "test span")
	ctx = logger.PinLoggerContext(ctx)
	logger.Of(ctx).Error("test error message 1", errors.New("test error 1"))
	logger.Of(ctx).Error("test error message 2", nil)
	span.End()

	assert.NoError(t, tracerProvider.Shutdown(context.Background()))

	event1 := exporter.AssertSpan(0, trace.SpanKindInternal, "test span", map[string]interface{}{}).MessageEvents[0]
	assert.Equal(t, "error", event1.Name)
	assert.EqualValues(t, map[string]interface{}{
		"message":       "test error message 1",
		"error.type":    "*errors.errorString",
		"error.message": "test error 1",
	}, ottesting.KVPairsToMap(event1.Attributes))

	event2 := exporter.AssertSpan(0, trace.SpanKindInternal, "test span", map[string]interface{}{}).MessageEvents[1]
	assert.Equal(t, "error", event2.Name)
	assert.EqualValues(t, map[string]interface{}{
		"message": "test error message 2",
	}, ottesting.KVPairsToMap(event2.Attributes))
}

func TestOtelErrorHandler(t *testing.T) {
	otel.Handle(nil)
	otel.Handle(errors.New("test error"))
}
