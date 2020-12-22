package opentelemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/label"
	exporttrace "go.opentelemetry.io/otel/sdk/export/trace"
	ottrace "go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"

	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/logger"
)

// OTFacility holds OT related objects
type OTFacility struct {
	Tracing *Tracing

	spanExporters []exporttrace.SpanExporter

	TestShutdownError error // testing purpose only
}

type injection struct {
	tracing tracingInjection
}

func init() {
	logger.SetGlobalErrorHandler(func(ctx context.Context, msg string, err error) {
		span := ottrace.SpanFromContext(ctx)
		if err != nil {
			span.RecordError(err, ottrace.WithAttributes(label.String("message", msg)))
		} else {
			span.AddEvent("error", ottrace.WithAttributes(label.String("message", msg)))
		}
	})
	otel.SetErrorHandler(otelErrorHandler(func(err error) {
		if err == nil {
			return // Current OT implementation unconditionally passes error to handler
		}

		// Should not use ERROR log level, it could cause infinite loop between logger and tracer.
		logger.Of(context.Background()).WarnError(logger.CatTracing, "error in OpenTelemetry SDK", err)
	}))
}

// NewOTFacility initializes OT facilities
func NewOTFacility(config *config.OpenTelemetryConfig) (facility *OTFacility, err error) {
	return newOTFacility(config, injection{})
}

func newOTFacility(config *config.OpenTelemetryConfig, injection injection) (facility *OTFacility, err error) {
	facility = &OTFacility{}
	if facility.spanExporters, err = newExporters(config); err != nil {
		err = fmt.Errorf("failed to initialize OpenTelemetry exporters: %w", err)
		return
	}
	injection.tracing.exporters = append(injection.tracing.exporters, facility.spanExporters...)
	if facility.Tracing, err = newTracing(config.Tracing, injection.tracing); err != nil {
		err = fmt.Errorf("failed to initialize OpenTelemetry tracer: %w", err)
		return
	}
	return
}

// Shutdown closes OT
func (ot *OTFacility) Shutdown(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return ot.TestShutdownError })
	g.Go(func() error { return ot.Tracing.Shutdown(ctx) })
	return g.Wait()
}

type otelErrorHandler func(error)

func (h otelErrorHandler) Handle(err error) {
	h(err)
}
