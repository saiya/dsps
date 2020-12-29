package pubsub

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"

	dspstesting "github.com/saiya/dsps/server/testing"
)

func TestWorkerHandlerCall(t *testing.T) {
	ctx := context.Background()
	lastReceived := make(chan *redis.Message, 1024)
	worker, pubsub := newHealthyWorker(t, func(m *redis.Message) {
		lastReceived <- m
	})
	defer worker.Shutdown(ctx)

	{
		msg := &redis.Message{Channel: "channel ID", Payload: "foo bar"}
		pubsub.EnqueueMessage(msg)
		received := <-lastReceived
		assert.Same(t, msg, received)
	}
	{
		pubsub.EnqueueMessage(&redis.Subscription{})
		select {
		case <-lastReceived:
			assert.Fail(t, "should not receive non-message object")
		default: // OK
		}
	}
}

func TestWorkerConnectionDown(t *testing.T) {
	lastReceived := make(chan *redis.Message, 1024)
	_, pubsub := newHealthyWorker(t, func(m *redis.Message) {
		lastReceived <- m
	})
	assert.False(t, pubsub.IsClosed())

	// Ensure background receiver running
	msg := &redis.Message{Channel: "channel ID", Payload: "foo bar"}
	pubsub.EnqueueMessage(msg)
	received := <-lastReceived
	assert.Same(t, msg, received)

	pubsub.CloseChannel()            // Make connection down
	time.Sleep(1 * time.Millisecond) // Switch to background processes
	assert.True(t, pubsub.IsClosed())
}

func TestWorkerNormalShutdown(t *testing.T) {
	ctx := context.Background()
	worker, pubsub := newHealthyWorker(t, nil)

	called1 := int32(0)
	called2 := int32(0)
	worker.OnShutdown(func() { atomic.AddInt32(&called1, 1) })
	worker.OnShutdown(func() { atomic.AddInt32(&called2, 1) })

	worker.Shutdown(ctx)
	assert.True(t, pubsub.IsClosed())
	time.Sleep(1 * time.Millisecond) // switch context to background goroutine that will call hooks
	assert.Equal(t, int32(1), atomic.LoadInt32(&called1))
	assert.Equal(t, int32(1), atomic.LoadInt32(&called2))

	worker.Shutdown(ctx) // No-op
	time.Sleep(1 * time.Millisecond)
	assert.Equal(t, int32(1), atomic.LoadInt32(&called1))
	assert.Equal(t, int32(1), atomic.LoadInt32(&called2))

	called3 := int32(0)
	worker.OnShutdown(func() { atomic.AddInt32(&called3, 1) })
	assert.Equal(t, int32(1), atomic.LoadInt32(&called3)) // Should immediately called
}

func TestWorkerPubSubCloseFailure(t *testing.T) {
	ctx := context.Background()
	worker, pubsub := newHealthyWorker(t, nil)

	pubsub.SetCloseResult(errors.New("test error"))
	assert.False(t, pubsub.IsClosed()) // Not closed yet.

	worker.Shutdown(ctx) // Should success
	assert.True(t, pubsub.IsClosed())
}

func TestWorkerAvailabilityCheck(t *testing.T) {
	ctx := context.Background()
	worker, pubsub := newHealthyWorker(t, nil)
	defer worker.Shutdown(ctx)

	// Success
	pubsub.EnqueuePingResult(1, nil)
	assert.NoError(t, worker.CheckAvailability(ctx))

	// Fail
	err := errors.New("test error")
	pubsub.EnqueuePingResult(1, err)
	dspstesting.IsError(t, err, worker.CheckAvailability(ctx))
}

func TestWorkerInitialReceiveFailure(t *testing.T) {
	pubsub := newRedisRawPubSubStub(t)
	defer func() { assert.True(t, pubsub.IsClosed()) }() // Should be closed on error

	err := errors.New("test error")
	pubsub.EnqueueMessage(err)
	_, actualErr := newWorker(context.Background(), func(ctx context.Context, channel RedisChannelID) RedisRawPubSub {
		return pubsub
	}, "*", func(m *redis.Message) {})
	dspstesting.IsError(t, err, actualErr)
}

func TestWorkerInitialReceiveInvalidMessage(t *testing.T) {
	pubsub := newRedisRawPubSubStub(t)
	defer func() { assert.True(t, pubsub.IsClosed()) }() // Should be closed on error

	pubsub.EnqueueMessage(struct{}{})
	_, actualErr := newWorker(context.Background(), func(ctx context.Context, channel RedisChannelID) RedisRawPubSub {
		return pubsub
	}, "*", func(m *redis.Message) {})
	assert.Regexp(t, `Unexpected response from Redis Pub/Sub subscription`, actualErr.Error())
}

func TestWorkerInitialPingFailure(t *testing.T) {
	pubsub := newRedisRawPubSubStub(t)
	defer func() { assert.True(t, pubsub.IsClosed()) }() // Should be closed on error

	pubsub.EnqueueMessage(&redis.Subscription{ // constructor expects (P)SUBSCRIBE response message
		Kind: "psubscribe",
	})
	err := errors.New("test error")
	pubsub.EnqueuePingResult(1, err)
	_, actualErr := newWorker(context.Background(), func(ctx context.Context, channel RedisChannelID) RedisRawPubSub {
		return pubsub
	}, "*", func(m *redis.Message) {})
	dspstesting.IsError(t, err, actualErr)
}

func newHealthyWorker(t *testing.T, handler func(*redis.Message)) (worker worker, pubsub *redisRawPubSubStub) {
	pubsub = newRedisRawPubSubStub(t).EnqueueDefaultSubscribeMessage()
	pubsub.EnqueuePingResult(1, nil) // constructor calls PING once.

	if handler == nil {
		handler = func(m *redis.Message) {}
	}
	worker, err := newWorker(context.Background(), func(ctx context.Context, channel RedisChannelID) RedisRawPubSub {
		return pubsub
	}, "*", handler)
	assert.NoError(t, err)

	// Register shutdown hook to make sure called
	shutdownCalled := int32(0)
	worker.OnShutdown(func() { atomic.AddInt32(&shutdownCalled, 1) })
	t.Cleanup(func() {
		time.Sleep(1 * time.Millisecond) // switch context to background goroutine that will call hooks
		assert.Equal(t, int32(1), atomic.LoadInt32(&shutdownCalled))
	})
	return
}
