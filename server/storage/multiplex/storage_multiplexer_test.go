package multiplex_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/domain"
	. "github.com/saiya/dsps/server/storage/multiplex"
	"github.com/saiya/dsps/server/storage/onmemory"
	. "github.com/saiya/dsps/server/storage/testing"
)

var onmemoryMultiplexCtor = func(onmemConfigs ...config.OnmemoryStorageConfig) StorageCtor {
	return func(ctx context.Context, systemClock domain.SystemClock, channelProvider domain.ChannelProvider) (domain.Storage, error) {
		storages := map[domain.StorageID]domain.Storage{}
		for i := range onmemConfigs {
			storage, err := onmemory.NewOnmemoryStorage(context.Background(), &(onmemConfigs[i]), systemClock, channelProvider)
			if err != nil {
				return nil, err
			}
			storages[domain.StorageID(fmt.Sprintf("storage%d", i+1))] = storage
		}
		return NewStorageMultiplexer(storages)
	}
}

func TestCoreFunction(t *testing.T) {
	CoreFunctionTest(t, onmemoryMultiplexCtor(
		config.OnmemoryStorageConfig{
			DisablePubSub: true,
			DisableJwt:    true,
		},
		config.OnmemoryStorageConfig{
			DisablePubSub: true,
			DisableJwt:    true,
		},
	))
}

func TestPubSub(t *testing.T) {
	PubSubTest(t, onmemoryMultiplexCtor(
		config.OnmemoryStorageConfig{
			DisableJwt: true,
		},
		config.OnmemoryStorageConfig{
			DisablePubSub: true, // Storage without feature support
			DisableJwt:    true,
		},
		config.OnmemoryStorageConfig{
			DisableJwt: true,
		},
	))
}

func TestJwt(t *testing.T) {
	JwtTest(t, onmemoryMultiplexCtor(
		config.OnmemoryStorageConfig{
			DisablePubSub: true,
		},
		config.OnmemoryStorageConfig{
			DisablePubSub: true,
			DisableJwt:    true, // Storage without feature support
		},
		config.OnmemoryStorageConfig{
			DisablePubSub: true,
		},
	))
}

func TestInsufficientStorages(t *testing.T) {
	ctx := context.Background()
	_, err := NewStorageMultiplexer(map[domain.StorageID]domain.Storage{})
	assert.EqualError(t, err, "List of storages must not be empty")

	pubSubDisabledCfg := config.OnmemoryStorageConfig{
		DisablePubSub: true,
	}
	pubSubDisabled1, err := onmemory.NewOnmemoryStorage(ctx, &pubSubDisabledCfg, domain.RealSystemClock, StubChannelProvider)
	assert.NoError(t, err)
	pubSubDisabled2, err := onmemory.NewOnmemoryStorage(ctx, &pubSubDisabledCfg, domain.RealSystemClock, StubChannelProvider)
	assert.NoError(t, err)
	multiWithoutPubSub, err := NewStorageMultiplexer(map[domain.StorageID]domain.Storage{
		"test1": pubSubDisabled1,
		"test2": pubSubDisabled2,
	})
	assert.NoError(t, err)
	assert.Nil(t, multiWithoutPubSub.AsPubSubStorage())
	assert.NotNil(t, multiWithoutPubSub.AsJwtStorage())

	jwtDisabledCfg := config.OnmemoryStorageConfig{
		DisableJwt: true,
	}
	jwtDisabled1, err := onmemory.NewOnmemoryStorage(ctx, &jwtDisabledCfg, domain.RealSystemClock, StubChannelProvider)
	assert.NoError(t, err)
	jwtDisabled2, err := onmemory.NewOnmemoryStorage(ctx, &jwtDisabledCfg, domain.RealSystemClock, StubChannelProvider)
	assert.NoError(t, err)
	multiWithoutJwt, err := NewStorageMultiplexer(map[domain.StorageID]domain.Storage{
		"test1": jwtDisabled1,
		"test2": jwtDisabled2,
	})
	assert.NoError(t, err)
	assert.NotNil(t, multiWithoutJwt.AsPubSubStorage())
	assert.Nil(t, multiWithoutJwt.AsJwtStorage())
}

// TODO: AckHandle 発行後に storage が増える・減るケースのテスト
// TODO: IsOldMessages は全 storage がエラーでも成功することのテスト

// TODO: 複数の storage が全て ErrSubscriptionNotFound を返したときだけ ErrSubscriptionNotFound になり、片方に存在すれば成功することの確認
// TODO: 一部の storage だけで ErrSubscriptionNotFound が発生した場合、自動で subscription を再作成して復旧され、後に subscription が合った方の storage がエラーになっても subscription を利用継続できることの確認
