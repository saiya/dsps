package redis

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/xerrors"

	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/logger"
	"github.com/saiya/dsps/server/storage/redis/internal/pubsub"
)

func (s *redisStorage) PublishMessages(ctx context.Context, msgs []domain.Message) error {
	if !domain.BelongsToSameChannel(msgs) {
		return xerrors.New("Messages belongs to various channels")
	}
	if len(msgs) == 0 {
		return nil
	}

	sentMsgs := 0
	defer func() {
		if sentMsgs > 0 {
			if err := s.RedisCmd.Publish(ctx, s.redisPubSubKeyOf(msgs[0].ChannelID), "new message"); err != nil {
				logger.Of(ctx).WarnError(logger.CatStorage, "Redis Pub/Sub publish failed. Subscribers could not receive messages immediately.", err)
			}
		}
	}()
	for _, msg := range msgs {
		ttl, err := s.channelRedisTTLSec(msg.ChannelID)
		if err != nil {
			return xerrors.Errorf("Unable to calcurate TTL of channel: %w", err)
		}
		if err := runPublishMessageScript(ctx, s.RedisCmd, ttl, msg); err != nil {
			return err
		}
		sentMsgs++
	}
	return nil
}

func (s *redisStorage) FetchMessages(ctx context.Context, sl domain.SubscriberLocator, max int, waituntil domain.Duration) (messages []domain.Message, moreMessages bool, ackHandle domain.AckHandle, err error) {
	var await pubsub.RedisPubSubAwaiter
	var awaitCancel func(error)
	defer func() {
		if awaitCancel != nil {
			awaitCancel(context.Canceled)
		}
	}()
	if waituntil.Duration > 0 {
		await, awaitCancel = s.pubsubDispatcher.Await(ctx, s.redisPubSubKeyOf(sl.ChannelID))
	}

	if messages, moreMessages, ackHandle, err = s.fetchMessagesNow(ctx, sl, max); err != nil || len(messages) > 0 {
		return
	}

	timeout := time.NewTimer(waituntil.Duration)
	defer timeout.Stop()
	for {
		var c chan interface{}
		if await != nil {
			c = await.Chan()
		}
		select {
		case <-timeout.C:
			return
		case <-ctx.Done():
			err = ctx.Err()
			return
		case <-c:
			if await != nil && await.Err() != nil {
				err = await.Err()
				return
			}
			if messages, moreMessages, ackHandle, err = s.fetchMessagesNow(ctx, sl, max); err != nil || len(messages) > 0 {
				return
			}
			// Await again because no messages found (spurious wakeup)
			await, awaitCancel = s.pubsubDispatcher.Await(ctx, s.redisPubSubKeyOf(sl.ChannelID))
		}
	}
}

func (s *redisStorage) fetchMessagesNow(ctx context.Context, sl domain.SubscriberLocator, max int) (messages []domain.Message, moreMessages bool, ackHandle domain.AckHandle, err error) {
	keys := keyOfChannel(sl.ChannelID)
	clocks, err := s.RedisCmd.MGet(ctx, keys.Clock(), keys.SubscriberCursor(sl.SubscriberID))
	if err != nil {
		err = xerrors.Errorf("FetchMessages failed due to Redis error (cursor MGET error): %w", err)
		return
	}

	if clocks[0] == nil || clocks[1] == nil {
		err = domain.ErrSubscriptionNotFound
		return
	}
	chClock := parseChannelClock(*clocks[0])
	sbscClock := parseChannelClock(*clocks[1])
	if chClock == nil || sbscClock == nil {
		err = domain.ErrSubscriptionNotFound
		return
	}
	if err := s.extendSubscriberTTL(ctx, sl); err != nil { // We could use GETEX (>= Redis 6.2.0) rather than issue MGET + EXPIRE in the future.
		logger.Of(ctx).WarnError(logger.CatStorage, `Failed to extend TTL of channel clock entry and/or subscription clock entry of Redis`, err)
	}

	msgClocks := iterateClocks(max, *sbscClock, *chClock)
	msgKeys := make([]string, len(msgClocks)) // Must same length with msgClocks
	moreMessages = (len(msgClocks) > 0)
	for i, clock := range msgClocks {
		msgKeys[i] = keys.MessageBody(clock)
		if clock == *chClock {
			moreMessages = false
		}
	}
	rawMsgs, err := s.RedisCmd.MGet(ctx, msgKeys...)
	if err != nil {
		err = xerrors.Errorf("FetchMessages failed due to Redis error (msg MGET error): %w", err)
		return
	}

	var lastMessageClock *channelClock = nil
	ackHandle = domain.AckHandle{}
	messages = make([]domain.Message, 0, max)
	for i, rawPtr := range rawMsgs {
		var raw string = ""
		if rawPtr != nil {
			raw = *rawPtr
		}
		msg, err := unwrapMessage(sl.ChannelID, raw)
		if err != nil || msg == nil {
			if err != nil {
				logger.Of(ctx).Error(fmt.Sprintf("Skipped corrupted message (chID: %s, clock: %d) fetched from Redis", sl.ChannelID, msgClocks[i]), err)
			}
			continue // may caused by message TTL expiration
		}
		messages = append(messages, *msg)
		if lastMessageClock == nil || *lastMessageClock < msgClocks[i] {
			lastMessageClock = &msgClocks[i]
		}
	}
	if lastMessageClock != nil {
		ackHandle = encodeAckHandle(sl, ackHandleData{
			LastMessageClock: *lastMessageClock,
		})
	}
	return
}

