package logger_test

import (
	"strings"
	"testing"

	. "github.com/saiya/dsps/server/logger"
	"github.com/stretchr/testify/assert"
)

func TestLevelOrdering(t *testing.T) {
	ordered := []Level{DEBUG, INFO, WARN, ERROR, FATAL}
	for i := range ordered {
		if i == 0 {
			continue
		}
		assert.True(t, ordered[i-1] < ordered[i])
	}
}

func TestLevelStrings(t *testing.T) {
	for str, expected := range map[string]Level{
		"DEBUG": DEBUG,
		"debug": DEBUG, // Should be case insensitive
		"INFO":  INFO,
		"WARN":  WARN,
		"ERROR": ERROR,
		"FATAL": FATAL,
	} {
		actual, err := ParseLevel(str)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Equal(t, strings.ToUpper(str), expected.String())
	}

	_, err := ParseLevel("")
	assert.EqualError(t, err, `invalid log level string given: ""`)
	_, err = ParseLevel("TRACE")
	assert.EqualError(t, err, `invalid log level string given: "TRACE"`)

	assert.Equal(t, "logger.Level(1024)", Level(1024).String())
}
