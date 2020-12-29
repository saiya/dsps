package redis

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"

	"github.com/saiya/dsps/server/domain"
	. "github.com/saiya/dsps/server/storage/redis/internal"
)

func TestSubscriberScript(t *testing.T) {
	ctx := context.Background()

	// Test newly created channel
	WithRedisClient(t, func(redisCmd RedisCmd) {
		channelID := randomChannelID(t)
		keys := keyOfChannel(channelID)
		ttl := channelTTLSec(3)
		sbscID := domain.SubscriberID("sbsc1")

		assert.NoError(t, runCreateSubscriberScript(ctx, redisCmd, channelID, ttl, sbscID))

		assertValueAndTTL(t, redisCmd, keys.Clock(), "0", time.Duration(ttl)*time.Second)
		assertValueAndTTL(t, redisCmd, keys.SubscriberCursor(sbscID), "0", time.Duration(ttl)*time.Second)
	})

	// Test existing channel
	WithRedisClient(t, func(redisCmd RedisCmd) {
		channelID := randomChannelID(t)
		keys := keyOfChannel(channelID)
		ttl := channelTTLSec(3)
		sbscID := domain.SubscriberID("sbsc1")

		clock := channelClock(-1024)
		assert.NoError(t, redisCmd.Set(ctx, keys.Clock(), clock))

		assert.NoError(t, runCreateSubscriberScript(ctx, redisCmd, channelID, ttl, sbscID))

		assertValueAndTTL(t, redisCmd, keys.Clock(), "-1024", time.Duration(ttl)*time.Second)
		assertValueAndTTL(t, redisCmd, keys.SubscriberCursor(sbscID), "-1024", time.Duration(ttl)*time.Second)
	})

	// Test Lua number format issue (large float format)
	WithRedisClient(t, func(redisCmd RedisCmd) {
		channelID := randomChannelID(t)
		keys := keyOfChannel(channelID)
		ttl := channelTTLSec(3)
		sbscID := domain.SubscriberID("sbsc1")

		clock := channelClock(clockMin)
		assert.NoError(t, redisCmd.Set(ctx, keys.Clock(), clock))

		assert.NoError(t, runCreateSubscriberScript(ctx, redisCmd, channelID, ttl, sbscID))

		assertValueAndTTL(t, redisCmd, keys.Clock(), fmt.Sprintf("%d", clockMin), time.Duration(ttl)*time.Second)
		assertValueAndTTL(t, redisCmd, keys.SubscriberCursor(sbscID), fmt.Sprintf("%d", clockMin), time.Duration(ttl)*time.Second)
	})
}

func TestSubscriberScriptAbormalResults(t *testing.T) {
	ctx := context.Background()

	channelID := randomChannelID(t)
	ttl := channelTTLSec(3)
	sbscID := domain.SubscriberID("sbsc1")

	// Test error
	WithRedisClient(t, func(redisCmd RedisCmd) {
		originalScript := createSubscriberScript
		defer func() { createSubscriberScript = originalScript }()

		createSubscriberScript = redis.NewScript(`syn tax error`)
		assert.Equal(
			t,
			`Failed to execute createSubscriberScript: ERR Error compiling script (new function): user_script:1: '=' expected near 'tax'`,
			runCreateSubscriberScript(ctx, redisCmd, channelID, ttl, sbscID).Error(),
		)
	})

	// Test invalid result
	WithRedisClient(t, func(redisCmd RedisCmd) {
		originalScript := createSubscriberScript
		defer func() { createSubscriberScript = originalScript }()

		createSubscriberScript = redis.NewScript(`return "What??"`)
		assert.Equal(
			t,
			`Unexpected result from createSubscriberScript: string(What??)`,
			runCreateSubscriberScript(ctx, redisCmd, channelID, ttl, sbscID).Error(),
		)
	})
}
