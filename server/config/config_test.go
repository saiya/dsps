package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/saiya/dsps/server/config"
)

func TestTabInYaml(t *testing.T) {
	configYaml := `
logging:
	# Here is hard TAB
	debug: true
`
	_, err := ParseConfig(Overrides{}, configYaml)
	assert.EqualError(t, err, "Configuration file could not contain tab character (0x09) because YAML spec forbit it, use space to indent")
}
