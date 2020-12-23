package logger

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOf(t *testing.T) {
	WithTestLogger(t, nil, func(lc *LogCapture) {
		rootCtx := context.Background()
		assert.Same(t, Of(rootCtx), Of(rootCtx))

		transparentCtx := context.WithValue(rootCtx, struct{}{}, true)
		assert.Same(t, Of(rootCtx), Of(transparentCtx))
		assert.NotSame(t, transparentCtx, of(transparentCtx).ctx)

		pinnedCtx := PinLoggerContext(transparentCtx)
		assert.Same(t, pinnedCtx, of(pinnedCtx).ctx)

		lv1Ctx := WithAttributes(rootCtx).WithStr("lv1", "lv1 value").WithInt("lv1int", 123).WithInt64("lv1int64", 9223372036854775807).WithBool("lv1bool", true).Build()
		assert.NotSame(t, Of(rootCtx), Of(lv1Ctx))
		assert.Same(t, Of(lv1Ctx), Of(lv1Ctx))
		assert.Same(t, lv1Ctx, of(lv1Ctx).ctx)

		Of(lv1Ctx).Infof(CatServer, "test info log")
		log := lc.LastLog(0)
		assert.Equal(t, "test info log", log.Message)
		assert.Equal(t, "lv1 value", log.FindStringField("lv1"))
		assert.Nil(t, log.FindErrorField())

		Of(rootCtx).Infof(CatServer, "test info log")
		log = lc.LastLog(0)
		assert.Equal(t, "test info log", log.Message)
		assert.Equal(t, "", log.FindStringField("lv1"))
	})
}
