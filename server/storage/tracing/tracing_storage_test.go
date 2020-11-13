package tracing_test

import (
	"context"
	"testing"

	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/storage/onmemory"
	. "github.com/saiya/dsps/server/storage/testing"
	. "github.com/saiya/dsps/server/storage/tracing"
)

var onmemoryTracingCtor = func(onmemConfig config.OnmemoryStorageConfig) StorageCtor {
	return func(ctx context.Context, systemClock domain.SystemClock, channelProvider domain.ChannelProvider) (domain.Storage, error) {
		storage, err := onmemory.NewOnmemoryStorage(&onmemConfig, systemClock, channelProvider)
		if err != nil {
			return nil, err
		}
		return NewTracingStorage(storage)
	}
}

func TestCoreFunction(t *testing.T) {
	CoreFunctionTest(t, onmemoryTracingCtor(config.OnmemoryStorageConfig{
		DisableJwt:    true,
		DisablePubSub: true,
	}))
}

func TestPubSub(t *testing.T) {
	PubSubTest(t, onmemoryTracingCtor(config.OnmemoryStorageConfig{
		DisableJwt: true,
	}))
}

func TestJwt(t *testing.T) {
	JwtTest(t, onmemoryTracingCtor(config.OnmemoryStorageConfig{
		DisablePubSub: true,
	}))
}
