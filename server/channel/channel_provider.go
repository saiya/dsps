package channel

import (
	"golang.org/x/xerrors"

	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/domain"
)

// NewChannelProvider initializes ChannelProvider
func NewChannelProvider(config *config.ServerConfig, clock domain.SystemClock) (domain.ChannelProvider, error) {
	atoms := make([]*channelAtom, 0, len(config.Channels))
	for i := range config.Channels {
		atom, err := newChannelAtom(&config.Channels[i], true)
		if err != nil {
			return nil, xerrors.Errorf("channels[%d] configuration error: %w", i, err)
		}
		atoms = append(atoms, atom)
	}

	return newCachedChannelProvider(func(id domain.ChannelID) domain.Channel {
		found := make([]*channelAtom, 0, 4)
		for _, atom := range atoms {
			if atom.IsMatch(id) {
				found = append(found, atom)
			}
		}
		if len(found) == 0 {
			return nil
		}
		return newChannelImpl(id, found)
	}, clock), nil
}
