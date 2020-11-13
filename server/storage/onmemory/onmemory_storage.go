package onmemory

import (
	"context"
	"time"

	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/sync"
)

// NewOnmemoryStorage creates Storage instance
func NewOnmemoryStorage(config *config.OnmemoryStorageConfig, systemClock domain.SystemClock, channelProvider domain.ChannelProvider) (domain.Storage, error) {
	s := &onmemoryStorage{
		lock: sync.NewLock(),
		stat: &onmemoryStorageStat{},

		systemClock:     systemClock,
		channelProvider: channelProvider,

		pubsubEnabled: !config.DisablePubSub,
		jwtEnabled:    !config.DisableJwt,

		runGcOnShutdown:         config.RunGCOnShutdown,
		gcTicker:                time.NewTicker(5 * time.Minute),
		gcTickerShutdownRequest: make(chan bool, 1),

		channels: map[domain.ChannelID]*onmemoryChannel{},

		revokedJwts: map[domain.JwtJti]domain.JwtExp{},
	}

	s.startGC()

	return s, nil
}

type onmemoryStorage struct {
	lock sync.Lock
	stat *onmemoryStorageStat

	pubsubEnabled bool
	jwtEnabled    bool

	systemClock     domain.SystemClock
	channelProvider domain.ChannelProvider

	runGcOnShutdown         bool
	gcTicker                *time.Ticker
	gcTickerShutdownRequest chan bool

	channels map[domain.ChannelID]*onmemoryChannel

	revokedJwts map[domain.JwtJti]domain.JwtExp
}

func (s *onmemoryStorage) String() string {
	return "onmemory"
}

func (s *onmemoryStorage) Shutdown(ctx context.Context) error {
	select {
	case s.gcTickerShutdownRequest <- true:
	default:
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
