package pubsub

import (
	"context"
	"sync"

	"github.com/go-redis/redis/v8"
	"golang.org/x/xerrors"

	"github.com/saiya/dsps/server/logger"
	"github.com/saiya/dsps/server/storage/redis/internal"
)

type worker interface {
	CheckAvailability(ctx context.Context) error
	Shutdown(ctx context.Context)
	ShutdownCorrupted(ctx context.Context)
}

type workerImpl struct {
	handler func(*redis.Message)

	redisPubSub *redis.PubSub

	shutdownRequestOnce sync.Once
	shutdownRequestCh   chan interface{}
	shutdownCompleteCh  chan interface{}
}

func newWorker(ctx context.Context, psubscribe internal.RedisSubscribeRawFunc, pattern internal.RedisChannelID, handler func(*redis.Message)) (newWorker worker, err error) {
	w := &workerImpl{
		handler:     handler,
		redisPubSub: psubscribe(ctx, pattern),

		shutdownRequestCh:  make(chan interface{}),
		shutdownCompleteCh: make(chan interface{}),
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
	return
}

func (w *workerImpl) CheckAvailability(ctx context.Context) error {
	if err := w.redisPubSub.Ping(ctx); err != nil {
		return xerrors.Errorf("redis PubSub connection PING failed: %w", err)
	}
	return nil
}

func (w *workerImpl) Shutdown(ctx context.Context) {
	w.shutdown(ctx, true)
}

func (w *workerImpl) ShutdownCorrupted(ctx context.Context) {
	w.shutdown(ctx, false)
}

func (w *workerImpl) shutdown(ctx context.Context, block bool) {
	w.shutdownRequestOnce.Do(func() { close(w.shutdownRequestCh) })
	if block {
		<-w.shutdownCompleteCh
	}
	if err := w.redisPubSub.Close(); err != nil {
		logger.Of(ctx).WarnError(logger.CatStorage, "failed to close Redis pub/sub connection", err)
	}
}

func (w *workerImpl) workerMain() {
	defer close(w.shutdownCompleteCh)
	for {
		select {
		case <-w.shutdownRequestCh:
			return
		case msg, alive := <-w.redisPubSub.Channel():
			if !alive {
				return
			}
			w.handler(msg)
		}
	}
}
