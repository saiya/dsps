package testing

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	domain "github.com/saiya/dsps/server/domain"
)

// StorageCtor should construct Storage instance to test
type StorageCtor func(ctx context.Context, systemClock domain.SystemClock, channelProvider domain.ChannelProvider) (domain.Storage, error)

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

	data, err := storage.Stat(ctx)
	if assert.NoError(t, err) {
		if _, err := json.Marshal(data); !assert.NoError(t, err) {
			return
		}
	}
}
