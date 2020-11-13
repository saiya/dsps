package sync_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	. "github.com/saiya/dsps/server/sync"
	. "github.com/saiya/dsps/server/testing"
)

func TestLock(t *testing.T) {
	ctx := context.Background()
	lock := NewLock()

	unlock, err := lock.Lock(ctx)
	assert.NoError(t, err)
	defer unlock()
}

func TestLockCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	lock := NewLock()

	unlock, err := lock.Lock(ctx)
	assert.NoError(t, err)
	defer unlock()

	var canceled int32 = 0
	time.AfterFunc(30*time.Millisecond, func() {
		atomic.StoreInt32(&canceled, 1)
		cancel()
	})
	_, err = lock.Lock(ctx)
	IsError(t, ErrLockCanceled, err)
	IsError(t, context.Canceled, err)
	assert.Equal(t, "lock canceled due to context canceled", err.Error())
	assert.Equal(t, int32(1), atomic.LoadInt32(&canceled)) // Lock() must return after Context cancellation
}

func TestDoubleUnlock(t *testing.T) {
	ctx := context.Background()
	lock := NewLock()

	unlock, err := lock.Lock(ctx)
	assert.NoError(t, err)
	unlock()
	unlock() // Do nothing

	// Should successfully acquire lock again
	unlock, err = lock.Lock(ctx)
	assert.NoError(t, err)
	unlock()
}
