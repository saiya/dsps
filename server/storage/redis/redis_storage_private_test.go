package redis

import (
	"context"
	"os"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/storage/redis/mock"
	storagetesting "github.com/saiya/dsps/server/storage/testing"
)

func GetRedisAddr(_ *testing.T) string {
	addr := os.Getenv("DSPS_REDIS")
	if addr == "" {
		addr = "127.0.0.1:6379"
	}
	return addr
}

func WithRedisClient(t *testing.T, f func(redisCmd redisCmd)) {
	client := redis.NewClient(&redis.Options{Addr: GetRedisAddr(t)})
	defer func() { assert.NoError(t, client.Close()) }()

	f(newRedisCmd(client, func(ctx context.Context, channel string) *redis.PubSub {
		return client.Subscribe(ctx, channel)
	}))
}

func newMockedRedisStorage(ctrl *gomock.Controller) (*redisStorage, *mock.MockredisCmd) {
	redisCmd := mock.NewMockredisCmd(ctrl)
	return &redisStorage{
		clock:           domain.RealSystemClock,
		channelProvider: storagetesting.StubChannelProvider,

		pubsubEnabled: false,
		jwtEnabled:    false,

		redisConnection: redisConnection{
			redisCmd:       redisCmd,
			close:          func() error { return nil },
			isSingleNode:   false,
			isCluster:      false,
			maxConnections: 1024,
		},
	}, redisCmd
}