func (s *redisStorage) AcknowledgeMessages(ctx context.Context, handle domain.AckHandle) error {
	ttl, err := s.channelRedisTTLSec(handle.ChannelID)
	if err != nil {
		return xerrors.Errorf("Unable to calcurate TTL of channel: %w", err)
	}
	h, err := decodeAckHandle(handle)
	if err != nil {
		return err
	}
	_, err = runAckScript(ctx, s.RedisCmd, handle.ChannelID, ttl, handle.SubscriberID, h.LastMessageClock)
	return err
}

func (s *redisStorage) IsOldMessages(ctx context.Context, sl domain.SubscriberLocator, msgs []domain.MessageLocator) (map[domain.MessageLocator]bool, error) {
	keys := keyOfChannel(sl.ChannelID)

	const mGetOffset = 2 // Clock and SubscriberCursor
	mGetKeys := make([]string, mGetOffset+len(msgs))
	mGetKeys[0] = keys.Clock()
	mGetKeys[1] = keys.SubscriberCursor(sl.SubscriberID)
	for i, msg := range msgs {
		mGetKeys[i+mGetOffset] = keys.MessageDedup(msg.MessageID)
	}

	clocks, err := s.RedisCmd.MGet(ctx, mGetKeys...)
	if err != nil {
		return nil, xerrors.Errorf("IsOldMessages failed due to Redis error (MGET error): %w", err)
	}

	if clocks[0] == nil || clocks[1] == nil {
		return nil, xerrors.Errorf("%w (%v)", domain.ErrSubscriptionNotFound, err)
	}
	chCursor := parseChannelClock(*clocks[0])
	sbscCursor := parseChannelClock(*clocks[1])
	if chCursor == nil || sbscCursor == nil {
		return nil, xerrors.Errorf("%w (%v)", domain.ErrSubscriptionNotFound, err)
	}
	result := make(map[domain.MessageLocator]bool, len(msgs))
	for i, clock := range clocks[mGetOffset:] {
		var msgClock *channelClock = nil
		if clock != nil {
			msgClock = parseChannelClock(*clock)
		}
		if msgClock == nil {
			result[msgs[i]] = false // Message not found (unsent or expired), return false because unsure.
			continue
		}
		result[msgs[i]] = !isClockWithin(*msgClock, *sbscCursor, *chCursor)
	}
	return result, nil
}

func (s *redisStorage) redisPubSubKeyOf(channelID domain.ChannelID) pubsub.RedisChannelID {
	return pubsub.RedisChannelID(fmt.Sprintf("dsps.c.{%s}", channelID))
}

func (s *redisStorage) redisPubSubKeyPattern() pubsub.RedisChannelID {
	return pubsub.RedisChannelID("dsps.c.*")
}
