package channel

import (
	"github.com/saiya/dsps/server/domain"
)

type channelImpl struct {
	id    domain.ChannelID
	atoms []*channelAtom

	expire domain.Duration
}

func (c *channelImpl) Expire() domain.Duration {
	return c.expire
}

func newChannelImpl(id domain.ChannelID, atoms []*channelAtom) *channelImpl {
	c := &channelImpl{
		id:    id,
		atoms: atoms,

		expire: atoms[0].Expire(),
	}
	for _, atom := range atoms {
		if c.expire.Duration < atom.Expire().Duration {
			c.expire = atom.Expire()
		}
	}
	return c
}
