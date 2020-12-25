package sync

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/saiya/dsps/server/sentry"
	"github.com/saiya/dsps/server/telemetry"
)

// DaemonSystem represents controller of daemons
type DaemonSystem struct {
	name         string
	errorHandler DaemonErrorHandler

	daemonsLock sync.Mutex
	daemons     map[string]*Daemon

	telemetry *telemetry.Telemetry
	sentry    sentry.Sentry
}

// DaemonErrorHandler catch error of daemons
type DaemonErrorHandler func(ctx context.Context, name string, err error)

// DaemonFunc is user-defined implementation of the daemon.
// Note: this function must return valid DaemonNextRun struct even if error occurs.
type DaemonFunc func(context.Context) (DaemonNextRun, error)

// DaemonNextRun controls next daemon execution
type DaemonNextRun struct {
	Interval time.Duration
	Abort    bool // true to terminate this daemon
}

// ErrDaemonClosed means daemon has been shutdown
var ErrDaemonClosed = errors.New("dsps.daemon.closed")

// DaemonSystemDeps is dependencies of daemon system.
type DaemonSystemDeps struct {
	Telemetry *telemetry.Telemetry
	Sentry    sentry.Sentry
}

func (deps DaemonSystemDeps) assertValid() error {
	if deps.Telemetry == nil {
		return fmt.Errorf("invalid DaemonSystemDeps: Telemetry should not be nil")
	}
	if deps.Sentry == nil {
		return fmt.Errorf("invalid DaemonSystemDeps: Sentry should not be nil")
	}
	return nil
}

// NewDaemonSystem create new runtime for daemons
func NewDaemonSystem(name string, deps DaemonSystemDeps, errorHandler DaemonErrorHandler) *DaemonSystem {
	if err := deps.assertValid(); err != nil {
		panic(err)
	}
	return &DaemonSystem{
		name:         name,
		errorHandler: errorHandler,
		daemons:      make(map[string]*Daemon, 16),
		telemetry:    deps.Telemetry,
		sentry:       deps.Sentry,
	}
}

func (ds *DaemonSystem) String() string {
	return fmt.Sprintf("DaemonSystem(%s)", ds.name)
}

// Start creates new Daemon instance.
// "name" parameter must be unique in the DaemonSystem.
func (ds *DaemonSystem) Start(name string, f DaemonFunc) *Daemon {
	var d *Daemon
	func() {
		ds.daemonsLock.Lock()
		defer ds.daemonsLock.Unlock()

		if _, found := ds.daemons[name]; found {
			panic(fmt.Sprintf(`daemon "%s" already exists on system "%s"`, name, ds.name))
		}
		d = ds.newDaemon(name, f)
		ds.daemons[name] = d
	}()
	d.start()
	return d
}

// Get returns Daemon instance in this system.
func (ds *DaemonSystem) Get(name string) *Daemon {
	ds.daemonsLock.Lock()
	defer ds.daemonsLock.Unlock()
	return ds.daemons[name]
}

// Shutdown closes this system, block until all daemons end.
func (ds *DaemonSystem) Shutdown(ctx context.Context) error {
	ds.daemonsLock.Lock()
	defer ds.daemonsLock.Unlock()

	for _, d := range ds.daemons {
		d.RequestShutdown()
	}

	wg, childCtx := errgroup.WithContext(ctx)
	for name := range ds.daemons {
		d := ds.daemons[name]
		wg.Go(func() error { return d.WaitUntilShutdown(childCtx) })
	}
	return wg.Wait()
}

// Daemon represents a running daemon
type Daemon struct {
	shutdownCtx        context.Context
	shutdownCtxCloser  context.CancelFunc
	shutdownCompleteCh chan interface{}

	system *DaemonSystem
	name   string

	f DaemonFunc

	cycleEndLock sync.Mutex
	cycleEnd     chan interface{}

	timerUpdateLock sync.Mutex
	timer           *time.Timer
}

