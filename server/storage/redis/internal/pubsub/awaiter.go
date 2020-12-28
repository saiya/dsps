package pubsub

import "sync"

// RedisPubSubAwaiter is Promise-like object repsresents Pub/Sub message await.
type RedisPubSubAwaiter interface {
	// Chan returns channel that will be closed when new message received or error occurred.
	Chan() chan interface{}
	// After Chan() has been closed, can obtain error object if error occurred.
	Err() error
}

// RedisPubSubPromise supports resolve, reject operations in addition to RedisPubSubAwaiter
type RedisPubSubPromise interface {
	Resolve()
	Reject(err error)

	Chan() chan interface{}
	Err() error
}

type awaiter struct {
	once sync.Once
	c    chan interface{}
	err  error
}

// NewPromise creates new RedisPubSubPromise
func NewPromise() RedisPubSubPromise {
	return &awaiter{
		c: make(chan interface{}),
	}
}

func (a *awaiter) Chan() chan interface{} {
	return a.c
}

func (a *awaiter) Err() error {
	return a.err
}

func (a *awaiter) Resolve() {
	a.once.Do(func() {
		a.err = nil
		close(a.c)
	})
}

func (a *awaiter) Reject(err error) {
	a.once.Do(func() {
		a.err = err
		close(a.c)
	})
}
