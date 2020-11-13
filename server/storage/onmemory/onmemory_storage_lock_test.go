package onmemory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/sync"
	. "github.com/saiya/dsps/server/testing"
)

func TestShutdownLockFail(t *testing.T) {
	testLockFail(t, "10ms", func(ctx context.Context, storage *onmemoryStorage) error {
		return storage.Shutdown(ctx)
	})
}

func TestProbeLockFail(t *testing.T) {
	var lockFailureError error
	testLockFail(t, "10ms", func(ctx context.Context, storage *onmemoryStorage) error {
		_, err := storage.Liveness(ctx)
		lockFailureError = err
		return err
	})
	testLockFail(t, "10ms", func(ctx context.Context, storage *onmemoryStorage) error {
		_, err := storage.Readiness(ctx)
		assert.NoError(t, err, "Readiness should success even if dead locked")
		return lockFailureError // Mimic test utility
	})
}

func TestPubSubLockFail(t *testing.T) {
	sl := domain.SubscriberLocator{
		ChannelID:    "ch1",
		SubscriberID: "s1",
	}

	testLockFail(t, "10ms", func(ctx context.Context, storage *onmemoryStorage) error {
		return storage.NewSubscriber(ctx, sl)
	})
	testLockFail(t, "10ms", func(ctx context.Context, storage *onmemoryStorage) error {
		return storage.RemoveSubscriber(ctx, sl)
	})
	testLockFail(t, "10ms", func(ctx context.Context, storage *onmemoryStorage) error {
		return storage.PublishMessages(ctx, []domain.Message{})
	})
	func() { // Test FetchMessages, lock failure before polling
		s := makeRawStorage(t)
		defer func() { assert.NoError(t, s.Shutdown(context.Background())) }()
		assert.NoError(t, s.NewSubscriber(context.Background(), sl))

		unlock, err := s.lock.Lock(context.Background()) // Make deadlock
		assert.NoError(t, err)
		defer unlock()

		ctx, cancel := context.WithTimeout(context.Background(), MakeDuration("10ms").Duration)
		defer cancel()
		_, _, _, err = s.FetchMessages(ctx, sl, 100, MakeDuration("100s"))
		IsError(t, sync.ErrLockCanceled, err)
	}()
	testLockFail(t, "10ms", func(ctx context.Context, storage *onmemoryStorage) error {
		return storage.AcknowledgeMessages(ctx, domain.AckHandle{})
	})
	testLockFail(t, "10ms", func(ctx context.Context, storage *onmemoryStorage) error {
		_, err := storage.IsOldMessages(ctx, sl, []domain.MessageLocator{{}})
		return err
	})
}

func TestJwtLockFail(t *testing.T) {
	exp := domain.JwtExp{}
	jti := domain.JwtJti("foobar")
	testLockFail(t, "10ms", func(ctx context.Context, storage *onmemoryStorage) error {
		return storage.AsJwtStorage().RevokeJwt(ctx, exp, jti)
	})
	testLockFail(t, "10ms", func(ctx context.Context, storage *onmemoryStorage) error {
		_, err := storage.AsJwtStorage().IsRevokedJwt(ctx, jti)
		return err
	})
}

func TestGCLockFail(t *testing.T) {
	testLockFail(t, "10ms", func(ctx context.Context, storage *onmemoryStorage) error {
		return storage.GC(ctx)
	})
}

func testLockFail(t *testing.T, duration string, f func(ctx context.Context, storage *onmemoryStorage) error) {
	s, cleanup := makeDeadLocked(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), MakeDuration(duration).Duration)
	defer cancel()

	err := f(ctx, s)
	IsError(t, sync.ErrLockCanceled, err)
	IsError(t, context.DeadlineExceeded, err)
}

// Build onmemoryStorage and lock it's internal state.
// Must call cleanup func to shutdown it properly.
func makeDeadLocked(t *testing.T) (storage *onmemoryStorage, cleanup func()) {
	raw := makeRawStorage(t)
	if raw == nil {
		return nil, func() {}
	}

	unlock, err := raw.lock.Lock(context.Background())
	assert.NoError(t, err)
	return raw, func() {
		unlock()
		assert.NoError(t, raw.Shutdown(context.Background()))
	}
}
