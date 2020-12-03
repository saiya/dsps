package util

import (
	"context"
)

// ServerClose provides way to terminate handler when server shutdown.
type ServerClose interface {
	// Wrap given Context, cancel it on server termination.
	WithCancel(ctxToWrap context.Context, action func(ctx context.Context))
	// Terminate handlers.
	Close()
}

// NewServerClose creates ServerClose instance.
func NewServerClose() ServerClose {
	return &serverClose{ch: make(chan bool)}
}

type serverClose struct {
	ch chan bool
}

func (sc *serverClose) Close() {
	go func() {
		for { // Notify to everyone
			sc.ch <- true
		}
	}()
}

func (sc *serverClose) WithCancel(ctxToWrap context.Context, action func(ctx context.Context)) {
	ctx, cancel := context.WithCancel(ctxToWrap)

	closeWatching := make(chan bool, 1)
	defer func() { closeWatching <- true }()
	go func() {
		select {
		case <-closeWatching:
			return
		case <-sc.ch:
			cancel()
			return
		}
	}()

	action(ctx)
}
