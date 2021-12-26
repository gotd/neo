package neo

import (
	"time"
)

type ticker struct {
	time *Time
	ch   chan time.Time
	id   int
	dur  time.Duration
}

func (t *ticker) C() <-chan time.Time {
	return t.ch
}

func (t *ticker) Stop() {
	t.time.stopTimer(t.id)
}

func (t *ticker) Reset(d time.Duration) {
	t.time.resetTicker(t, d)
}

// do is the ticker’s moment callback. It sends the now time to the underlying
// channel and plans a new moment for the next tick. Note that do runs under
// Time’s lock.
func (t *ticker) do(now time.Time) {
	t.ch <- now

	// It is safe to mutate ID without a lock since at most one
	// moment exists for the given ticker and moments run under
	// the Time’s lock.
	t.id = t.time.planUnlocked(now.Add(t.dur), t.do)
}
