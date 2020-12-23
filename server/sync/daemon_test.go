package sync_test

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	. "github.com/saiya/dsps/server/sync"
	"github.com/saiya/dsps/server/telemetry"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
)

func TestDaemonSystem(t *testing.T) {
	testWithTimeout(t, 1*time.Second, func() {
		ds := NewDaemonSystem("test", telemetry.NewEmptyTelemetry(t), ensureNoDaemonError(t))
		defer closeDaemon(t, ds)
		defer closeDaemon(t, ds) // double close should success

		test1count := int32(0)
		test1blocker := make(chan interface{}, 1)
		test1 := ds.Start("test1", func(c context.Context) (DaemonNextRun, error) {
			select {
			case <-c.Done():
				return DaemonNextRun{}, c.Err()
			case <-test1blocker:
			}
			atomic.AddInt32(&test1count, 1)
			return DaemonNextRun{
				Interval: 1 * time.Millisecond,
				Abort:    false,
			}, nil
		})

		test2count := int32(0)
		test2blocker := make(chan interface{}, 1)
		test2 := ds.Start("test2", func(c context.Context) (DaemonNextRun, error) {
			select {
			case <-c.Done():
				return DaemonNextRun{}, c.Err()
			case <-test2blocker:
			}
			atomic.AddInt32(&test2count, 1)
			return DaemonNextRun{
				Interval: 1 * time.Millisecond,
				Abort:    false,
			}, nil
		})

		// Run test1 daemon twice
		// test2 daemon is still blocking, but it should not interfere test1
		for i := 1; i <= 2; i++ {
			assert.NoError(t, test1.WaitNextCycle(context.Background(), func() {
				test1blocker <- struct{}{}
			}))
			assert.Equal(t, int32(i), atomic.LoadInt32(&test1count))
		}

		// Run test2 daemon also
		for i := 1; i <= 2; i++ {
			assert.NoError(t, test2.WaitNextCycle(context.Background(), func() {
				test2blocker <- struct{}{}
			}))
			assert.Equal(t, int32(i), atomic.LoadInt32(&test2count))
		}
		// Close test2, it should close context passed to daemon function
		time.Sleep(3 * time.Millisecond) // Wait until next test2 cycle, blocks on select{}
		test2.RequestShutdown()
		assert.NoError(t, test2.WaitUntilShutdown(context.Background()))
		assert.Equal(t, ErrDaemonClosed, test2.WaitNextCycle(context.Background(), nil))
	})
}

func TestDaemonTelemetry(t *testing.T) {
	invoked := make(chan interface{})
	tr := telemetry.WithStubTracing(t, func(telemetry *telemetry.Telemetry) {
		ds := NewDaemonSystem("test.system", telemetry, func(ctx context.Context, name string, err error) {
			assert.NoError(t, err)
		})
		ds.Start("test.job", func(c context.Context) (DaemonNextRun, error) {
			close(invoked)
			return DaemonNextRun{Abort: true}, nil
		})
		<-invoked
		defer closeDaemon(t, ds)
	})
	tr.OT.AssertSpan(0, trace.SpanKindInternal, "BackgroundJob test.system test.job", map[string]interface{}{
		"dsps.daemon.system": "test.system",
		"dsps.daemon.name":   "test.job",
	})
}

func TestErrorHandler(t *testing.T) {
	receivedErrors := make(chan error)
	ds := NewDaemonSystem("test system", telemetry.NewEmptyTelemetry(t), func(ctx context.Context, name string, err error) {
		assert.Equal(t, "test1", name)
		receivedErrors <- err
	})
	defer closeDaemon(t, ds)

	expectedError := errors.New("test daemon error")
	ds.Start("test1", func(c context.Context) (DaemonNextRun, error) {
		return DaemonNextRun{Interval: 300 * time.Second}, expectedError
	})

	assert.Same(t, expectedError, <-receivedErrors)
}

func TestPanicHandling(t *testing.T) {
	errorHandlerCalled := int32(0)
	ds := NewDaemonSystem("test system", telemetry.NewEmptyTelemetry(t), func(ctx context.Context, name string, err error) {
		atomic.AddInt32(&errorHandlerCalled, 1)
		panic("test error") // Should not stop system
	})
	defer closeDaemon(t, ds)

	invoked := int32(0)
	fulfilled := make(chan interface{})
	ds.Start("test1", func(c context.Context) (DaemonNextRun, error) {
		switch atomic.AddInt32(&invoked, 1) {
		case 1:
			panic("test error") // Should not stop system
		case 2:
			panic(errors.New("test error")) // Should not stop system
		case 3:
			close(fulfilled)
			return DaemonNextRun{Abort: true}, nil
		}
		panic("Unreachable code")
	})

	<-fulfilled
	assert.Equal(t, int32(2), atomic.LoadInt32(&errorHandlerCalled))
	assert.Equal(t, int32(3), atomic.LoadInt32(&invoked))
}

