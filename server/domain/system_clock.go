package domain

import "time"

// SystemClock is an interface to get current time
// Do not use time.Now() directly to make codes testable.
type SystemClock interface {
	Now() Time
}

var realSystemClock = realSystemClockImpl{}

// RealSystemClock is just a time.Now()
var RealSystemClock SystemClock = &realSystemClock

type realSystemClockImpl struct{}

func (*realSystemClockImpl) Now() Time {
	return Time{
		Time: time.Now(),
	}
}
