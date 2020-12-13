package lifecycle_test

import (
	"context"
	"sync"
	"testing"
	"time"

	. "github.com/saiya/dsps/server/http/lifecycle"
	. "github.com/saiya/dsps/server/testing"
	"github.com/stretchr/testify/assert"
)

func TestNotClosedServerClose(t *testing.T) {
	sc := NewServerClose()

	wg := sync.WaitGroup{}
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			sc.WithCancel(context.Background(), func(ctx context.Context) {
				assert.NoError(t, ctx.Err())
			})
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestAlreadyClosedServer(t *testing.T) {
	sc := NewServerClose()
	sc.Close()

	sc.WithCancel(context.Background(), func(ctx context.Context) {
		IsError(t, context.Canceled, ctx.Err())
	})
}

func TestServerClosing(t *testing.T) {
	sc := NewServerClose()

	funcCalled := make(chan interface{})
	closerCalled := make(chan interface{})
	go func() {
		<-funcCalled
		sc.Close()
		time.Sleep(1 * time.Millisecond) // Context switch to run internal goroutine
		close(closerCalled)
	}()
	sc.WithCancel(context.Background(), func(ctx context.Context) {
		assert.NoError(t, ctx.Err())
		close(funcCalled)
		<-closerCalled
		IsError(t, context.Canceled, ctx.Err())
	})
}
