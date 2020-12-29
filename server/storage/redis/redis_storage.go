package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/logger"
	"github.com/saiya/dsps/server/storage/deps"
	"github.com/saiya/dsps/server/storage/redis/internal"
	"github.com/saiya/dsps/server/storage/redis/internal/pubsub"
	"github.com/saiya/dsps/server/sync"
)

// In case of clock drift
const ttlMargin = 15 * time.Second

// NewRedisStorage creates Storage instance
func NewRedisStorage(ctx context.Context, config *config.RedisStorageConfig, systemClock domain.SystemClock, channelProvider domain.ChannelProvider, deps deps.StorageDeps) (domain.Storage, error) {
	conn, err := internal.NewRedisConnection(ctx, config)
	if err != nil {
		return nil, err
	}
	s := &redisStorage{
		clock:           systemClock,
		channelProvider: channelProvider,

		pubsubEnabled: !config.DisablePubSub,
		jwtEnabled:    !config.DisableJwt,

		RedisConnection: conn,
		daemonSystem: sync.NewDaemonSystem("dsps.storage.redis", sync.DaemonSystemDeps{
			Telemetry: deps.Telemetry,
			Sentry:    deps.Sentry,
		}, func(ctx context.Context, name string, err error) {
			logger.Of(ctx).Error(fmt.Sprintf(`error in background routine "%s"`, name), err)
		}),
	}
	s.pubsubDispatcher = pubsub.NewDispatcher(ctx, deps, pubsub.DispatcherParams{}, conn.RedisCmd.PSubscribeFunc(), s.redisPubSubKeyPattern())
	if err := s.loadScripts(ctx); err != nil {
		return nil, err
	}
	s.daemonSystem.Start("scriptLoader", func(ctx context.Context) (sync.DaemonNextRun, error) {
		err := s.loadScripts(ctx)
		return sync.DaemonNextRun{
			Interval: config.ScriptReloadInterval.Duration,
		}, err
	})
	return s, nil
}

type redisStorage struct {
	clock           domain.SystemClock
	channelProvider domain.ChannelProvider

	pubsubEnabled bool
	jwtEnabled    bool

	internal.RedisConnection
	daemonSystem     *sync.DaemonSystem
	pubsubDispatcher pubsub.RedisPubSubDispatcher
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
	if s.RedisConnection.IsSingleNode {
		return "redis-singlenode"
	}
	return "redis-cluster"
}

func (s *redisStorage) Shutdown(ctx context.Context) error {
	if err := s.daemonSystem.Shutdown(ctx); err != nil {
		logger.Of(ctx).WarnError(logger.CatStorage, `Failed to stop background routines`, err)
	}

	s.pubsubDispatcher.Shutdown(ctx)

	logger.Of(ctx).Debugf(logger.CatStorage, "Closing Redis storage connections...")
	return s.RedisConnection.Close()
}

func (s *redisStorage) GetFileDescriptorPressure() int {
	return s.MaxConnections
}
