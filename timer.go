package neo

import "time"

type timer struct {
	time *Time
	ch   chan time.Time
	id   int
}

func (t timer) C() <-chan time.Time {
	return t.ch
}

func (t timer) Stop() bool {
	return t.time.stopTimer(t.id)
}

func (t timer) Reset(d time.Duration) {
	t.time.resetTimer(d, t.id, t.ch)
}
