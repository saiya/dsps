package channel

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/sentry"
	"github.com/saiya/dsps/server/telemetry"
	dspstesting "github.com/saiya/dsps/server/testing"
)

func TestProvider(t *testing.T) {
	cfg, err := config.ParseConfig(context.Background(), config.Overrides{}, `channels: [ { regex: "test.+", expire: "1s" } ]`)
	assert.NoError(t, err)
	clock := dspstesting.NewStubClock(t)
	cp, err := NewChannelProvider(context.Background(), &cfg, ProviderDeps{
		Clock:     clock,
		Telemetry: telemetry.NewEmptyTelemetry(t),
		Sentry:    sentry.NewEmptySentry(),
	})
	assert.NoError(t, err)

	test1, err := cp.Get("test1")
	assert.NoError(t, err)
	test1Again, err := cp.Get("test1")
	assert.NoError(t, err)
	assert.NotNil(t, test1)
	assert.Same(t, test1, test1Again)

	notfound, err := cp.Get("not-found")
	assert.Nil(t, notfound)
	dspstesting.IsError(t, domain.ErrInvalidChannel, err)
}

func TestProviderWithInvalidDeps(t *testing.T) {
	_, err := NewChannelProvider(context.Background(), nil, ProviderDeps{})
	assert.Regexp(t, `invalid ProviderDeps`, err.Error())
}

func TestValidateProviderDeps(t *testing.T) {
	valid := ProviderDeps{
		Clock:     domain.RealSystemClock,
		Telemetry: telemetry.NewEmptyTelemetry(t),
		Sentry:    sentry.NewEmptySentry(),
	}
	assert.NoError(t, valid.validateProviderDeps())

	invalid := valid
	invalid.Clock = nil
	assert.Regexp(t, `invalid ProviderDeps: Clock should not be nil`, invalid.validateProviderDeps())

	invalid = valid
	invalid.Telemetry = nil
	assert.Regexp(t, `invalid ProviderDeps: Telemetry should not be nil`, invalid.validateProviderDeps())

	invalid = valid
	invalid.Sentry = nil
	assert.Regexp(t, `invalid ProviderDeps: Sentry should not be nil`, invalid.validateProviderDeps())
}
