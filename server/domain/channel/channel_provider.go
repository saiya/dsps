package channel

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/telemetry"
)

// NewChannelProvider initializes ChannelProvider
func NewChannelProvider(ctx context.Context, config *config.ServerConfig, clock domain.SystemClock, telemetry *telemetry.Telemetry) (domain.ChannelProvider, error) {
	atoms := make([]*channelAtom, 0, len(config.Channels))
	for i := range config.Channels {
		atom, err := newChannelAtom(ctx, &config.Channels[i], clock, telemetry, true)
		if err != nil {
			return nil, xerrors.Errorf("channels[%d] configuration error: %w", i, err)
		}
		atoms = append(atoms, atom)
	}
	return newCachedChannelProvider(&channelProvider{atoms: atoms}, clock), nil
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
