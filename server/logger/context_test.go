package logger_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	. "github.com/saiya/dsps/server/logger"
)

func TestOf(t *testing.T) {
	WithTestLogger(t, nil, func(lc *LogCapture) {
		rootCtx := context.Background()
		assert.Same(t, Of(rootCtx), Of(rootCtx))

		transparentCtx := context.WithValue(rootCtx, struct{}{}, true)
		assert.Same(t, Of(rootCtx), Of(transparentCtx))

		lv1Ctx := WithAttributes(rootCtx).WithStr("lv1", "lv1 value").WithInt("lv1int", 123).WithInt64("lv1int64", 9223372036854775807).WithBool("lv1bool", true).Build()
		assert.NotSame(t, Of(rootCtx), Of(lv1Ctx))
		assert.Same(t, Of(lv1Ctx), Of(lv1Ctx))

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

func TestGinContext(t *testing.T) {
	WithTestLogger(t, nil, func(lc *LogCapture) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		ModifyGinContext(ctx).WithStr("attr1", "attr value").Build()

		Of(ctx).Infof(CatServer, "info log")
		assert.Equal(t, "info log", lc.LastLog(0).Message)
		assert.Equal(t, "attr value", lc.LastLog(0).FindStringField("attr1")) // gin.context itself should be changed
	})
}
