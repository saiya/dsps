package redis

import (
	"context"
	"errors"

	"github.com/go-redis/redis/v8"
	"golang.org/x/xerrors"

	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/logger"
)

func (s *redisStorage) loadPubSubSubscriberScripts(ctx context.Context) error {
	if err := s.redisCmd.LoadScript(ctx, createSubscriberScript); err != nil {
		return xerrors.Errorf("Failed to load createSubscriberScript: %w", err)
	}
	return nil
}

// @returns "OK" (Redis status reply) if succeeded
// @returns false (Nil bulk reply) if already exists
var createSubscriberScript = redis.NewScript(`
	local clockKey = KEYS[1]	      -- Clock (c.{{channel}}.clock)
	local subscriberKey = KEYS[2]     -- XXXX (c.{{channel}}.r.{subscriber})
	local ttlSec = tonumber(ARGV[1])  -- (number) ttl [sec]

	local chClock = tonumber(redis.call("get", clockKey))
	if chClock == nil then
		chClock = 0
		redis.call("set", clockKey, 0, "EX", ttlSec)  -- Create channel
	else
		redis.call("expire", clockKey, ttlSec)  -- Extend channel life
	end

	-- Create subscriber if not exists
	return redis.call("set", subscriberKey, string.format("%d", chClock), "EX", ttlSec, "NX")
`)

func runCreateSubscriberScript(ctx context.Context, redisCmd redisCmd, channelID domain.ChannelID, ttl channelTTLSec, sbscID domain.SubscriberID) error {
	keys := keyOfChannel(channelID)
	result, err := redisCmd.RunScript(
		ctx, createSubscriberScript,
		[]string{keys.Clock(), keys.SubscriberCursor(sbscID)},
		ttl,
	)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			logger.Of(ctx).Debugf(logger.CatStorage, "Subscriber already exists: %s on %s", sbscID, keys.channelID)
		} else {
			return xerrors.Errorf("Failed to execute createSubscriberScript: %w", err)
		}
	} else if result != "OK" {
		return xerrors.Errorf("Unexpected result from createSubscriberScript: %T(%v)", result, result)
	}
	return nil
}
