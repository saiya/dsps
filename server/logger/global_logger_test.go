package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/saiya/dsps/server/config"
)

func TestInitLogger(t *testing.T) {
	old := rootLogger
	defer func() { rootLogger = old }()

	initImpl() // Ensure default
	assert.False(t, rootLogger.filter.Filter(DEBUG, CatAuth))

	assert.NoError(t, InitLogger(&config.LoggingConfig{
		Category: map[string]string{
			CatAuth: "DEBUG",
		},
		Attributes: map[string]string{
			"tag1": "value 1",
		},
	}))
	assert.NotNil(t, rootLogger)
	assert.True(t, rootLogger.filter.Filter(DEBUG, CatAuth))

	assert.Regexp(t, `invalid log level string given: "INVALID-LEVEL"`, InitLogger(&config.LoggingConfig{
		Category: map[string]string{
			CatAuth: "INVALID-LEVEL",
		},
	}).Error())
}
