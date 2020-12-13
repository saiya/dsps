package config_test

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/saiya/dsps/server/config"
	. "github.com/saiya/dsps/server/testing"
)

func TestTabInYaml(t *testing.T) {
	configYaml := `
logging:
	# Here is hard TAB
	category: "*": INFO
`
	_, err := ParseConfig(Overrides{}, configYaml)
	assert.EqualError(t, err, "Configuration file could not contain tab character (0x09) because YAML spec forbit it, use space to indent")
}

func TestLoadConfigFile(t *testing.T) {
	configYaml := strings.ReplaceAll(`
logging:
	category: "*": DEBUG
`, "\t", "  ")

	// Default config
	cfg, err := LoadConfigFile("", Overrides{})
	assert.NoError(t, err)
	assert.Equal(t, "", cfg.Logging.Category["*"])

	// Read from file
	WithTextFile(t, configYaml, func(filename string) {
		cfg, err := LoadConfigFile(filename, Overrides{})
		assert.NoError(t, err)
		assert.Equal(t, "DEBUG", cfg.Logging.Category["*"])
	})

	// Read from stdin
	WithTextFile(t, configYaml, func(filename string) {
		realStdin := os.Stdin
		defer func() { os.Stdin = realStdin }()
		stdin, err := os.OpenFile(filename, os.O_RDONLY, os.ModePerm) //nolint:gosec
		assert.NoError(t, err)
		os.Stdin = stdin

		cfg, err := LoadConfigFile("-", Overrides{})
		assert.NoError(t, err)
		assert.Equal(t, "DEBUG", cfg.Logging.Category["*"])
	})

	// Invalid config
	WithTextFile(t, `xxx: {}`, func(filename string) {
		_, err := LoadConfigFile(filename, Overrides{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), `unknown field "xxx"`)
	})
}

func TestDumpConfig(t *testing.T) {
	rountTripTest := func(yaml string, validator func(*ServerConfig)) {
		cfg, err := ParseConfig(Overrides{}, yaml)
		assert.NoError(t, err)
		validator(&cfg)

		dump := strings.Builder{}
		assert.NoError(t, cfg.DumpConfig(&dump))

		cfg, err = ParseConfig(Overrides{}, dump.String())
		assert.NoError(t, err)
		validator(&cfg)
	}

	rountTripTest(strings.ReplaceAll(`
	logging:
		category: "*": "DEBUG"
	`, "\t", "  "), func(cfg *ServerConfig) {
		assert.Equal(t, "DEBUG", cfg.Logging.Category["*"])
	})
}
