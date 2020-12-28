package pubsub

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/saiya/dsps/server/logger"
	"github.com/saiya/dsps/server/sentry"
	"github.com/saiya/dsps/server/storage/deps"
	"github.com/saiya/dsps/server/storage/redis/internal"
	"github.com/saiya/dsps/server/telemetry"
)

// RedisPubSubDispatcher subscribe Redis PubSub with PSUBSCRIBE (wildcard subscription), then broadcast message to redisPubsubAwaiter.
// Because go-redis open/close underlying TCP connection for each subscription, it cause massive TCP CLOSE_WAIT connections if Storage.FetchMessage make SUBSCRIBE for each call.
type RedisPubSubDispatcher interface {
	Await(ctx context.Context, channel internal.RedisChannelID) (RedisPubSubAwaiter, AwaitCancelFunc)
	Shutdown(ctx context.Context)
}

// AwaitCancelFunc is function to cancel awaiter.
type AwaitCancelFunc func(err error)

// ErrClosed represents dispatcher has been closed.
var ErrClosed = fmt.Errorf("Redis PSUBSCRIBE stream closed due to storage shutdown: %w", context.Canceled)

type dispatcher struct {
	telemetry        *telemetry.Telemetry
	sentry           sentry.Sentry
	backgroundLogger logger.Logger

	psubscribe internal.RedisSubscribeRawFunc
	pattern    internal.RedisChannelID

	shutdownOnce sync.Once
	shutdownCh   chan interface{}

	workerLock sync.Mutex
	worker     worker

	awaitersLock  sync.Mutex
	nextAwaiterID awaiterID
	awaiters      map[internal.RedisChannelID]map[awaiterID]RedisPubSubPromise
}

type awaiterID uint64

const reconcileInterval = 1 * time.Minute
const reconcileRetryInterval = 5 * time.Second

// NewDispatcher creates instance
func NewDispatcher(ctx context.Context, deps deps.StorageDeps, psubscribe internal.RedisSubscribeRawFunc, pattern internal.RedisChannelID) RedisPubSubDispatcher {
	d := &dispatcher{
		telemetry:        deps.Telemetry,
		sentry:           deps.Sentry,
		backgroundLogger: logger.Of(context.Background()),

		psubscribe: psubscribe,
		pattern:    pattern,
		shutdownCh: make(chan interface{}),
		awaiters:   make(map[internal.RedisChannelID]map[awaiterID]RedisPubSubPromise),
	}
	go func() {
		intervalTimer := time.NewTimer(1)
		intervalTimer.Stop()       // Just want to initialize variable with valid timer instance.
		defer intervalTimer.Stop() // Cleanup a timer that is created in the loop below
		for {
			var interval time.Duration
			if d.reconcileCycle() {
				interval = reconcileInterval
			} else {
				interval = reconcileRetryInterval
			}

			intervalTimer = time.NewTimer(interval)
			select {
			case <-d.shutdownCh:
				return
			case <-intervalTimer.C: // sleep until next cycle
			}
		}
	}()
	return d
}

func (d *dispatcher) Shutdown(ctx context.Context) {
	d.shutdownOnce.Do(func() { close(d.shutdownCh) })
	d.rejectAll(ErrClosed) // Must after close(chan)
	d.terminateWorker(ctx) // Must after close(chan)
}

func (d *dispatcher) Await(ctx context.Context, channel internal.RedisChannelID) (RedisPubSubAwaiter, AwaitCancelFunc) {
	result := NewPromise()

	d.awaitersLock.Lock()
	defer d.awaitersLock.Unlock()

	select {
	case <-d.shutdownCh: // This check must after awaitersLock to avoid race with rejectAll(Closed)
		logger.Of(ctx).Debugf(logger.CatStorage, `Awaiting on already closed RedisPubSubDispatcher (channel: %s)`, channel)
		result.Reject(ErrClosed)
		return result, func(error) {}
	default:
	}

	id := d.nextAwaiterID
	d.nextAwaiterID++

	chain := d.awaiters[channel]
	if chain == nil {
		chain = make(map[awaiterID]RedisPubSubPromise)
		d.awaiters[channel] = chain
	}
	chain[id] = result

	return result, func(err error) {
		d.reject(channel, id, err)
	}
}

