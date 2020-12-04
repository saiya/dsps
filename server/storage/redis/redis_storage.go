package redis

import (
	"context"
	"errors"

	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/domain"
)

// NewRedisStorage creates Storage instance
func NewRedisStorage(ctx context.Context, config *config.RedisStorageConfig, systemClock domain.SystemClock, channelProvider domain.ChannelProvider) (domain.Storage, error) {
	return &redisStorage{
		stat: &redisStorageStat{},

		pubsubEnabled: !config.DisablePubSub,
		jwtEnabled:    !config.DisableJwt,
	}, nil
}

type redisStorage struct {
	stat *redisStorageStat

	pubsubEnabled bool
	jwtEnabled    bool
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
	return "redis" // TODO: Add "-cluster" / "-singlenode" suffix
}

func (s *redisStorage) Shutdown(ctx context.Context) error {
	return errors.New("Not Implemented yet")
}
