package config_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/saiya/dsps/server/config"
	. "github.com/saiya/dsps/server/testing"
)

func TestChannelDefaultConfig(t *testing.T) {
	configYaml := strings.ReplaceAll(`
channels:
-
	regex: 'chat-room-(?P<id>\d+)'
`, "\t", "  ")
	config, err := ParseConfig(context.Background(), Overrides{}, configYaml)
	if err != nil {
		t.Error(err)
		return
	}

	cfg := config.Channels[0]
	assert.Equal(t, "chat-room-(?P<id>\\d+)", cfg.Regex.String())
	assert.Equal(t, MakeDurationPtr("30m"), cfg.Expire)

	assert.Equal(t, 0, len(cfg.Webhooks))
	assert.Nil(t, cfg.Jwt)
}

func TestChannelNonDefaultConfig(t *testing.T) {
	configYaml := strings.ReplaceAll(`
channels:
-
	regex: 'chat-room-(?P<id>\d+)'
	# Must be larger than final retry attempt time
	expire: 15m
`, "\t", "  ")
	config, err := ParseConfig(context.Background(), Overrides{}, configYaml)
	if err != nil {
		t.Error(err)
		return
	}

	cfg := config.Channels[0]
	assert.Equal(t, "chat-room-(?P<id>\\d+)", cfg.Regex.String())
	assert.Equal(t, MakeDurationPtr("15m"), cfg.Expire)
}
