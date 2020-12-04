package channel

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/saiya/dsps/server/config"
)

func newChannelAtomByYaml(t *testing.T, yaml string, validate bool) *channelAtom { //nolint:golint
	yaml = fmt.Sprintf("channels:\n  - %s", strings.ReplaceAll(strings.ReplaceAll(yaml, "\t", "  "), "\n", "\n    "))
	cfg, err := config.ParseConfig(config.Overrides{}, yaml)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(cfg.Channels))

	atom, err := newChannelAtom(&cfg.Channels[0], validate)
	assert.NoError(t, err)
	return atom
}
