package channel

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/sentry"
	"github.com/saiya/dsps/server/telemetry"
)

func newChannelAtomByYaml(t *testing.T, yaml string, validate bool) *channelAtom { //nolint:golint
	yaml = fmt.Sprintf("channels:\n  - %s", strings.ReplaceAll(strings.ReplaceAll(yaml, "\t", "  "), "\n", "\n    "))
	cfg, err := config.ParseConfig(context.Background(), config.Overrides{}, yaml)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(cfg.Channels))

	atom, err := newChannelAtom(context.Background(), &cfg.Channels[0], ProviderDeps{
		Clock:     domain.RealSystemClock,
		Telemetry: telemetry.NewEmptyTelemetry(t),
		Sentry:    sentry.NewEmptySentry(),
	}, validate)
	assert.NoError(t, err)
	return atom
}
