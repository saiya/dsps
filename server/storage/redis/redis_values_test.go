package redis

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseRedisInt64(t *testing.T) {
	assert.Nil(t, parseRedisInt64(""))
	assert.Nil(t, parseRedisInt64("<< invalid >>"))

	assert.Equal(t, int64(0), *parseRedisInt64("0"))
	assert.Equal(t, int64(-9223372036854775808), *parseRedisInt64("-9223372036854775808"))
	assert.Equal(t, int64(9223372036854775807), *parseRedisInt64("9223372036854775807"))
}
