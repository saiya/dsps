package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"

	"github.com/saiya/dsps/server/domain"
	. "github.com/saiya/dsps/server/storage/redis/internal"
	dspstesting "github.com/saiya/dsps/server/testing"
)

func TestPublishMessageScript(t *testing.T) {
	ctx := context.Background()

	WithRedisClient(t, func(redisCmd RedisCmd) {
		for _, testcase := range []struct {
			newChannel       bool
			clockBefore      channelClock
			duplicateMessage bool
		}{
			{newChannel: true, duplicateMessage: false},
			{newChannel: true, duplicateMessage: true},
			{newChannel: false, clockBefore: 123, duplicateMessage: false},
			{newChannel: false, clockBefore: 123, duplicateMessage: true},
			// Test clock overflow and also Lua number format issue of large float numbers:
			{newChannel: false, clockBefore: clockMax, duplicateMessage: false},
			{newChannel: false, clockBefore: clockMax, duplicateMessage: true},
		} {
			channelID := randomChannelID(t)
			keys := keyOfChannel(channelID)
			ttl := channelTTLSec(3)
			msgID := domain.MessageID("msg1")
			content := `{"hi":"hello"}`
			msg := domain.Message{
				MessageLocator: domain.MessageLocator{
					ChannelID: channelID,
					MessageID: msgID,
				},
				Content: json.RawMessage(content),
			}

			var clockAfter channelClock
			if testcase.newChannel {
				// Clock of first message must be 1 because newly created subscribers reads messages from 1 (not zero).
				clockAfter = 1
			} else {
				assert.NoError(t, redisCmd.Set(context.Background(), keys.Clock(), testcase.clockBefore))
				clockAfter = testcase.clockBefore + 1
				if clockAfter > clockMax {
					clockAfter = clockMin + (clockAfter - clockMax - 1)
				}
			}

			// 1st publish
			assert.NoError(t, runPublishMessageScript(ctx, redisCmd, ttl, msg))
			assertValueAndTTL(t, redisCmd, keys.Clock(), fmt.Sprintf("%d", clockAfter), time.Duration(ttl)*time.Second)

			// 2nd publish (duplicate)
			if testcase.duplicateMessage {
				assert.NoError(t, runPublishMessageScript(ctx, redisCmd, ttl, msg))
				// Should not advance clock
				assertValueAndTTL(t, redisCmd, keys.Clock(), fmt.Sprintf("%d", clockAfter), time.Duration(ttl)*time.Second)
			}

			// After publish
			assertValueAndTTL(t, redisCmd, keys.Clock(), fmt.Sprintf("%d", clockAfter), time.Duration(ttl)*time.Second)
			assertValueAndTTL(t, redisCmd, keys.MessageDedup(msgID), fmt.Sprintf("%d", clockAfter), time.Duration(ttl)*time.Second)
			assertValueAndTTL(t, redisCmd, keys.MessageBody(clockAfter), fmt.Sprintf(`{"id":"%s","content":%s}`, msgID, content), time.Duration(ttl)*time.Second)
		}
	})
}

func TestPublishMessageScriptAbormalResults(t *testing.T) {
	ctx := context.Background()
	channelID := randomChannelID(t)
	ttl := channelTTLSec(3)
	msgID := domain.MessageID("msg1")
	content := `{"hi":"hello"}`
	msg := domain.Message{
		MessageLocator: domain.MessageLocator{
			ChannelID: channelID,
			MessageID: msgID,
		},
		Content: json.RawMessage(content),
	}

	// Test error
	WithRedisClient(t, func(redisCmd RedisCmd) {
		originalScript := publishMessageScript
		defer func() { publishMessageScript = originalScript }()

		publishMessageScript = redis.NewScript(`syn tax error`)
		assert.Equal(
			t,
			`Failed to execute publishMessageScript: ERR Error compiling script (new function): user_script:1: '=' expected near 'tax'`,
			runPublishMessageScript(ctx, redisCmd, ttl, msg).Error(),
		)

		publishMessageScript = redis.NewScript(`return "What??"`)
		assert.Equal(
			t,
			`Unexpected result from publishMessageScript: string(What??)`,
			runPublishMessageScript(ctx, redisCmd, ttl, msg).Error(),
		)
	})
}

