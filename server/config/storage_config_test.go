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

func TestStorageConfigError(t *testing.T) {
	_, err := ParseConfig(Overrides{}, `storages: test: {} } ]`)
	assert.Regexp(t, `there is a configuration error on storage\[test\]: no storage type under the item`, err.Error())

	_, err = ParseConfig(Overrides{}, `storages: test: { onmemory: {}, redis: { singleNode: "localhost:0000" } } ]`)
	assert.Regexp(t, `there is a configuration error on storage\[test\]: found multiple storage type under single item. To configure multiple storages, write separate storage definitions`, err.Error())
}
