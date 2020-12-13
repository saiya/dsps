package redis

import (
	"context"
	"errors"

	"github.com/go-redis/redis/v8"
	"golang.org/x/xerrors"

	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/logger"
)

func (s *redisStorage) loadPubSubMessagingScripts(ctx context.Context) error {
	if err := s.redisCmd.LoadScript(ctx, publishMessageScript); err != nil {
		return xerrors.Errorf("Failed to load publishMessageScript: %w", err)
	}
	if err := s.redisCmd.LoadScript(ctx, ackScript); err != nil {
		return xerrors.Errorf("Failed to load ackScript: %w", err)
	}
	return nil
}

var publishMessageScript = redis.NewScript(`
	local clockKey = KEYS[1]	      -- Clock (c.{{channel}}.clock)
	local msgBodyKeyPrefix = KEYS[2]  -- MessageBodyPrefix (c.{{channel}}.m.)
	local msgDedupKey = KEYS[3]       -- MessageDedup (c.{{channel}}.mid.{messageID})
	local ttlSec = tonumber(ARGV[1])    -- (number) ttl [sec]
	local content = ARGV[2]             -- (string) content
	local clockMin = tonumber(ARGV[3])  -- (number) clockMin
	local clockMax = tonumber(ARGV[4])  -- (number) clockMax

	-- Increment chanel clock
	local nextClock = tonumber(redis.call("incr", clockKey))
	local oldClock = nextClock - 1
	if nextClock > clockMax then
		nextClock = clockMin
		redis.call("set", clockKey, nextClock)
	end
	redis.call("expire", clockKey, ttlSec)

	-- Publish message
	if redis.call("set", msgDedupKey, string.format("%d", nextClock), "EX", ttlSec, "NX") == false then
		-- Rollback
		redis.call("set", clockKey, string.format("%d", oldClock))
		redis.call("expire", clockKey, ttlSec)
		return false
	end
	redis.call("set", msgBodyKeyPrefix .. string.format("%d", nextClock), content, "EX", ttlSec)
	return redis.status_reply("OK")
`)

func runPublishMessageScript(ctx context.Context, redisCmd redisCmd, ttl channelTTLSec, msg domain.Message) error {
	wrapped, err := wrapMessage(msg)
	if err != nil {
		return xerrors.Errorf("Unable to encode message \"%s\": %w", msg.MessageID, err)
	}

	keys := keyOfChannel(msg.ChannelID)
	result, err := redisCmd.RunScript(
		ctx, publishMessageScript,
		[]string{keys.Clock(), keys.MessageBodyPrefix(), keys.MessageDedup(msg.MessageID)},
		ttl, wrapped, clockMin, clockMax,
	)
	logger.Of(ctx).Debugf(logger.CatStorage, "runPublishMessageScript(ttl = %d, msg = %v) resulted in %v (%v)", ttl, msg, result, err)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			logger.Of(ctx).Debugf(logger.CatStorage, "Duplicated message %s / %s", msg.ChannelID, msg.MessageID)
		} else {
			return xerrors.Errorf("Failed to execute publishMessageScript: %w", err)
		}
	} else {
		if result != "OK" {
			return xerrors.Errorf("Unexpected result from publishMessageScript: %T(%v)", result, result)
		}
	}
	return nil
}

var ackScript = redis.NewScript(`
	local channelClockKey = KEYS[1]  -- Clock of the channel (c.{channel}.clock)
	local sbscClockKey = KEYS[2]     -- Clock of the subscriber (c.{{channel}}.r.{subscriber})
	local ttlSec = tonumber(ARGV[1])            -- (number) ttl [sec]
	local acknowledgedClock = tonumber(ARGV[2]) -- (number) Clock of the latest acknowledged message

	local channelClock = redis.call("get", channelClockKey)
	local sbscClock = redis.call("get", sbscClockKey)
	if channelClock == false then return "channel-not-found" end
	if sbscClock == false then return "subscription-not-found" end
	channelClock = tonumber(channelClock)
	sbscClock = tonumber(sbscClock)

	if channelClock < sbscClock then
		-- Valid range is (sbscClock, clockMax] and [clockMin, channelClock]
		if not ((sbscClock < acknowledgedClock) or (acknowledgedClock <= channelClock)) then
			return "stale"
		end
	else
		-- Valid range is (sbscClock, channelClock]
		if not ((sbscClock < acknowledgedClock) and (acknowledgedClock <= channelClock)) then
			return "stale"
		end
	end
	redis.call("set", sbscClockKey, string.format("%d", acknowledgedClock), "EX", ttlSec)
	redis.call("expire", channelClockKey, ttlSec)  -- Also extend channel expiry
	return redis.status_reply("OK")
`)

func runAckScript(ctx context.Context, redisCmd redisCmd, channelID domain.ChannelID, ttl channelTTLSec, sbscID domain.SubscriberID, acknowledgedClock channelClock) (string, error) {
	keys := keyOfChannel(channelID)
	result, err := redisCmd.RunScript(
		ctx, ackScript,
		[]string{
			keys.Clock(),
			keys.SubscriberCursor(sbscID),
		},
		ttl, int64(acknowledgedClock),
	)
	logger.Of(ctx).Debugf(logger.CatStorage, `runAckScript(channelID = %s, ttl = %d, sbscID = %s, acknowledgedClock = %s) resulted in %v (%v)`, channelID, ttl, sbscID, acknowledgedClock, result, err)
	if err != nil {
		return "", xerrors.Errorf("Failed to execute ackScript: %w", err)
	}
	if strResult, ok := result.(string); ok {
		switch strResult {
		case "OK":
			return strResult, nil
		case "channel-not-found", "subscription-not-found":
			return strResult, xerrors.Errorf("%s (%w)", strResult, domain.ErrSubscriptionNotFound)
		case "stale":
			return strResult, nil // Could occur due to client retry
		default:
			return strResult, xerrors.Errorf("Unexpected result from ackScript: string(%s)", strResult)
		}
	}
	return "", xerrors.Errorf("Unexpected result from ackScript: %T(%v)", result, result)
}
