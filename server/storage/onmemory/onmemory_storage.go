package onmemory

import (
	"context"
	"fmt"

	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/logger"
	"github.com/saiya/dsps/server/sync"
)

// NewOnmemoryStorage creates Storage instance
func NewOnmemoryStorage(ctx context.Context, config *config.OnmemoryStorageConfig, systemClock domain.SystemClock, channelProvider domain.ChannelProvider) (domain.Storage, error) {
	s := &onmemoryStorage{
		lock: sync.NewLock(),

		systemClock:     systemClock,
		channelProvider: channelProvider,

		pubsubEnabled: !config.DisablePubSub,
		jwtEnabled:    !config.DisableJwt,

		runGcOnShutdown: config.RunGCOnShutdown,
		daemonSystem: sync.NewDaemonSystem("dsps.storage.onmemory", func(ctx context.Context, name string, err error) {
			logger.Of(ctx).Error(fmt.Sprintf(`error in background routine "%s"`, name), err)
		}),

		channels: map[domain.ChannelID]*onmemoryChannel{},

		revokedJwts: map[domain.JwtJti]domain.JwtExp{},
	}

	s.startGC()

	return s, nil
}

type onmemoryStorage struct {
	lock sync.Lock

	pubsubEnabled bool
	jwtEnabled    bool

	systemClock     domain.SystemClock
	channelProvider domain.ChannelProvider

	daemonSystem    *sync.DaemonSystem
	runGcOnShutdown bool

	channels map[domain.ChannelID]*onmemoryChannel

	revokedJwts map[domain.JwtJti]domain.JwtExp
}

func (s *onmemoryStorage) String() string {
	return "onmemory"
}

func (s *onmemoryStorage) Shutdown(ctx context.Context) error {
	logger.Of(ctx).Debugf(logger.CatStorage, "Closing on-memory storage...")

	if err := s.daemonSystem.Shutdown(ctx); err != nil {
		logger.Of(ctx).WarnError(logger.CatStorage, "Failed to stop background routines", err)
	}
	if s.runGcOnShutdown {
		// Note: GC locks s.lock, so that do not call this after s.lock
		if err := s.GC(ctx); err != nil {
			return err
		}
	}

	unlock, err := s.lock.Lock(ctx)
	if err != nil {
		return err
	}
	defer unlock()

	s.channels = map[domain.ChannelID]*onmemoryChannel{} // Drop all data
	return nil
}

func (s *onmemoryStorage) AsPubSubStorage() domain.PubSubStorage {
	if !s.pubsubEnabled {
		return nil
	}
	return s
}
func (s *onmemoryStorage) AsJwtStorage() domain.JwtStorage {
	if !s.jwtEnabled {
		return nil
	}
	return s
}

func (s *onmemoryStorage) GetFileDescriptorPressure() int {
	return 0
}
