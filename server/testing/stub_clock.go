package testing

import (
	"testing"
	"time"

	"github.com/saiya/dsps/server/domain"
)

// StubClock extends domain.SystemClock interface for testing purpose.
type StubClock interface {
	Now() domain.Time

	Set(time time.Time)
	Add(d time.Duration) domain.Time
}

// NewStubClock creates new clock instance
func NewStubClock(_ *testing.T) StubClock {
	return &stubClock{now: time.Now()}
}

var _ domain.SystemClock = &stubClock{}

type stubClock struct {
	now time.Time
}

func (clk *stubClock) Now() domain.Time {
	return domain.Time{Time: clk.now}
}

func (clk *stubClock) Set(time time.Time) {
	clk.now = time
}

func (clk *stubClock) Add(d time.Duration) domain.Time {
	clk.now = clk.now.Add(d)
	return clk.Now()
}
