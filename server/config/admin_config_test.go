package config_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/domain"
)

func TestAdminDefaultConfig(t *testing.T) {
	config, err := ParseConfig(context.Background(), Overrides{}, `admin: { auth: {} }`)
	if err != nil {
		t.Error(err)
		return
	}

	cfg := config.Admin
	assert.Equal(t, len(domain.PrivateCIDRs), len(cfg.Auth.Networks))
	assert.Equal(t, domain.PrivateCIDRs[0].String(), cfg.Auth.Networks[0].String())
	assert.Equal(t, 1, len(cfg.Auth.BearerTokens))
	assert.NotEmpty(t, cfg.Auth.BearerTokens[0])
}

func TestAdminNonDefaultConfig(t *testing.T) {
	configYaml := strings.ReplaceAll(`
admin:
	auth:
		networks:
			- 10.1.2.0/8
			- 11.1.2.0/8
		bearer:
			- 'my-api-key1'
			- 'my-api-key2'
`, "\t", "  ")
	config, err := ParseConfig(context.Background(), Overrides{}, configYaml)
	if err != nil {
		t.Error(err)
		return
	}

	cfg := config.Admin
	assert.Equal(t, 2, len(cfg.Auth.Networks))
	assert.Equal(t, "10.1.2.0/8", cfg.Auth.Networks[0].String())
	assert.Equal(t, "11.1.2.0/8", cfg.Auth.Networks[1].String())
	assert.Equal(t, 2, len(cfg.Auth.BearerTokens))
	assert.Equal(t, "my-api-key1", cfg.Auth.BearerTokens[0])
	assert.Equal(t, "my-api-key2", cfg.Auth.BearerTokens[1])
}
