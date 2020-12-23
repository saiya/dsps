package redis

import (
	"context"

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
