package sync

import (
	"context"
	"fmt"
	"sync"
)

// Lock is an interface of context-aware lock system.
type Lock interface {
	Lock(ctx context.Context) (UnlockFunc, error)
}

// UnlockFunc is an func() that unlock acquired lock.
type UnlockFunc func()

// NewLock creates Lock instance.
func NewLock() Lock {
	return &lockImpl{
		ch: make(chan struct{}, 1),
	}
}

type lockImpl struct {
	ch chan struct{}
}

func (impl *lockImpl) Lock(ctx context.Context) (UnlockFunc, error) {
	select {
	case <-ctx.Done():
		return func() {}, &errLockCanceled{cause: ctx.Err()}
	case impl.ch <- struct{}{}:
		// Lock acquired
		unlockOnce := sync.Once{}
		return func() {
			unlockOnce.Do(func() {
				<-impl.ch
			})
		}, nil
	}
}

// ErrLockCanceled is for error.Is() type checking.
var ErrLockCanceled error = &errLockCanceled{}

type errLockCanceled struct {
	cause error
}

func (e *errLockCanceled) Error() string {
	return fmt.Sprintf("lock canceled due to %v", e.cause)
}
func (e *errLockCanceled) Unwrap() error {
	return e.cause
}
func (e *errLockCanceled) Is(expected error) bool {
	if _, ok := expected.(*errLockCanceled); ok {
		return true
	}
	// Note: errors.Is() dive into Unwrap() as a fallback, thus no need to check it here.
	return false
}
