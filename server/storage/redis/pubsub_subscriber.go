package redis

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/saiya/dsps/server/domain"
)

func (s *redisStorage) NewSubscriber(ctx context.Context, sl domain.SubscriberLocator) error {
	ttl, err := s.channelRedisTTLSec(sl.ChannelID)
	if err != nil {
		return xerrors.Errorf("Unable to calcurate TTL of channel: %w", err)
	}
	return runCreateSubscriberScript(ctx, s.redisCmd, sl.ChannelID, ttl, sl.SubscriberID)
}

func (s *redisStorage) RemoveSubscriber(ctx context.Context, sl domain.SubscriberLocator) error {
	keys := keyOfChannel(sl.ChannelID)
	if err := s.redisCmd.Del(ctx, keys.SubscriberCursor(sl.SubscriberID)); err != nil {
		return xerrors.Errorf("Failed to delete subscriber: %w", err)
	}
	return nil
}
