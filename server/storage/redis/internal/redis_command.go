package internal

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/saiya/dsps/server/storage/redis/internal/pubsub"
)

//go:generate mockgen -source=${GOFILE} -package=mock -destination=./mock/${GOFILE}

// RedisCmd wraps Redis command system
type RedisCmd interface {
	Ping(ctx context.Context) error

	Publish(ctx context.Context, channel pubsub.RedisChannelID, message interface{}) error
	PSubscribeFunc() pubsub.RedisSubscribeRawFunc

	Get(ctx context.Context, key string) (*string, error)
	MGet(ctx context.Context, keys ...string) ([]*string, error)
	TTL(ctx context.Context, key string) (*time.Duration, error)
	// EXPIRE command set TTL of the entry, not discarding the entry (name came from https://redis.io/commands/expire)
	Expire(ctx context.Context, key string, ttl time.Duration) error
	Set(ctx context.Context, key string, value interface{}) error
	SetEX(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Del(ctx context.Context, key string) error

	LoadScript(ctx context.Context, script *redis.Script) error
	RunScript(ctx context.Context, script *redis.Script, keys []string, args ...interface{}) (interface{}, error)
}

// NewRedisCmd creates new RedisCmd instance.
func NewRedisCmd(raw redis.Cmdable, psubscribeFunc pubsub.RedisSubscribeRawFunc) RedisCmd {
	return &redisCmdImpl{raw: raw, psubscribeFunc: psubscribeFunc}
}

type redisCmdImpl struct {
	raw            redis.Cmdable
	psubscribeFunc pubsub.RedisSubscribeRawFunc
}

func (impl *redisCmdImpl) Ping(ctx context.Context) error {
	_, err := impl.raw.Ping(ctx).Result()
	return err
}

func (impl *redisCmdImpl) Publish(ctx context.Context, channel pubsub.RedisChannelID, message interface{}) error {
	return impl.raw.Publish(ctx, string(channel), message).Err()
}

func (impl *redisCmdImpl) PSubscribeFunc() pubsub.RedisSubscribeRawFunc {
	return impl.psubscribeFunc
}

func (impl *redisCmdImpl) Get(ctx context.Context, key string) (*string, error) {
	value, err := impl.raw.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	return &value, err
}

func (impl *redisCmdImpl) MGet(ctx context.Context, keys ...string) ([]*string, error) {
	if len(keys) == 0 { // Redis does not allow 0-length MGET
		return []*string{}, nil
	}

	raws, err := impl.raw.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}
	result := make([]*string, len(raws))
	for i, raw := range raws {
		if raw == nil {
			result[i] = nil
		} else {
			str := raw.(string)
			result[i] = &str
		}
	}
	return result, nil
}

func (impl *redisCmdImpl) TTL(ctx context.Context, key string) (*time.Duration, error) {
	value, err := impl.raw.TTL(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	return &value, err
}

func (impl *redisCmdImpl) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return impl.raw.Expire(ctx, key, ttl).Err()
}

func (impl *redisCmdImpl) Set(ctx context.Context, key string, value interface{}) error {
	// 0 means no expiration https://github.com/go-redis/redis/issues/143
	return impl.raw.Set(ctx, key, value, 0).Err()
}

func (impl *redisCmdImpl) SetEX(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return impl.raw.SetEX(ctx, key, value, expiration).Err()
}

func (impl *redisCmdImpl) Del(ctx context.Context, key string) error {
	return impl.raw.Del(ctx, key).Err()
}

func (impl *redisCmdImpl) LoadScript(ctx context.Context, script *redis.Script) error {
	return script.Load(ctx, impl.raw).Err()
}

func (impl *redisCmdImpl) RunScript(ctx context.Context, script *redis.Script, keys []string, args ...interface{}) (interface{}, error) {
	return script.Run(ctx, impl.raw, keys, args...).Result()
}