func TestAckScript(t *testing.T) {
	ctx := context.Background()
	WithRedisClient(t, func(redisCmd RedisCmd) {
		for _, testcase := range []struct {
			result       string
			channelClock channelClock
			ackClock     channelClock // Advance subscriber's clock toward this
			sbscClock    channelClock // Subscriber's clock before ack operation
		}{
			// stale/ok boundary check
			{result: "stale", channelClock: 4, ackClock: 0, sbscClock: 0},
			{result: "OK", channelClock: 4, ackClock: 1, sbscClock: 0},
			{result: "OK", channelClock: 4, ackClock: 2, sbscClock: 0},
			{result: "OK", channelClock: 4, ackClock: 4, sbscClock: 0},
			{result: "stale", channelClock: 4, ackClock: 5, sbscClock: 0},
			// negative clock
			{result: "stale", channelClock: -1000, ackClock: -1024, sbscClock: -1024},
			{result: "OK", channelClock: -1000, ackClock: -1012, sbscClock: -1024},
			{result: "stale", channelClock: -1000, ackClock: -999, sbscClock: -1024},
			// overflow (clockMin/Max ± 0), also tests Lua number formatting issue of large numbers
			{result: "stale", channelClock: clockMin, ackClock: clockMax - 1, sbscClock: clockMax},
			{result: "stale", channelClock: clockMin, ackClock: clockMax, sbscClock: clockMax},
			{result: "OK", channelClock: clockMin, ackClock: clockMin, sbscClock: clockMax},
			{result: "stale", channelClock: clockMin, ackClock: clockMin + 1, sbscClock: clockMax},
			// overflow (clockMin/Max ± 16)
			{result: "stale", channelClock: clockMin + 16, ackClock: clockMax - 16, sbscClock: clockMax - 16},
			{result: "OK", channelClock: clockMin + 16, ackClock: clockMax - 15, sbscClock: clockMax - 16},
			{result: "OK", channelClock: clockMin + 16, ackClock: clockMax, sbscClock: clockMax - 16},
			{result: "OK", channelClock: clockMin + 16, ackClock: clockMin, sbscClock: clockMax - 16},
			{result: "OK", channelClock: clockMin + 16, ackClock: clockMin + 16, sbscClock: clockMax - 16},
			{result: "stale", channelClock: clockMin + 16, ackClock: clockMin + 17, sbscClock: clockMax - 16},
		} {
			// Check test case validity
			if isClockWithin(testcase.ackClock, testcase.sbscClock, testcase.channelClock) {
				assert.Equal(t, "OK", testcase.result)
			} else {
				assert.Equal(t, "stale", testcase.result)
			}

			channelID := randomChannelID(t)
			keys := keyOfChannel(channelID)
			ttlBefore := 1000 * time.Second
			ttl := channelTTLSec(3)
			sbscID := domain.SubscriberID("sbsc-1")

			var sbscClockExpected channelClock
			var ttlExpected time.Duration
			if testcase.result == "stale" {
				sbscClockExpected = testcase.sbscClock
				ttlExpected = ttlBefore
			} else {
				sbscClockExpected = testcase.ackClock
				ttlExpected = time.Duration(ttl) * time.Second
			}

			assert.NoError(t, redisCmd.SetEX(ctx, keys.Clock(), testcase.channelClock, ttlBefore))
			assert.NoError(t, redisCmd.SetEX(ctx, keys.SubscriberCursor(sbscID), testcase.sbscClock, ttlBefore))

			result, err := runAckScript(ctx, redisCmd, channelID, ttl, sbscID, testcase.ackClock)

			assert.NoError(t, err)
			assert.Equal(t, testcase.result, result)
			assertValueAndTTL(t, redisCmd, keys.SubscriberCursor(sbscID), fmt.Sprintf("%d", sbscClockExpected), ttlExpected)
			assertValueAndTTL(t, redisCmd, keys.Clock(), fmt.Sprintf("%d", testcase.channelClock), ttlExpected)
		}
	})
}

func TestAckScriptNotFoundCase(t *testing.T) {
	ctx := context.Background()
	WithRedisClient(t, func(redisCmd RedisCmd) {
		channelID := randomChannelID(t)
		keys := keyOfChannel(channelID)
		ttl := channelTTLSec(3)
		sbscID := domain.SubscriberID("sbsc-1")

		// Channel not found
		result, err := runAckScript(ctx, redisCmd, channelID, ttl, sbscID, 123)
		assert.Equal(t, "channel-not-found", result)
		dspstesting.IsError(t, domain.ErrSubscriptionNotFound, err)

		// Create channel
		assert.NoError(t, redisCmd.Set(context.Background(), keys.Clock(), 123))

		// Subscription not found
		result, err = runAckScript(ctx, redisCmd, channelID, ttl, sbscID, 123)
		assert.Equal(t, "subscription-not-found", result)
		dspstesting.IsError(t, domain.ErrSubscriptionNotFound, err)
	})
}

func TestAckScriptAbormalResults(t *testing.T) {
	ctx := context.Background()
	channelID := randomChannelID(t)
	ttl := channelTTLSec(3)
	sbscID := domain.SubscriberID("sbsc-1")

	// Test error
	WithRedisClient(t, func(redisCmd RedisCmd) {
		originalScript := ackScript
		defer func() { ackScript = originalScript }()

		ackScript = redis.NewScript(`syn tax error`)
		_, err := runAckScript(ctx, redisCmd, channelID, ttl, sbscID, 123)
		assert.Equal(
			t,
			`Failed to execute ackScript: ERR Error compiling script (new function): user_script:1: '=' expected near 'tax'`,
			err.Error(),
		)

		ackScript = redis.NewScript(`return "What??"`)
		_, err = runAckScript(ctx, redisCmd, channelID, ttl, sbscID, 123)
		assert.Equal(
			t,
			`Unexpected result from ackScript: string(What??)`,
			err.Error(),
		)

		ackScript = redis.NewScript(`return true`)
		_, err = runAckScript(ctx, redisCmd, channelID, ttl, sbscID, 123)
		assert.Equal(
			t,
			`Unexpected result from ackScript: int64(1)`,
			err.Error(),
		)
	})
}
