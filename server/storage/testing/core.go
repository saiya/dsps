package testing

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	domain "github.com/saiya/dsps/server/domain"
)

// CoreFunctionTest tests common Storage behaviors
func CoreFunctionTest(t *testing.T, storageCtor StorageCtor) {
	ctx := context.Background()
	storage, err := storageCtor(ctx, domain.RealSystemClock, StubChannelProvider)
	if !assert.NoError(t, err) {
		return
	}
	defer func() { assert.NoError(t, storage.Shutdown(ctx)) }()

	assert.NotEmpty(t, storage.String())

	if _, err := storage.Liveness(ctx); !assert.NoError(t, err) {
		return
	}
	if _, err := storage.Readiness(ctx); !assert.NoError(t, err) {
		return
	}
}
