package logger_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/saiya/dsps/server/logger"
)

func TestFilter(t *testing.T) {
	filter, err := NewFilter(map[string]string{
		"*":    "WARN",
		"auth": "INFO",
		"http": "ERROR",
	})
	assert.NoError(t, err)

	// default threshold
	assert.True(t, filter.Filter(WARN, "any"))
	assert.False(t, filter.Filter(INFO, "any"))

	// category specific threshold
	assert.True(t, filter.Filter(INFO, "auth"))
	assert.False(t, filter.Filter(DEBUG, "auth"))
	assert.True(t, filter.Filter(ERROR, "http"))
	assert.False(t, filter.Filter(WARN, "http"))

	// Changing threshold
	filter.SetThreshold("http", INFO)
	assert.True(t, filter.Filter(INFO, "http"))
	assert.False(t, filter.Filter(DEBUG, "http"))
}

func TestInvalidFilterDefinition(t *testing.T) {
	_, err := NewFilter(map[string]string{"auth": "INVALID_LEVEL"})
	assert.EqualError(t, err, `invalid log level string given: "INVALID_LEVEL"`)
}
