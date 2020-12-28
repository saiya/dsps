package stub

import (
	"context"
	"sync"

	"github.com/saiya/dsps/server/storage/redis/internal"
	"github.com/saiya/dsps/server/storage/redis/internal/pubsub"
)

// RedisPubSubDispatcherStub implements RedisPubSubDispatcher
type RedisPubSubDispatcherStub struct {
	lock     sync.Mutex
	isClosed bool

	hooks    map[internal.RedisChannelID][]func(pubsub.RedisPubSubPromise)
	promises map[internal.RedisChannelID][]pubsub.RedisPubSubPromise
}

// NewRedisPubSubDispatcherStub returns new RedisPubSubDispatcherStub
func NewRedisPubSubDispatcherStub() *RedisPubSubDispatcherStub {
	return &RedisPubSubDispatcherStub{
		hooks:    make(map[internal.RedisChannelID][]func(pubsub.RedisPubSubPromise)),
		promises: make(map[internal.RedisChannelID][]pubsub.RedisPubSubPromise),
	}
}

// Shutdown implements RedisPubSubDispatcher
func (d *RedisPubSubDispatcherStub) Shutdown(ctx context.Context) {
	func() {
		d.lock.Lock()
		defer d.lock.Unlock()
		d.isClosed = true
	}()
	d.Reject(pubsub.ErrClosed)
}

// HookAwaitOnce is to register hook to future Await() call.
func (d *RedisPubSubDispatcherStub) HookAwaitOnce(channel internal.RedisChannelID, f func(pubsub.RedisPubSubPromise)) {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.hooks[channel] = append(d.hooks[channel], f)
}

// Await implements RedisPubSubDispatcher
func (d *RedisPubSubDispatcherStub) Await(ctx context.Context, channel internal.RedisChannelID) (pubsub.RedisPubSubAwaiter, pubsub.AwaitCancelFunc) {
	p := pubsub.NewPromise()
	var hooks []func(pubsub.RedisPubSubPromise)
	defer func() { // Make sure to call hooks on the end of function
		// Call hooks outside of lock, otherwise may cause deadlock (recursive lock)
		for _, hook := range hooks {
			hook(p)
		}
	}()

	d.lock.Lock()
	defer d.lock.Unlock()

	hooks = d.hooks[channel]
	delete(d.hooks, channel)

	if d.isClosed {
		p.Reject(pubsub.ErrClosed)
		return p, func(error) {}
	}

	d.promises[channel] = append(d.promises[channel], p)

	return p, func(err error) {
		p.Reject(err)
		// Because this is stub implementation, not take care memory leak.
	}
}

// Reject all promises
func (d *RedisPubSubDispatcherStub) Reject(err error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	for _, chain := range d.promises {
		for _, p := range chain {
			p.Reject(err)
		}
	}
}

// Resolve all promises of the channel
func (d *RedisPubSubDispatcherStub) Resolve(channel internal.RedisChannelID) {
	d.lock.Lock()
	defer d.lock.Unlock()

	for _, p := range d.promises[channel] {
		p.Resolve()
	}
}