// Reject all awaiters.
func (d *dispatcher) rejectAll(err error) {
	d.backgroundLogger.Debugf(logger.CatStorage, `RedisPubSubDispatcher.rejectAll: %v`, err)

	d.awaitersLock.Lock()
	defer d.awaitersLock.Unlock()

	for _, awaiters := range d.awaiters {
		for _, awaiter := range awaiters {
			awaiter.Reject(err)
		}
	}
	d.awaiters = make(map[internal.RedisChannelID]map[awaiterID]RedisPubSubPromise)
}

func (d *dispatcher) reject(channel internal.RedisChannelID, id awaiterID, err error) {
	d.backgroundLogger.Debugf(logger.CatStorage, `RedisPubSubDispatcher.reject(channel: %s, id: %s): %v`, channel, id, err)

	d.awaitersLock.Lock()
	defer d.awaitersLock.Unlock()

	chain, chainFound := d.awaiters[channel]
	if !chainFound {
		return
	}
	awaiter := chain[id]
	if awaiter == nil {
		return
	}

	awaiter.Reject(err)
	delete(chain, id)
}

func (d *dispatcher) resolve(channel internal.RedisChannelID) {
	d.backgroundLogger.Debugf(logger.CatStorage, `RedisPubSubDispatcher.resolve(channel: %s)`, channel)

	d.awaitersLock.Lock()
	defer d.awaitersLock.Unlock()

	for _, awaiter := range d.awaiters[channel] {
		awaiter.Resolve()
	}
	delete(d.awaiters, channel)
}

func (d *dispatcher) reconcileCycle() bool {
	ctx, ctxEnd := d.telemetry.StartDaemonSpan(d.sentry.WrapContext(context.Background()), "storage.redis", "pubsub-dispatcher")
	defer ctxEnd()

	if !d.checkWorkerLiveness(ctx) {
		if err := d.repairWorker(ctx); err != nil {
			logger.Of(ctx).WarnError(logger.CatStorage, `failed to (re-)establish Redis PSUBSCRIBE stream`, err)
			return false
		}
	}
	return true
}

func (d *dispatcher) checkWorkerLiveness(ctx context.Context) bool {
	var worker worker
	func() {
		d.workerLock.Lock()
		defer d.workerLock.Unlock()
		worker = d.worker
	}()

	if worker == nil {
		return false
	}

	if err := worker.CheckAvailability(ctx); err != nil {
		err = fmt.Errorf("Redis PSUBSCRIBE connection down (may overlooked Redis PUBLISH message lost), subscription interrupted: %w", err)
		logger.Of(ctx).Warnf(logger.CatStorage, `%v`, err)

		// Because subscription connection down, may overlooked some messages.
		// So that notify awaiters that subscription interrupted.
		d.rejectAll(err)
		return false
	}
	return true
}

func (d *dispatcher) terminateWorker(ctx context.Context) {
	d.workerLock.Lock()
	defer d.workerLock.Unlock()

	if d.worker == nil {
		return
	}
	d.worker.Shutdown(ctx)
}

func (d *dispatcher) repairWorker(ctx context.Context) error {
	newWorker, err := newWorker(ctx, d.psubscribe, d.pattern, func(m *redis.Message) {
		d.resolve(internal.RedisChannelID(m.Channel))
	})
	if err != nil {
		return err
	}

	var oldWorker worker
	func() {
		d.workerLock.Lock()
		defer d.workerLock.Unlock()
		oldWorker = d.worker
		d.worker = newWorker
	}()

	if oldWorker != nil {
		oldWorker.ShutdownCorrupted(ctx)
	}
	return nil
}
