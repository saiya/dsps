package multiplex_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/domain"
	. "github.com/saiya/dsps/server/storage/multiplex"
	"github.com/saiya/dsps/server/storage/onmemory"
	. "github.com/saiya/dsps/server/storage/testing"
	"github.com/saiya/dsps/server/telemetry"
)

var onmemoryMultiplexCtor = func(t *testing.T, onmemConfigs ...config.OnmemoryStorageConfig) StorageCtor {
	return func(ctx context.Context, systemClock domain.SystemClock, channelProvider domain.ChannelProvider) (domain.Storage, error) {
		storages := map[domain.StorageID]domain.Storage{}
		for i := range onmemConfigs {
			storage, err := onmemory.NewOnmemoryStorage(context.Background(), &(onmemConfigs[i]), systemClock, channelProvider, telemetry.NewEmptyTelemetry(t))
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
		t,
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
		t,
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
		t,
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
