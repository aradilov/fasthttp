package fasthttp

import (
	"sync"
	"sync/atomic"
	"time"
)

var initTimerReset uint64

func GetInitTimerReset() uint64 {
	return atomic.LoadUint64(&initTimerReset)
}

func initTimer(t *time.Timer, timeout time.Duration) *time.Timer {
	if t == nil {
		return time.NewTimer(timeout)
	}
	if t.Reset(timeout) {
		atomic.AddUint64(&initTimerReset, 1)
		return time.NewTimer(timeout)
	}
	return t
}

func stopTimer(t *time.Timer) {
	if !t.Stop() {
		// Collect possibly added time from the channel
		// if timer has been stopped and nobody collected its' value.
		select {
		case <-t.C:
		default:
		}
	}
}

// AcquireTimer returns a time.Timer from the pool and updates it to
// send the current time on its channel after at least timeout.
//
// The returned Timer may be returned to the pool with ReleaseTimer
// when no longer needed. This allows reducing GC load.
func AcquireTimer(timeout time.Duration) *time.Timer {
	v := timerPool.Get()
	if v == nil {
		return time.NewTimer(timeout)
	}
	t := v.(*time.Timer)
	initTimer(t, timeout)
	return t
}

// ReleaseTimer returns the time.Timer acquired via AcquireTimer to the pool
// and prevents the Timer from firing.
//
// Do not access the released time.Timer or read from it's channel otherwise
// data races may occur.
func ReleaseTimer(t *time.Timer) {
	stopTimer(t)
	timerPool.Put(t)
}

var timerPool sync.Pool
