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
	t.time.stop(t.id)
}

func (t *ticker) Reset(d time.Duration) {
	t.time.reset(d, t.id, t.do, &t.dur)
}

// do is the ticker’s moment callback. It sends the now time to the underlying
// channel and plans a new moment for the next tick. Note that do runs under
// Time’s lock.
func (t *ticker) do(now time.Time) {
	t.ch <- now

	// It is safe to mutate ID without a lock since at most one moment
	// exists for the given ticker and moments run under the Time’s lock.
	//
	// Additionally, while we probably should be resetting the moment with
	// the initial ticker’s ID, it is not possible since that would break
	// backwards compatibility for users that rely on Time’s Observe method
	// to observe ticks.
	//
	//  t.time.resetUnlocked(t.dur, t.id, t.do, nil)
	//
	t.id = t.time.planUnlocked(now.Add(t.dur), t.do)
}
