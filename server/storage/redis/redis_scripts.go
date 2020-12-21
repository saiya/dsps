package redis

import (
	"context"
	"sync"
	"time"

	"github.com/saiya/dsps/server/logger"
	"golang.org/x/sync/errgroup"
)

func (s *redisStorage) loadScripts(ctx context.Context) error {
	logger.Of(ctx).Debugf(logger.CatStorage, "loading Redis Lua scripts...")

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return s.loadPubSubMessagingScripts(ctx) })
	g.Go(func() error { return s.loadPubSubSubscriberScripts(ctx) })
	return g.Wait()
}

type scriptLoader struct {
	timerLock sync.Mutex
	timer     *time.Timer

	shutdownOnce sync.Once
	shutdown     chan interface{}
}

func (s *redisStorage) startScriptLoader(ctx context.Context, interval time.Duration) *scriptLoader {
	loader := &scriptLoader{shutdown: make(chan interface{})}
	var f func()
	f = func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		complete := make(chan interface{})
		go func() {
			defer close(complete)
			if err := s.loadScripts(ctx); err != nil {
				logger.Of(ctx).WarnError(logger.CatStorage, "Failed to pre-load Redis Lua scripts", err)
			}
		}()

		select {
		case <-complete:
			break
		case <-loader.shutdown:
			logger.Of(ctx).Debugf(logger.CatStorage, "Redis Lua script preloader stopped.")
			return
		}
		loader.scriptLoaderRunAfter(interval, f)
	}
	loader.scriptLoaderRunAfter(interval, f)
	return loader
}

func (loader *scriptLoader) scriptLoaderRunAfter(interval time.Duration, f func()) {
	loader.timerLock.Lock()
	defer loader.timerLock.Unlock()
	loader.timer = time.AfterFunc(interval, f)
}

func (loader *scriptLoader) stopScriptLoader(ctx context.Context) {
	loader.timerLock.Lock()
	defer loader.timerLock.Unlock()

	loader.shutdownOnce.Do(func() {
		close(loader.shutdown)
	})
	loader.timer.Stop()
}
