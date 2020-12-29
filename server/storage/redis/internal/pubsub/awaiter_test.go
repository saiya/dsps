package pubsub_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/saiya/dsps/server/storage/redis/internal/pubsub"
)

func TestResolve(t *testing.T) {
	p := NewPromise()

	assertNotFulfilled(t, p)
	assert.NoError(t, p.Err())

	doInGoroutine(func() { p.Resolve() })
	assertResolved(t, p)
}

func TestRejection(t *testing.T) {
	p := NewPromise()

	assertNotFulfilled(t, p)
	assert.NoError(t, p.Err())

	err := errors.New("test error")
	doInGoroutine(func() { p.Reject(err) })
	assertRejectedWith(t, p, err)
}

func TestDuplicateFulfill(t *testing.T) {
	success := NewPromise()
	doInGoroutine(func() { success.Resolve() })
	doInGoroutine(func() { success.Reject(errors.New("test error")) })
	assertResolved(t, success)

	rejection := NewPromise()
	err := errors.New("test error")
	doInGoroutine(func() { rejection.Reject(err) })
	doInGoroutine(func() { rejection.Resolve() })
	assertRejectedWith(t, rejection, err)
}

func assertNotFulfilled(t *testing.T, p RedisPubSubAwaiter) {
	select {
	default: // OK
	case <-p.Chan():
		assert.Fail(t, "channel should not be closed")
	}
}

func assertResolved(t *testing.T, p RedisPubSubAwaiter) {
	assertFulfilled(t, p)
	assert.NoError(t, p.Err())
}

func assertRejectedWith(t *testing.T, p RedisPubSubAwaiter, err error) {
	assertFulfilled(t, p)
	assert.Same(t, err, p.Err())
}

func assertFulfilled(t *testing.T, p RedisPubSubAwaiter) {
	select {
	case _, open := <-p.Chan():
		assert.False(t, open, "channel should be closed")
	default:
		assert.Fail(t, "channel should be closed")
	}
}

// Call given func in separate goroutine.
// Useful in conjunction with race detector to make sure memory access serialization.
func doInGoroutine(f func()) {
	done := make(chan interface{})
	go func() {
		f()
		close(done)
	}()
	<-done
}
