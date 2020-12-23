package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/logger"
	"github.com/saiya/dsps/server/sync"
	"github.com/saiya/dsps/server/telemetry"
)

// In case of clock drift
const ttlMargin = 15 * time.Second

// NewRedisStorage creates Storage instance
func NewRedisStorage(ctx context.Context, config *config.RedisStorageConfig, systemClock domain.SystemClock, channelProvider domain.ChannelProvider, telemetry *telemetry.Telemetry) (domain.Storage, error) {
	conn, err := connect(ctx, config)
	if err != nil {
		return nil, err
	}
	s := &redisStorage{
		clock:           systemClock,
		channelProvider: channelProvider,

		pubsubEnabled: !config.DisablePubSub,
		jwtEnabled:    !config.DisableJwt,

		redisConnection: conn,
		daemonSystem: sync.NewDaemonSystem("dsps.storage.redis", telemetry, func(ctx context.Context, name string, err error) {
			logger.Of(ctx).Error(fmt.Sprintf(`error in background routine "%s"`, name), err)
		}),
	}
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

	redisConnection
	daemonSystem *sync.DaemonSystem
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
	if err := s.daemonSystem.Shutdown(ctx); err != nil {
		logger.Of(ctx).WarnError(logger.CatStorage, `Failed to stop background routines`, err)
	}

	logger.Of(ctx).Debugf(logger.CatStorage, "Closing Redis storage connections...")
	return s.redisConnection.close()
}

func (s *redisStorage) GetFileDescriptorPressure() int {
	return s.maxConnections
}
