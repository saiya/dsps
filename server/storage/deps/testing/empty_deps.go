package testing

import (
	"testing"

	"github.com/saiya/dsps/server/sentry"
	"github.com/saiya/dsps/server/storage/deps"
	"github.com/saiya/dsps/server/telemetry"
)

// EmptyDeps fills stub objects
func EmptyDeps(t *testing.T) deps.StorageDeps {
	return deps.StorageDeps{
		Telemetry: telemetry.NewEmptyTelemetry(t),
		Sentry:    sentry.NewEmptySentry(),
	}
}
