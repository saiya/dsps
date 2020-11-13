package config_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/saiya/dsps/server/config"
)

func TestEmptyStorages(t *testing.T) {
	configYaml := strings.ReplaceAll(`
storages:
`, "\t", "  ")
	config, err := ParseConfig(Overrides{}, configYaml)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(config.Storages))
	assert.Equal(t, DefaultStoragesConfig(), config.Storages)
}
