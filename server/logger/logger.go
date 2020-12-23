package logger

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// Logger represents implementation independent logger, also easy to inject mock.
type Logger interface {
	// output fatal log then exit this process immediately
	FatalExitProcess(msg string, err error)

	Error(msg string, err error)
	WarnError(cat Category, msg string, err error)
	InfoError(cat Category, msg string, err error)

	Warnf(cat Category, template string, args ...interface{})
	Infof(cat Category, template string, args ...interface{})
	Debugf(cat Category, template string, args ...interface{})
}

type loggerImpl struct {
	ctx context.Context
	zap *zap.Logger

	// Log filter, must point same instance across all logger instances belongs to same tree.
	filter *Filter
}

func (logger *loggerImpl) copy() *loggerImpl {
	return &loggerImpl{
		ctx:    logger.ctx,
		zap:    logger.zap,
		filter: logger.filter,
	}
}

func (logger *loggerImpl) withAttributes(fields []zap.Field) *loggerImpl {
	result := logger.copy()
	result.zap = result.zap.With(fields...)
	return result
}

func (logger *loggerImpl) withFilter(filter *Filter) *loggerImpl {
	result := logger.copy()
	result.filter = filter
	return result
}

func (logger *loggerImpl) FatalExitProcess(msg string, err error) {
	logger.zap.Fatal(msg, zap.Error(err), zap.String("category", "FATAL"))
}

func (logger *loggerImpl) Error(msg string, err error) {
	if !(logger.filter.Filter(ERROR, "ERROR")) {
		return
	}
	logger.zap.Error(msg, zap.Error(err), zap.String("category", "ERROR"))
	invokeGlobalErrorHandlers(logger.ctx, msg, err)
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
