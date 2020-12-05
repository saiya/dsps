package redis

import (
	"context"

	"golang.org/x/sync/errgroup"
)

// TODO: Periodically call this.
func (s *redisStorage) loadScripts(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return s.loadPubSubMessagingScripts(ctx) })
	g.Go(func() error { return s.loadPubSubSubscriberScripts(ctx) })
	return g.Wait()
}
