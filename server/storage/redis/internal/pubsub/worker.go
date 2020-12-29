package pubsub

import (
	"context"
	"sync"

	"github.com/go-redis/redis/v8"
	"golang.org/x/xerrors"

	"github.com/saiya/dsps/server/logger"
)

type worker interface {
	CheckAvailability(ctx context.Context) error

	OnShutdown(f func())
	Shutdown(ctx context.Context)
	ShutdownCorrupted(ctx context.Context)
}

type workerImpl struct {
	handler func(*redis.Message)

	redisPubSub RedisRawPubSub

	shutdownRequestOnce   sync.Once
	shutdownRequestCh     chan interface{}
	workerMainEnded       chan interface{}
	shutdownCompletedOnce sync.Once
	shutdownCompleted     chan interface{}

	shutdownHooksLock sync.Mutex
	shutdownHooks     []func()
}

// Size of channel that redis.PubSub internally creates
const redisPubSubChannelSize = 100

func newWorker(ctx context.Context, psubscribe RedisSubscribeRawFunc, pattern RedisChannelID, handler func(*redis.Message)) (newWorker worker, err error) {
	w := &workerImpl{
		handler:     handler,
		redisPubSub: psubscribe(ctx, pattern),

		shutdownRequestCh: make(chan interface{}),
		workerMainEnded:   make(chan interface{}),
		shutdownCompleted: make(chan interface{}),
	}
	defer func() {
		if err != nil {
			w.ShutdownCorrupted(ctx)
		}
	}()

	subscribeResult, err := w.redisPubSub.Receive(ctx)
	if err != nil {
		err = xerrors.Errorf("Failed to make Redis Pub/Sub subscription: %w", err)
		return
	}
	if subscribeResponse, ok := subscribeResult.(*redis.Subscription); !(ok && subscribeResponse.Kind == "psubscribe") {
		err = xerrors.Errorf("Unexpected response from Redis Pub/Sub subscription: %v", subscribeResult)
		return
	}

	if err = w.CheckAvailability(ctx); err != nil {
		err = xerrors.Errorf(`ping failed just after Redis PubSub connection allocation: %w`, err)
		return
	}

	// Success
	newWorker = w
	go w.workerMain()
	go w.willCallShutdownHooks()
	return
}

func (w *workerImpl) CheckAvailability(ctx context.Context) error {
	if err := w.redisPubSub.Ping(ctx); err != nil {
		return xerrors.Errorf("redis PubSub connection PING failed: %w", err)
	}
	return nil
}

func (w *workerImpl) OnShutdown(f func()) {
	w.shutdownHooksLock.Lock()
	defer w.shutdownHooksLock.Unlock()

	select {
	case <-w.shutdownCompleted: // Must check after lock to avoid race
		f()
	default:
		w.shutdownHooks = append(w.shutdownHooks, f)
	}
}

func (w *workerImpl) willCallShutdownHooks() {
	go func() {
		<-w.shutdownCompleted

		var hooks []func()
		func() {
			w.shutdownHooksLock.Lock()
			defer w.shutdownHooksLock.Unlock()
			hooks = w.shutdownHooks
		}()
		for _, hook := range hooks {
			hook()
		}
	}()
}

func (w *workerImpl) Shutdown(ctx context.Context) {
	w.shutdown(ctx, true)
}

func (w *workerImpl) ShutdownCorrupted(ctx context.Context) {
	w.shutdown(ctx, false)
}

func (w *workerImpl) shutdown(ctx context.Context, block bool) {
	defer w.shutdownCompletedOnce.Do(func() {
		close(w.shutdownCompleted)
	})

	w.shutdownRequestOnce.Do(func() { close(w.shutdownRequestCh) })
	if block {
		<-w.workerMainEnded
	}
	if err := w.redisPubSub.Close(); err != nil {
		logger.Of(ctx).WarnError(logger.CatStorage, "failed to close Redis pub/sub connection", err)
	}
}

func (w *workerImpl) workerMain() {
	defer close(w.workerMainEnded)

	ch := w.redisPubSub.ChannelSize(redisPubSubChannelSize)
	for {
		select {
		case <-w.shutdownRequestCh:
			return
		case msg, alive := <-ch:
			if !alive { // redis.PubSub closes its channel when underlying connection down
				go w.shutdown(context.Background(), true)
				return
			}
			w.handler(msg)
		}
	}
}
