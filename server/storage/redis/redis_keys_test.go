package redis

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChannelKeys(t *testing.T) {
	keys := keyOfChannel("my-channel")

	// All redis keys must contain {channel-id} string to control partitioning, otherwise Lua script / transaction fails due to cross partition operation.
	assert.Contains(t, keys.Clock(), "{my-channel}")
	assert.Contains(t, keys.SubscriberCursor("sbsc-1"), "{my-channel}")
	assert.Contains(t, keys.MessageBodyPrefix(), "{my-channel}")
	assert.Contains(t, keys.MessageBody(1234), "{my-channel}")
	assert.Contains(t, keys.MessageDedup("msg-1"), "{my-channel}")

	// MessageBody must start with MessageBodyPrefix
	assert.True(t, strings.HasPrefix(keys.MessageBody(1234), keys.MessageBodyPrefix()))

	// Check uniqueness
	keys2 := keyOfChannel("my-channel-X")
	assert.NotEqual(t, keys.Clock(), keys2.Clock())
	assert.NotEqual(t, keys.SubscriberCursor("sbsc-1"), keys.SubscriberCursor("sbsc-X"))
	assert.NotEqual(t, keys.SubscriberCursor("sbsc-1"), keys2.SubscriberCursor("sbsc-1"))
	assert.NotEqual(t, keys.MessageBodyPrefix(), keys2.MessageBodyPrefix())
	assert.NotEqual(t, keys.MessageBody(1234), keys.MessageBody(1234+1))
	assert.NotEqual(t, keys.MessageBody(1234), keys2.MessageBody(1234))
	assert.NotEqual(t, keys.MessageDedup("msg-1"), keys.MessageDedup("msg-X"))
	assert.NotEqual(t, keys.MessageDedup("msg-1"), keys2.MessageDedup("msg-1"))
}

func TestJtiKeys(t *testing.T) {
	keys := keyOfJti("my-jwt")

	// Check uniqueness
	keys2 := keyOfJti("my-jwt-X")
	assert.NotEqual(t, keys.Revocation(), keys2.Revocation())
}
