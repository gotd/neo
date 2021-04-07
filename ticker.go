package neo

import (
	"sync/atomic"
	"time"
)

type ticker struct {
	time *Time
	ch   chan time.Time
	id   int
	dur  int64
}

func (t *ticker) C() <-chan time.Time {
	return t.ch
}

func (t *ticker) Stop() {
	t.time.stopTimer(t.id)
}

func (t *ticker) Reset(d time.Duration) {
	atomic.StoreInt64(&t.dur, int64(d))
	t.time.resetTimer(d, t.id, t.ch)
}
