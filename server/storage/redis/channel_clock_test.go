package redis

import (
	"context"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

func TestClockOverflow(t *testing.T) {
	// Lua (Redis script) number is always float64
	// https://stackoverflow.com/a/6060451
	assert.Less(t, float64(clockMin-1), float64(clockMin))
	assert.Less(t, float64(clockMax), float64(clockMax+1))
}

func TestClockOverflowWithRedisLua(t *testing.T) {
	r := redis.NewClient(&redis.Options{Addr: GetRedisAddr(t)})
	defer func() { assert.NoError(t, r.Close()) }()

	for _, num := range []channelClock{clockMin, clockMax} {
		result, err := redis.NewScript(`return ARGV[1] + 1`).Run(context.Background(), r, []string{}, int64(num)).Result()
		assert.NoError(t, err)
		assert.Equal(t, int64(num+1), result)

		result, err = redis.NewScript(`return ARGV[1] - 1`).Run(context.Background(), r, []string{}, int64(num)).Result()
		assert.NoError(t, err)
		assert.Equal(t, int64(num-1), result)
	}
}

func TestIterateClocks(t *testing.T) {
	assert.Equal(
		t,
		[]channelClock{},
		iterateClocks(5, 100, 100),
	)
	assert.Equal(
		t,
		[]channelClock{-1, 0, 1, 2},
		iterateClocks(10, -2, 2),
	)
	assert.Equal(
		t,
		[]channelClock{-101, -100, -99},
		iterateClocks(5, -102, -99),
	)

	// Overflow
	assert.Equal(
		t,
		[]channelClock{clockMax - 1, clockMax, clockMin, clockMin + 1, clockMin + 2},
		iterateClocks(5, clockMax-2, clockMin+100),
	)
	assert.Equal(
		t,
		[]channelClock{clockMax - 1, clockMax, clockMin, clockMin + 1, clockMin + 2},
		iterateClocks(10, clockMax-2, clockMin+2),
	)
}
