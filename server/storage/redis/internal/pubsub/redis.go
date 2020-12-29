package pubsub

import (
	"context"

	"github.com/go-redis/redis/v8"
)

// RedisSubscribeRawFunc represents (P)SUBSCRIBE command implementation.
type RedisSubscribeRawFunc func(ctx context.Context, channel RedisChannelID) RedisRawPubSub

// RedisChannelID represents channel ID of Redis Pub/Sub
type RedisChannelID string

// RedisRawPubSub is subset of *redis.PubSub
type RedisRawPubSub interface {
	Receive(context.Context) (interface{}, error)
	Ping(context.Context, ...string) error
	ChannelSize(int) <-chan *redis.Message
	Close() error
}
