package pubsub

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	storagetesting "github.com/saiya/dsps/server/storage/deps/testing"
)

func TestDispatcherParams(t *testing.T) {
	p := DispatcherParams{
		ReconcileInterval:        123 * time.Second,
		ReconcileRetryInterval:   234 * time.Second,
		ReconcileMinimumInterval: 345 * time.Second,
	}
	assert.Equal(t, p, p.fillDefaults())
	assert.Equal(t, dispatcherParamsDefault, DispatcherParams{}.fillDefaults())
}

func TestDispatcherAwaitCancel(t *testing.T) {
	ctx := context.Background()
	pubsub := newRedisRawPubSubStub(t).EnqueueDefaultSubscribeMessage().EnqueuePingResultForever(nil)
	dispatcher, pubsubActivated := newDispatcher(t, pubsub)
	defer func() {
		dispatcher.Shutdown(ctx)
		time.Sleep(10 * time.Millisecond) // Wait until background processes exits
	}()
	<-pubsubActivated

	err := errors.New("test error")
	cancelBeforeEvent, cancel1 := dispatcher.Await(ctx, "ch-1")
	alive, _ := dispatcher.Await(ctx, "ch-1") // Should not be canceled
	cancelAfterEvent, cancel2 := dispatcher.Await(ctx, "ch-1")
	cancel1(err)
	cancel1(err) // Cancel twice (no-op)
	pubsub.EnqueueEvent("ch-1")
	<-cancelBeforeEvent.Chan()
	<-alive.Chan()
	cancel2(err)
	assert.Same(t, err, cancelBeforeEvent.Err())
	assert.NoError(t, alive.Err())
	assert.NoError(t, cancelAfterEvent.Err())
}

func TestDispatcherAwaitAfterShutdown(t *testing.T) {
	ctx := context.Background()
	pubsub := newRedisRawPubSubStub(t).EnqueueDefaultSubscribeMessage().EnqueuePingResultForever(nil)
	dispatcher, pubsubActivated := newDispatcher(t, pubsub)
	<-pubsubActivated
	time.Sleep(10 * time.Millisecond) // Await reconcile completion

	dispatcher.Shutdown(ctx)
	await, cancel := dispatcher.Await(ctx, "ch-1")
	<-await.Chan()
	assert.Same(t, ErrClosed, await.Err())
	cancel(nil)                       // No-op, should success
	time.Sleep(10 * time.Millisecond) // Wait until background processes exits
}

func TestDispatcherConnectionDown(t *testing.T) {
	ctx := context.Background()
	pubsub1 := newRedisRawPubSubStub(t).EnqueueDefaultSubscribeMessage().EnqueuePingResultForever(nil)
	pubsub2 := newRedisRawPubSubStub(t).EnqueueDefaultSubscribeMessage().EnqueuePingResultForever(errors.New("test pubsub PING failure"))
	pubsub3 := newRedisRawPubSubStub(t).EnqueueDefaultSubscribeMessage().EnqueuePingResultForever(nil)
	dispatcher, pubsubActivated := newDispatcher(t, pubsub1, pubsub2, pubsub3)
	defer func() {
		dispatcher.Shutdown(ctx)
		time.Sleep(10 * time.Millisecond) // Wait until background processes exits
	}()
	<-pubsubActivated

	{ // Successful
		await, _ := dispatcher.Await(ctx, "ch-1")
		pubsub1.EnqueueEvent("ch-1")
		<-await.Chan() // Should receive message
		assert.NoError(t, await.Err())
	}

	{ // Connection down (pubsub1), retry failure (pubsub2), retry success (pubsub3)
		await, _ := dispatcher.Await(ctx, "ch-1")
		pubsub1.CloseChannel()            // Kill pubsub1
		<-pubsubActivated                 // Wait reconcile starts (try pubsub2)
		<-pubsubActivated                 // Wait reconcile starts (try pubsub3)
		time.Sleep(50 * time.Millisecond) // Wait reconcile completion
		<-await.Chan()                    // Should be rejected due to PubSub connection down
		assert.Error(t, await.Err())
		assert.Regexp(t, `Redis PSUBSCRIBE connection down \(may overlooked Redis PUBLISH message lost\), subscription interrupted`, await.Err().Error())
	}

	{ // After recovery
		await, _ := dispatcher.Await(ctx, "ch-1")
		pubsub3.EnqueueEvent("ch-1")
		<-await.Chan() // Should receive message
		assert.NoError(t, await.Err())
	}
}

func newDispatcher(t *testing.T, pubsubStubs ...*redisRawPubSubStub) (RedisPubSubDispatcher, chan *redisRawPubSubStub) {
	activeStub := int32(0)
	stubActivated := make(chan *redisRawPubSubStub, len(pubsubStubs))
	return NewDispatcher(
		context.Background(),
		storagetesting.EmptyDeps(t),
		DispatcherParams{
			ReconcileInterval:        100 * time.Millisecond,
			ReconcileRetryInterval:   100 * time.Millisecond,
			ReconcileMinimumInterval: 1,
		},
		func(ctx context.Context, channel RedisChannelID) RedisRawPubSub {
			i := int(atomic.AddInt32(&activeStub, 1) - 1)
			if i >= len(pubsubStubs) {
				t.Logf("(re-)PSUBSCRIBE requested but no more pubsub stub")
				return newRedisRawPubSubStub(t).EnqueueDefaultSubscribeMessage().EnqueuePingResultForever(nil)
			}
			stubActivated <- pubsubStubs[i]
			return pubsubStubs[i]
		},
		"*",
	), stubActivated
}
