package onmemory_test

import (
	"context"
	"testing"

	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/domain"
	. "github.com/saiya/dsps/server/storage/onmemory"
	. "github.com/saiya/dsps/server/storage/testing"
)

var storageCtor StorageCtor = func(ctx context.Context, systemClock domain.SystemClock, channelProvider domain.ChannelProvider) (domain.Storage, error) {
	config := config.OnmemoryStorageConfig{
		RunGCOnShutdown: true,
	}
	return NewOnmemoryStorage(&config, systemClock, channelProvider)
}

func TestCoreFunction(t *testing.T) {
	CoreFunctionTest(t, storageCtor)
}

func TestPubSub(t *testing.T) {
	PubSubTest(t, storageCtor)
}

func TestJwt(t *testing.T) {
	JwtTest(t, storageCtor)
}
