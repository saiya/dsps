package logger

import (
	"fmt"

	"go.uber.org/zap"
)

// Logger represents implementation independent logger, also easy to inject mock.
type Logger interface {
	Fatal(msg string, err error)
	Error(msg string, err error)
	WarnError(cat Category, msg string, err error)
	InfoError(cat Category, msg string, err error)

	Warnf(cat Category, template string, args ...interface{})
	Infof(cat Category, template string, args ...interface{})
	Debugf(cat Category, template string, args ...interface{})
}

type loggerImpl struct {
	zap    *zap.Logger
	filter *Filter
}

func (logger *loggerImpl) WithAttributes(fields []zap.Field) *loggerImpl {
	return &loggerImpl{
		zap:    logger.zap.With(fields...),
		filter: logger.filter,
	}
}

func (logger *loggerImpl) WithFilter(filter *Filter) *loggerImpl {
	return &loggerImpl{
		zap:    logger.zap,
		filter: filter,
	}
}

func (logger *loggerImpl) Fatal(msg string, err error) {
	logger.zap.Fatal(msg, zap.Error(err), zap.String("category", "FATAL"))
}

func (logger *loggerImpl) Error(msg string, err error) {
	logger.zap.Error(msg, zap.Error(err), zap.String("category", "ERROR"))
}

func (logger *loggerImpl) WarnError(cat Category, msg string, err error) {
	if !(logger.filter.Filter(WARN, cat)) {
		return
	}
	logger.zap.Warn(msg, zap.Error(err), zap.String("category", string(cat)))
}

func (logger *loggerImpl) InfoError(cat Category, msg string, err error) {
	if !(logger.filter.Filter(INFO, cat)) {
		return
	}
	logger.zap.Info(msg, zap.Error(err), zap.String("category", string(cat)))
}

func (logger *loggerImpl) Warnf(cat Category, template string, args ...interface{}) {
	if !(logger.filter.Filter(WARN, cat)) {
		return
	}
	logger.zap.Warn(fmt.Sprintf(template, args...), zap.String("category", string(cat)))
}

func (logger *loggerImpl) Infof(cat Category, template string, args ...interface{}) {
	if !(logger.filter.Filter(INFO, cat)) {
		return
	}
	logger.zap.Info(fmt.Sprintf(template, args...), zap.String("category", string(cat)))
}

func (logger *loggerImpl) Debugf(cat Category, template string, args ...interface{}) {
	if !(logger.filter.Filter(DEBUG, cat)) {
		return
	}
	logger.zap.Debug(fmt.Sprintf(template, args...), zap.String("category", string(cat)))
}
