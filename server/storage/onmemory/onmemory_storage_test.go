package onmemory_test

import (
	"context"
	"testing"

	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/domain"
	. "github.com/saiya/dsps/server/storage/onmemory"
	. "github.com/saiya/dsps/server/storage/testing"
	"github.com/stretchr/testify/assert"
)

var storageCtor StorageCtor = func(ctx context.Context, systemClock domain.SystemClock, channelProvider domain.ChannelProvider) (domain.Storage, error) {
	config := config.OnmemoryStorageConfig{
		RunGCOnShutdown: true,
	}
	return NewOnmemoryStorage(context.Background(), &config, systemClock, channelProvider)
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

func TestFeatureFlags(t *testing.T) {
	s, err := NewOnmemoryStorage(context.Background(), &config.OnmemoryStorageConfig{
		DisablePubSub: true,
		DisableJwt:    true,
	}, domain.RealSystemClock, StubChannelProvider)
	assert.NoError(t, err)
	assert.Nil(t, s.AsPubSubStorage())
	assert.Nil(t, s.AsJwtStorage())

	s, err = NewOnmemoryStorage(context.Background(), &config.OnmemoryStorageConfig{
		DisablePubSub: false,
		DisableJwt:    false,
	}, domain.RealSystemClock, StubChannelProvider)
	assert.NoError(t, err)
	assert.Same(t, s, s.AsPubSubStorage())
	assert.Same(t, s, s.AsJwtStorage())
}

func TestGetFileDescriptorPressure(t *testing.T) {
	s, err := NewOnmemoryStorage(context.Background(), &config.OnmemoryStorageConfig{}, domain.RealSystemClock, StubChannelProvider)
	assert.NoError(t, err)
	assert.Equal(t, 0, s.GetFileDescriptorPressure())
}
