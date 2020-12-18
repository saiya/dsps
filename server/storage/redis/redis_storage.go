package redis

import (
	"context"
	"time"

	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/domain"
)

// In case of clock drift
const ttlMargin = 15 * time.Second

// NewRedisStorage creates Storage instance
func NewRedisStorage(ctx context.Context, config *config.RedisStorageConfig, systemClock domain.SystemClock, channelProvider domain.ChannelProvider) (domain.Storage, error) {
	conn, err := connect(ctx, config)
	if err != nil {
		return &redisStorage{}, err
	}
	s := &redisStorage{
		clock:           systemClock,
		channelProvider: channelProvider,

		pubsubEnabled: !config.DisablePubSub,
		jwtEnabled:    !config.DisableJwt,

		redisConnection: conn,
	}
	if err := s.loadScripts(ctx); err != nil {
		return nil, err
	}
	return s, nil
}

type redisStorage struct {
	clock           domain.SystemClock
	channelProvider domain.ChannelProvider

	pubsubEnabled bool
	jwtEnabled    bool

	redisConnection
}

func (s *redisStorage) AsPubSubStorage() domain.PubSubStorage {
	if !s.pubsubEnabled {
		return nil
	}
	return s
}
func (s *redisStorage) AsJwtStorage() domain.JwtStorage {
	if !s.jwtEnabled {
		return nil
	}
	return s
}

func (s *redisStorage) String() string {
	if s.redisConnection.isSingleNode {
		return "redis-singlenode"
	}
	return "redis-cluster"
}

func (s *redisStorage) Shutdown(ctx context.Context) error {
	return s.redisConnection.close()
}

func (s *redisStorage) GetFileDescriptorPressure() int {
	return s.maxConnections
}
