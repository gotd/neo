package neo

import (
	"sort"
	"sync"
	"time"
)

// Timer abstracts a single event.
type Timer interface {
	C() <-chan time.Time
	Stop() bool
	Reset(d time.Duration)
}

// Ticker abstracts a channel that delivers ``ticks'' of a clock at intervals.
type Ticker interface {
	C() <-chan time.Time
	Stop()
	Reset(d time.Duration)
}

// NewTime returns new temporal simulator.
func NewTime(now time.Time) *Time {
	return &Time{
		now:     now,
		moments: map[int]moment{},
	}
}

// Time simulates temporal interactions.
//
// All methods are goroutine-safe.
type Time struct {
	mux      sync.Mutex
	now      time.Time
	momentID int

	moments   map[int]moment
	observers []chan struct{}
}

func (t *Time) Timer(d time.Duration) Timer {
	done := make(chan time.Time, 1)

	return timer{
		time: t,
		ch:   done,
		id: t.plan(t.When(d), func(now time.Time) {
			done <- now
		}),
	}
}

func (t *Time) Ticker(d time.Duration) Ticker {
	tick := &ticker{
		time: t,
		ch:   make(chan time.Time, 1),
		dur:  d,
	}
	tick.id = t.plan(t.When(d), tick.do)
	return tick
}

func (t *Time) planUnlocked(when time.Time, do func(now time.Time)) int {
	id := t.momentID
	t.momentID++
	t.moments[id] = moment{
		when: when,
		do:   do,
	}
	t.observe()
	return id
}

func (t *Time) plan(when time.Time, do func(now time.Time)) int {
	t.mux.Lock()
	defer t.mux.Unlock()

	return t.planUnlocked(when, do)
}

func (t *Time) stopTimer(id int) bool {
	t.mux.Lock()
	defer t.mux.Unlock()

	_, ok := t.moments[id]
	delete(t.moments, id)
	return ok
}

func (t *Time) resetTimer(d time.Duration, id int, ch chan time.Time) {
	t.mux.Lock()
	defer t.mux.Unlock()

	m, ok := t.moments[id]
	if !ok {
		m = moment{
			do: func(now time.Time) {
				ch <- now
			},
		}
	}

	m.when = t.now.Add(d)
	t.moments[id] = m
}

// resetTicker resets the moment of the given ticker to run after the duration d.
func (t *Time) resetTicker(tick *ticker, d time.Duration) {
	t.mux.Lock()
	defer t.mux.Unlock()

	tick.dur = d
	id := tick.id

	m, ok := t.moments[id]
	if !ok {
		m = moment{do: tick.do}
	}

	m.when = t.now.Add(d)
	t.moments[id] = m
}

// tick applies all scheduled temporal effects.
//
// The mux lock is expected.
func (t *Time) tick() moments {
	var past moments

	for id, m := range t.moments {
		if m.when.After(t.now) {
			continue
		}
		delete(t.moments, id)
		past = append(past, m)
	}
	sort.Sort(past)

	return past
}

// Now returns the current time.
func (t *Time) Now() time.Time {
	t.mux.Lock()
	now := t.now
	t.mux.Unlock()
	return now
}

// Set travels to specified time.
//
// Also triggers temporal effects.
func (t *Time) Set(now time.Time) {
	t.mux.Lock()
	t.now = now
	t.mux.Unlock()
	t.tick().do(now)
}

// Travel adds duration to current time and returns result.
//
// Also triggers temporal effects.
func (t *Time) Travel(d time.Duration) time.Time {
	t.mux.Lock()
	now := t.now.Add(d)
	t.now = now
	t.tick().do(now)
	t.mux.Unlock()
	return now
}

// TravelDate applies AddDate to current time and returns result.
//
// Also triggers temporal effects.
func (t *Time) TravelDate(years, months, days int) time.Time {
	t.mux.Lock()
	now := t.now.AddDate(years, months, days)
	t.now = now
	t.tick().do(now)
	t.mux.Unlock()
	return now
}

// Sleep blocks until duration is elapsed.
func (t *Time) Sleep(d time.Duration) { <-t.After(d) }

// When returns relative time point.
func (t *Time) When(d time.Duration) time.Time {
	return t.Now().Add(d)
}

// After returns new channel that will receive time.Time value with current tme after
// specified duration.
func (t *Time) After(d time.Duration) <-chan time.Time {
	done := make(chan time.Time, 1)
	t.plan(t.When(d), func(now time.Time) {
		done <- now
	})
	return done
}

// Observe return channel that closes on clock calls.
func (t *Time) Observe() <-chan struct{} {
	observer := make(chan struct{})
	t.mux.Lock()
	t.observers = append(t.observers, observer)
	t.mux.Unlock()

	return observer
}

func (t *Time) observe() {
	for _, observer := range t.observers {
		close(observer)
	}
	t.observers = t.observers[:0]
}
