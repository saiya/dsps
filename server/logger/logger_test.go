package logger_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"

	. "github.com/saiya/dsps/server/logger"
)

func TestLoggerMethods(t *testing.T) {
	logAllFilter, err := NewFilter(map[string]string{"*": "DEBUG"})
	assert.NoError(t, err)
	WithTestLogger(t, logAllFilter, func(lc *LogCapture) {
		logger := Of(context.Background())

		logger.Error(`error msg`, errors.New(`error error`))
		assert.Equal(t, zapcore.ErrorLevel, lc.LastLog(0).Level)
		assert.Regexp(t, `error msg`, lc.LastLog(0).Message)
		assert.Equal(t, errors.New(`error error`), lc.LastLog(0).FindErrorField())

		logger.WarnError(CatAuth, `warn msg`, errors.New(`warn error`))
		assert.Equal(t, zapcore.WarnLevel, lc.LastLog(0).Level)
		assert.Regexp(t, `warn msg`, lc.LastLog(0).Message)
		assert.Equal(t, `auth`, lc.LastLog(0).FindStringField("category"))
		assert.Equal(t, errors.New(`warn error`), lc.LastLog(0).FindErrorField())

		logger.InfoError(CatAuth, `info msg`, errors.New(`info error`))
		assert.Equal(t, zapcore.InfoLevel, lc.LastLog(0).Level)
		assert.Regexp(t, `info msg`, lc.LastLog(0).Message)
		assert.Equal(t, `auth`, lc.LastLog(0).FindStringField("category"))
		assert.Equal(t, errors.New(`info error`), lc.LastLog(0).FindErrorField())

		logger.Warnf(CatHTTP, `warn log: %s, %d`, `arg1`, 1234)
		assert.Equal(t, zapcore.WarnLevel, lc.LastLog(0).Level)
		assert.Regexp(t, `warn log: arg1, 1234`, lc.LastLog(0).Message)
		assert.Equal(t, `http`, lc.LastLog(0).FindStringField("category"))

		logger.Infof(CatHTTP, `info log: %s, %d`, `arg1`, 1234)
		assert.Equal(t, zapcore.InfoLevel, lc.LastLog(0).Level)
		assert.Regexp(t, `info log: arg1, 1234`, lc.LastLog(0).Message)
		assert.Equal(t, `http`, lc.LastLog(0).FindStringField("category"))

		logger.Debugf(CatHTTP, `debug log: %s, %d`, `arg1`, 1234)
		assert.Equal(t, zapcore.DebugLevel, lc.LastLog(0).Level)
		assert.Regexp(t, `debug log: arg1, 1234`, lc.LastLog(0).Message)
		assert.Equal(t, `http`, lc.LastLog(0).FindStringField("category"))
	})

	logNothingFilter, err := NewFilter(map[string]string{"*": "FATAL"})
	assert.NoError(t, err)
	WithTestLogger(t, logNothingFilter, func(lc *LogCapture) {
		logger := Of(context.Background())

		logger.Error(`error msg`, errors.New(`error error`))
		logger.WarnError(CatAuth, `warn msg`, errors.New(`warn error`))
		logger.InfoError(CatAuth, `info msg`, errors.New(`info error`))
		logger.Warnf(CatHTTP, `warn log: %s, %d`, `arg1`, 1234)
		logger.Infof(CatHTTP, `info log: %s, %d`, `arg1`, 1234)
		logger.Debugf(CatHTTP, `debug log: %s, %d`, `arg1`, 1234)

		assert.Equal(t, 0, lc.Count())
	})
}