func (ds *DaemonSystem) newDaemon(name string, f DaemonFunc) *Daemon {
	shutdownCtx, shutdownCtxCloser := context.WithCancel(context.Background())
	return &Daemon{
		shutdownCtx:        shutdownCtx,
		shutdownCtxCloser:  shutdownCtxCloser,
		shutdownCompleteCh: make(chan interface{}),

		system: ds,
		name:   name,
		f:      f,

		cycleEnd: make(chan interface{}),
	}
}

func (d *Daemon) String() string {
	return fmt.Sprintf("Daemon(%s)", d.name)
}

func (d *Daemon) start() {
	d.timer = time.NewTimer(0)
	go func() {
		defer close(d.shutdownCompleteCh)
		defer d.shutdownCtxCloser()
		for {
			if !d.cycle() {
				return
			}
		}
	}()
}

func (d *Daemon) cycle() bool {
	defer func() {
		d.cycleEndLock.Lock()
		defer d.cycleEndLock.Unlock()
		close(d.cycleEnd)
		d.cycleEnd = make(chan interface{})
	}()

	var timer <-chan time.Time
	func() {
		d.timerUpdateLock.Lock()
		defer d.timerUpdateLock.Unlock()
		if d.timer != nil { // timer == nil if shutdown requested
			timer = d.timer.C
		}
	}()
	select {
	case <-d.shutdownCtx.Done():
		return false // Shutdown requested
	case <-timer:
	}

	fCtx, end := d.system.telemetry.StartDaemonSpan(d.system.sentry.WrapContext(d.shutdownCtx), d.system.name, d.name)
	defer end()

	var nextRun DaemonNextRun
	var fErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				nextRun = DaemonNextRun{
					Abort:    false,
					Interval: 500 * time.Millisecond, // Retry with some delay...
				}

				panicAsError, isErr := r.(error)
				if !isErr {
					panicAsError = fmt.Errorf("panic in background job: %v", r)
				}
				fErr = panicAsError
			}
		}()
		nextRun, fErr = d.f(fCtx)
	}()
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Avoid using logger/tracer/sentry because it may cause panic again.
				fmt.Fprintf(os.Stderr, "panic in background job system: %v\n", r)
			}
		}()
		if fErr != nil && !errors.Is(fErr, context.Canceled) {
			sentry.RecordError(fCtx, fErr)
			d.system.telemetry.RecordError(fCtx, fErr)
			d.system.errorHandler(fCtx, d.name, fErr)
		}
	}()

	d.timerUpdateLock.Lock()
	defer d.timerUpdateLock.Unlock()
	if nextRun.Abort {
		d.timer = nil
		return false // Self-shutdown
	}
	d.timer = time.NewTimer(nextRun.Interval)
	return true
}

// WaitNextCycle block until this daemon runs at least once since now.
// Callback f() is called while waiting cycle, useful to prevent missfire.
// Returns DaemonClosed if daemon has been shutdown.
func (d *Daemon) WaitNextCycle(ctx context.Context, f func()) error {
	var cycleEnd chan interface{}
	func() {
		d.cycleEndLock.Lock()
		defer d.cycleEndLock.Unlock()
		cycleEnd = d.cycleEnd
	}()

	if f != nil {
		f()
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-d.shutdownCtx.Done():
		return ErrDaemonClosed
	case <-cycleEnd:
		return nil
	}
}

// RequestShutdown ask this daemon to stop. Not wait until shutdown.
func (d *Daemon) RequestShutdown() {
	d.shutdownCtxCloser()

	d.timerUpdateLock.Lock()
	defer d.timerUpdateLock.Unlock()
	if d.timer != nil {
		d.timer.Stop()
		d.timer = nil
	}
}

// WaitUntilShutdown blocks until this daemon ends.
func (d *Daemon) WaitUntilShutdown(ctx context.Context) error {
	select {
	case <-d.shutdownCompleteCh:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
