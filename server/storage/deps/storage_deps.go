package deps

import (
	"github.com/saiya/dsps/server/sentry"
	"github.com/saiya/dsps/server/telemetry"
)

// StorageDeps contains objects required by storage implementations.
type StorageDeps struct {
	Telemetry *telemetry.Telemetry
	Sentry    sentry.Sentry
}