func TestDaemonSelfShutdown(t *testing.T) {
	ds := NewDaemonSystem("test system", telemetry.NewEmptyTelemetry(t), ensureNoDaemonError(t))
	defer closeDaemon(t, ds)

	test1 := ds.Start("test1", func(c context.Context) (DaemonNextRun, error) {
		return DaemonNextRun{Abort: true}, nil
	})
	assert.NoError(t, test1.WaitUntilShutdown(context.Background()))
	assert.Equal(t, ErrDaemonClosed, test1.WaitNextCycle(context.Background(), nil))
}

func TestDaemonShutdownWaitShouldBlock(t *testing.T) { // regression test case
	ds := NewDaemonSystem("test system", telemetry.NewEmptyTelemetry(t), ensureNoDaemonError(t))
	defer closeDaemon(t, ds)

	test1 := ds.Start("test1", func(c context.Context) (DaemonNextRun, error) {
		return DaemonNextRun{Interval: 3 * time.Millisecond}, nil
	})

	// test1 daemon is running, WaitUntilShutdown should block until context timeout
	shutdownWaitCtx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()
	assert.Equal(t, context.DeadlineExceeded, test1.WaitUntilShutdown(shutdownWaitCtx))
}

func TestDaemonShutdownShouldCancelTasks(t *testing.T) {
	testWithTimeout(t, 1*time.Second, func() {
		ds := NewDaemonSystem("test system", telemetry.NewEmptyTelemetry(t), ensureNoDaemonError(t))
		defer closeDaemon(t, ds)

		test1invoked := make(chan interface{})
		test1canceled := make(chan interface{})
		ds.Start("test1", func(ctx context.Context) (DaemonNextRun, error) {
			close(test1invoked)
			<-ctx.Done()
			close(test1canceled)
			return DaemonNextRun{Abort: true}, nil
		})

		<-test1invoked
		assert.NoError(t, ds.Shutdown(context.Background()))
		<-test1canceled
	})
}

func TestDaemonWaitCanceled(t *testing.T) {
	ds := NewDaemonSystem("test system", telemetry.NewEmptyTelemetry(t), ensureNoDaemonError(t))
	defer closeDaemon(t, ds)

	test1 := ds.Start("test1", func(c context.Context) (DaemonNextRun, error) {
		return DaemonNextRun{Interval: 100 * time.Second}, nil
	})

	ctxCanceled, cancel := context.WithCancel(context.Background())
	cancel()

	assert.Equal(t, context.Canceled, test1.WaitNextCycle(ctxCanceled, nil))
	assert.Equal(t, context.Canceled, test1.WaitUntilShutdown(ctxCanceled))
}

func TestGetDaemon(t *testing.T) {
	ds := NewDaemonSystem("test system", telemetry.NewEmptyTelemetry(t), ensureNoDaemonError(t))
	defer closeDaemon(t, ds)

	assert.Nil(t, ds.Get("test1"))
	test1 := ds.Start("test1", func(c context.Context) (DaemonNextRun, error) {
		return DaemonNextRun{Interval: 100 * time.Second}, nil
	})
	assert.Same(t, test1, ds.Get("test1"))
}

func TestDaemonStringer(t *testing.T) {
	ds := NewDaemonSystem("test system", telemetry.NewEmptyTelemetry(t), ensureNoDaemonError(t))
	defer closeDaemon(t, ds)
	assert.Equal(t, "DaemonSystem(test system)", ds.String())

	test1 := ds.Start("test1", func(c context.Context) (DaemonNextRun, error) {
		return DaemonNextRun{Interval: 100 * time.Second}, nil
	})
	assert.Equal(t, "Daemon(test1)", test1.String())
}

func TestDaemonNameDuplicated(t *testing.T) {
	ds := NewDaemonSystem("test system", telemetry.NewEmptyTelemetry(t), ensureNoDaemonError(t))
	defer closeDaemon(t, ds)

	ds.Start("test daemon", func(c context.Context) (DaemonNextRun, error) {
		return DaemonNextRun{}, nil
	})
	assert.Panicsf(t, func() {
		ds.Start("test daemon", func(c context.Context) (DaemonNextRun, error) {
			return DaemonNextRun{}, nil
		})
	}, `daemon "test daemon" already exists on system "test system"`)
}

func ensureNoDaemonError(t *testing.T) DaemonErrorHandler {
	return func(ctx context.Context, name string, err error) {
		assert.NoError(t, err, fmt.Sprintf(`error from daemon "%s"`, name))
	}
}

func closeDaemon(t *testing.T, ds *DaemonSystem) {
	assert.NoError(t, ds.Shutdown(context.Background()))
}

func testWithTimeout(t *testing.T, d time.Duration, f func()) {
	end := make(chan interface{})
	go func() {
		defer close(end)
		f()
	}()

	select {
	case <-time.After(d):
		assert.FailNow(t, fmt.Sprintf("Test timeout after %s", d))
	case <-end: // OK
	}
}
