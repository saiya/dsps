package logger

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Due to gin.Context, cannot use non-string (unique) key.
const loggerContextKey = "github.com/saiya/dsps/server/logger"

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
		modify:     false,
		baseLogger: of(ctx),
		fields:     make([]zap.Field, 0, 16),
	}
}

// ModifyGinContext updated logger of the given gin.Context (not to create new Context)
func ModifyGinContext(ctx *gin.Context) ContextLoggerBuilder {
	return &contextLoggerBuilder{
		ctx:        ctx,
		modify:     true,
		baseLogger: of(ctx),
		fields:     make([]zap.Field, 0, 16),
	}
}

// ContextLoggerBuilder is an interface to create child context that holds child logger
type ContextLoggerBuilder interface {
	Build() context.Context

	WithStr(key string, value string) ContextLoggerBuilder
	WithInt(key string, value int) ContextLoggerBuilder
	WithInt64(key string, value int64) ContextLoggerBuilder
	WithBool(key string, value bool) ContextLoggerBuilder
}

type contextLoggerBuilder struct {
	ctx    context.Context
	modify bool

	baseLogger *loggerImpl
	fields     []zap.Field
}

func (b *contextLoggerBuilder) Build() context.Context {
	newLogger := b.baseLogger.WithAttributes(b.fields)
	if gctx, ok := b.ctx.(*gin.Context); b.modify && ok {
		// Special handling for gin
		gctx.Set(loggerContextKey, newLogger)
		return gctx
	}
	return context.WithValue(b.ctx, loggerContextKey, newLogger) //nolint:golint,staticcheck
}

func (b *contextLoggerBuilder) WithStr(key string, value string) ContextLoggerBuilder {
	b.fields = append(b.fields, zap.String(key, value))
	return b
}

func (b *contextLoggerBuilder) WithInt(key string, value int) ContextLoggerBuilder {
	b.fields = append(b.fields, zap.Int(key, value))
	return b
}

func (b *contextLoggerBuilder) WithInt64(key string, value int64) ContextLoggerBuilder {
	b.fields = append(b.fields, zap.Int64(key, value))
	return b
}

func (b *contextLoggerBuilder) WithBool(key string, value bool) ContextLoggerBuilder {
	b.fields = append(b.fields, zap.Bool(key, value))
	return b
}
