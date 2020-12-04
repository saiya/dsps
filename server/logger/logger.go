package logger

import (
	"fmt"

	"go.uber.org/zap"
)

// Logger represents implementation independent logger, also easy to inject mock.
type Logger interface {
	Fatal(msg string, err error)
	Error(msg string, err error)
	WarnError(msg string, err error)
	InfoError(msg string, err error)

	Warnf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Debugf(template string, args ...interface{})
}

type loggerImpl struct {
	zap *zap.Logger
}

func (logger *loggerImpl) WithAttributes(fields []zap.Field) *loggerImpl {
	return &loggerImpl{
		zap: logger.zap.With(fields...),
	}
}

func (logger *loggerImpl) Fatal(msg string, err error) {
	logger.zap.Fatal(msg, zap.Error(err))
}

func (logger *loggerImpl) Error(msg string, err error) {
	logger.zap.Error(msg, zap.Error(err))
}

func (logger *loggerImpl) WarnError(msg string, err error) {
	logger.zap.Warn(msg, zap.Error(err))
}

func (logger *loggerImpl) InfoError(msg string, err error) {
	logger.zap.Info(msg, zap.Error(err))
}

func (logger *loggerImpl) Warnf(template string, args ...interface{}) {
	logger.zap.Warn(fmt.Sprintf(template, args...))
}

func (logger *loggerImpl) Infof(template string, args ...interface{}) {
	logger.zap.Info(fmt.Sprintf(template, args...))
}

func (logger *loggerImpl) Debugf(template string, args ...interface{}) {
	logger.zap.Debug(fmt.Sprintf(template, args...))
}
