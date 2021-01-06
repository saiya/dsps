package channel

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/sentry"
	"github.com/saiya/dsps/server/telemetry"
)

// ProviderDeps contains objects required by ChannelProvider
type ProviderDeps struct {
	Clock     domain.SystemClock
	Telemetry *telemetry.Telemetry
	Sentry    sentry.Sentry
}

func (deps ProviderDeps) validateProviderDeps() error {
	if deps.Clock == nil {
		return xerrors.Errorf("invalid ProviderDeps: Clock should not be nil")
	}
	if deps.Telemetry == nil {
		return xerrors.Errorf("invalid ProviderDeps: Telemetry should not be nil")
	}
	if deps.Sentry == nil {
		return xerrors.Errorf("invalid ProviderDeps: Sentry should not be nil")
	}
	return nil
}

// NewChannelProvider initializes ChannelProvider
func NewChannelProvider(ctx context.Context, config *config.ServerConfig, deps ProviderDeps) (domain.ChannelProvider, error) {
	if err := deps.validateProviderDeps(); err != nil {
		return nil, err
	}

	atoms := make([]*channelAtom, 0, len(config.Channels))
	for i := range config.Channels {
		atom, err := newChannelAtom(ctx, &config.Channels[i], deps, true)
		if err != nil {
			return nil, xerrors.Errorf("channels[%d] configuration error: %w", i, err)
		}
		atoms = append(atoms, atom)
	}
	return newCachedChannelProvider(&channelProvider{atoms: atoms}, deps.Clock), nil
}

type channelProvider struct {
	atoms []*channelAtom
}

func (cp *channelProvider) GetFileDescriptorPressure() int {
	result := 0
	for _, atom := range cp.atoms {
		result += atom.GetFileDescriptorPressure()
	}
	return result
}

func (cp *channelProvider) JWTClockSkewLeewayMax() domain.Duration {
	var result domain.Duration
	for _, atom := range cp.atoms {
		d := atom.JWTClockSkewLeewayMax()
		if result.Duration < d.Duration {
			result = d
		}
	}
	return result
}

func (cp *channelProvider) Get(id domain.ChannelID) (domain.Channel, error) {
	found := make([]*channelAtom, 0, 4)
	for _, atom := range cp.atoms {
		if atom.IsMatch(id) {
			found = append(found, atom)
		}
	}
	if len(found) == 0 {
		return nil, domain.ErrInvalidChannel
	}
	return newChannelImpl(id, found)
}

func (cp *channelProvider) Shutdown(ctx context.Context) {
	for _, atom := range cp.atoms {
		atom.Shutdown(ctx)
	}
}
