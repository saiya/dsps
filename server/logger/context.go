package logger

import (
	"context"

	"go.uber.org/zap"
)

type contextKey int

const (
	loggerContextKey contextKey = iota
)

// Of returns or creates Logger instance associated to the context.
func Of(ctx context.Context) Logger {
	return of(ctx)
}

func of(ctx context.Context) *loggerImpl {
	if logger, ok := ctx.Value(loggerContextKey).(*loggerImpl); ok {
		return logger
	}
	return rootLogger
}

// WithAttributes returns builder to create child context that holds child logger
func WithAttributes(ctx context.Context) ContextLoggerBuilder {
	return &contextLoggerBuilder{
		ctx:        ctx,
		baseLogger: of(ctx),
		fields:     make([]zap.Field, 0, 16),
	}
}

// ContextLoggerBuilder is an interface to create child context that holds child logger
type ContextLoggerBuilder interface {
	Build() context.Context

	WithStr(key string, value string) ContextLoggerBuilder
	WithInt(key string, value int) ContextLoggerBuilder
	WithBool(key string, value bool) ContextLoggerBuilder
}

type contextLoggerBuilder struct {
	ctx        context.Context
	baseLogger *loggerImpl
	fields     []zap.Field
}

func (b *contextLoggerBuilder) Build() context.Context {
	return context.WithValue(b.ctx, loggerContextKey, b.baseLogger.WithAttributes(b.fields))
}

func (b *contextLoggerBuilder) WithStr(key string, value string) ContextLoggerBuilder {
	b.fields = append(b.fields, zap.String(key, value))
	return b
}

func (b *contextLoggerBuilder) WithInt(key string, value int) ContextLoggerBuilder {
	b.fields = append(b.fields, zap.Int(key, value))
	return b
}

func (b *contextLoggerBuilder) WithBool(key string, value bool) ContextLoggerBuilder {
	b.fields = append(b.fields, zap.Bool(key, value))
	return b
}
