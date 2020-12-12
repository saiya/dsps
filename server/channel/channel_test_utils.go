package channel

import (
	"testing"

	"github.com/saiya/dsps/server/domain"
	"github.com/stretchr/testify/assert"
)

// NewChannelByAtomYamls creates Channel object with given channel configuration YAML fragments.
func NewChannelByAtomYamls(t *testing.T, id domain.ChannelID, yamls []string) domain.Channel {
	atoms := make([]*channelAtom, len(yamls))
	for i, yaml := range yamls {
		atoms[i] = newChannelAtomByYaml(t, yaml, true)
	}
	c, err := newChannelImpl(id, atoms)
	assert.NoError(t, err)
	return c
}
