package internal

import (
	"context"
	"errors"
	"sync"
	"time"

	"golang.org/x/xerrors"

	"github.com/go-redis/redis/v8"

	"github.com/saiya/dsps/server/logger"
)

//go:generate mockgen -source=${GOFILE} -package=mock -destination=./mock/${GOFILE}
// RedisCmd wraps Redis command system
type RedisCmd interface {
	Ping(ctx context.Context) error

	Publish(ctx context.Context, channel RedisChannelID, message interface{}) error
	PSubscribe(ctx context.Context, pattern RedisChannelID) (c chan string, close func() error, err error)

	Get(ctx context.Context, key string) (*string, error)
	MGet(ctx context.Context, keys ...string) ([]*string, error)
	TTL(ctx context.Context, key string) (*time.Duration, error)
	Set(ctx context.Context, key string, value interface{}) error
	SetEX(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Del(ctx context.Context, key string) error

	LoadScript(ctx context.Context, script *redis.Script) error
	RunScript(ctx context.Context, script *redis.Script, keys []string, args ...interface{}) (interface{}, error)
}

// RedisChannelID represents channel ID of Redis Pub/Sub
type RedisChannelID string

// NewRedisCmd creates new RedisCmd instance.
func NewRedisCmd(raw redis.Cmdable, psubscribeFunc RedisSubscribeRawFunc) RedisCmd {
	return &redisCmdImpl{raw: raw, psubscribeFunc: psubscribeFunc}
}

// RedisSubscribeRawFunc represents (P)SUBSCRIBE command implementation.
type RedisSubscribeRawFunc func(ctx context.Context, channel RedisChannelID) *redis.PubSub
type redisCmdImpl struct {
	raw            redis.Cmdable
	psubscribeFunc RedisSubscribeRawFunc
}

func (impl *redisCmdImpl) Ping(ctx context.Context) error {
	_, err := impl.raw.Ping(ctx).Result()
	return err
}

func (impl *redisCmdImpl) Publish(ctx context.Context, channel RedisChannelID, message interface{}) error {
	return impl.raw.Publish(ctx, string(channel), message).Err()
}

func (impl *redisCmdImpl) PSubscribe(ctx context.Context, channel RedisChannelID) (c chan string, closer func() error, err error) {
	redisPubSub := impl.psubscribeFunc(ctx, channel)
	subscribeResult, err := redisPubSub.Receive(ctx)
	if err != nil {
		err = xerrors.Errorf("Failed to make Redis Pub/Sub subscription: %w", err)
		return
	}
	if _, ok := subscribeResult.(*redis.Subscription); !ok {
		err = xerrors.Errorf("Unexpected response from Redis Pub/Sub subscription: %v", subscribeResult)
		return
	}

	c = make(chan string, 16)
	closeCh := make(chan interface{}, 1)
	closeChOnce := sync.Once{}
	closer = func() error {
		closeChOnce.Do(func() { close(closeCh) })
		if err := redisPubSub.Close(); err != nil {
			logger.Of(ctx).WarnError(logger.CatStorage, "Failed to stop Redis Pub/Sub subscription", err)
		}
		return nil
	}
	go func() {
		for {
			select {
			case <-closeCh:
				return
			case msg, alive := <-redisPubSub.Channel():
				if !alive {
					return
				}
				c <- msg.Payload
			}
		}
	}()
	return
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
