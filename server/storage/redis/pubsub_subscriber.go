package redis

import (
	"context"

	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"

	"github.com/saiya/dsps/server/domain"
)

func (s *redisStorage) NewSubscriber(ctx context.Context, sl domain.SubscriberLocator) error {
	ttl, err := s.channelRedisTTLSec(sl.ChannelID)
	if err != nil {
		return xerrors.Errorf("Unable to calcurate TTL of channel: %w", err)
	}
	return runCreateSubscriberScript(ctx, s.RedisCmd, sl.ChannelID, ttl, sl.SubscriberID)
}

func (s *redisStorage) RemoveSubscriber(ctx context.Context, sl domain.SubscriberLocator) error {
	keys := keyOfChannel(sl.ChannelID)
	if err := s.RedisCmd.Del(ctx, keys.SubscriberCursor(sl.SubscriberID)); err != nil {
		return xerrors.Errorf("Failed to delete subscriber: %w", err)
	}
	return nil
}

// extendSubscriberTTL extends TTL of channel clock and subscriber.
// If no new messages comes in to the channel, fetchMessages operation should extend TTLs otherwise channel clock or subscriber could be vanished due to TTL outage.
func (s *redisStorage) extendSubscriberTTL(ctx context.Context, sl domain.SubscriberLocator) error {
	ttl, err := s.channelRedisTTLSec(sl.ChannelID)
	if err != nil {
		return xerrors.Errorf("Unable to calcurate TTL of channel: %w", err)
	}

	keys := keyOfChannel(sl.ChannelID)
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return s.RedisCmd.Expire(ctx, keys.Clock(), ttl.asDuration()) })
	g.Go(func() error { return s.RedisCmd.Expire(ctx, keys.SubscriberCursor(sl.SubscriberID), ttl.asDuration()) })
	return g.Wait()
}
