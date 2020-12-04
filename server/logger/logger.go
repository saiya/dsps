package logger

import (
	"fmt"

	"go.uber.org/zap"
)

// Logger represents implementation independent logger, also easy to inject mock.
type Logger interface {
	Fatalf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
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

func (logger *loggerImpl) Fatalf(template string, args ...interface{}) {
	logger.zap.Fatal(fmt.Sprintf(template, args...))
}

func (logger *loggerImpl) Errorf(template string, args ...interface{}) {
	logger.zap.Error(fmt.Sprintf(template, args...))
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
