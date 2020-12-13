package lifecycle

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
	return &serverClose{ch: make(chan interface{})}
}

type serverClose struct {
	ch chan interface{}
}

func (sc *serverClose) Close() {
	close(sc.ch)
}

func (sc *serverClose) WithCancel(ctxToWrap context.Context, action func(ctx context.Context)) {
	ctx, cancel := context.WithCancel(ctxToWrap)

	closeWatching := make(chan interface{})
	defer close(closeWatching)
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
