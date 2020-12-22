package opentelemetry

import (
	"context"
	"errors"
	"testing"

	ottesting "github.com/saiya/dsps/server/telemetry/opentelemetry/testing"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	otsdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"github.com/saiya/dsps/server/logger"
)

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
