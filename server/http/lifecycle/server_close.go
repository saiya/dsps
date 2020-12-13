package lifecycle

import (
	"context"
	"sync"
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
	ch        chan interface{}
	closeOnce sync.Once
}

func (sc *serverClose) Close() {
	sc.closeOnce.Do(func() {
		close(sc.ch)
	})
}

func (sc *serverClose) WithCancel(ctxToWrap context.Context, action func(ctx context.Context)) {
	ctx, cancel := context.WithCancel(ctxToWrap)
	select {
	case <-sc.ch: // Already closed
		cancel()
		action(ctx)
		return
	default:
	}

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
