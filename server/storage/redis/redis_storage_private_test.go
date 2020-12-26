package redis

import (
	"context"
	"os"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/saiya/dsps/server/domain"
	. "github.com/saiya/dsps/server/storage/redis/internal"
	. "github.com/saiya/dsps/server/storage/redis/internal/mock"
	storagetesting "github.com/saiya/dsps/server/storage/testing"
)

func GetRedisAddr(_ *testing.T) string {
	addr := os.Getenv("DSPS_REDIS")
	if addr == "" {
		addr = "127.0.0.1:6379"
	}
	return addr
}

func WithRedisClient(t *testing.T, f func(redisCmd RedisCmd)) {
	client := redis.NewClient(&redis.Options{Addr: GetRedisAddr(t)})
	defer func() { assert.NoError(t, client.Close()) }()

	f(NewRedisCmd(client, func(ctx context.Context, channel RedisChannelID) *redis.PubSub {
		return client.PSubscribe(ctx, string(channel))
	}))
}

func newMockedRedisStorage(ctrl *gomock.Controller) (*redisStorage, *MockRedisCmd) {
	redisCmd := NewMockRedisCmd(ctrl)
	return &redisStorage{
		clock:           domain.RealSystemClock,
		channelProvider: storagetesting.StubChannelProvider,

		pubsubEnabled: true,
		jwtEnabled:    true,

		RedisConnection: RedisConnection{
			RedisCmd:       redisCmd,
			Close:          func() error { return nil },
			IsSingleNode:   false,
			IsCluster:      false,
			MaxConnections: 1024,
		},
	}, redisCmd
}

func TestRedisStorageFeatureFlag(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s, _ := newMockedRedisStorage(ctrl)

	s.pubsubEnabled = true
	assert.Same(t, s, s.AsPubSubStorage())
	s.pubsubEnabled = false
	assert.Nil(t, s.AsPubSubStorage())

	s.jwtEnabled = true
	assert.Same(t, s, s.AsJwtStorage())
	s.jwtEnabled = false
	assert.Nil(t, s.AsJwtStorage())
}

func TestGetFileDescriptorPressure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s, _ := newMockedRedisStorage(ctrl)

	s.MaxConnections = 1234
	assert.Equal(t, 1234, s.GetFileDescriptorPressure())
}
