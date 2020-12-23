package telemetry

import (
	"context"

	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/logger"
	"github.com/saiya/dsps/server/telemetry/opentelemetry"
)

// Telemetry represents tracing/metrics system
type Telemetry struct {
	ot *opentelemetry.OTFacility
}

// InitTelemetry initialize telemetry facility
func InitTelemetry(config *config.TelemetryConfig) (telemetry *Telemetry, err error) {
	telemetry = &Telemetry{}
	if telemetry.ot, err = opentelemetry.NewOTFacility(config.OT); err != nil {
		return
	}
	return
}

// Shutdown closes telemetry facility
func (t *Telemetry) Shutdown(ctx context.Context) {
	if t.ot != nil {
		if err := t.ot.Shutdown(ctx); err != nil {
			logger.Of(ctx).WarnError(logger.CatTracing, "failed to shutdown OpenTelemetry facilities: %w", err)
		}
	}
}
