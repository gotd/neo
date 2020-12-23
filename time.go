package neo

import (
	"sync"
	"time"
)

// NewTime returns new temporal simulator.
func NewTime(now time.Time) *Time {
	return &Time{
		now: now,
	}
}

// Time simulates temporal interactions.
//
// All methods are goroutine-safe.
type Time struct {
	mux sync.Mutex
	now time.Time

	moments   []moment
	observers []chan struct{}
}

func (t *Time) plan(when time.Time, do func(now time.Time)) {
	t.mux.Lock()
	defer t.mux.Unlock()

	t.moments = append(t.moments, moment{
		when: when,
		do:   do,
	})
	t.observe()
}

// tick applies all scheduled temporal effects.
//
// The mux lock is expected.
func (t *Time) tick() {
	var (
		past   []moment
		future []moment
	)
	for _, m := range t.moments {
		switch {
		case m.when.After(t.now):
			future = append(future, m)
		case m.when.Before(t.now):
			past = append(past, m)
		}
	}

	t.moments = future
	for _, m := range past {
		m.do(t.now)
	}
}

type moment struct {
	when time.Time
	do   func(time time.Time)
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
	t.tick()
	t.mux.Unlock()
}

// Travel adds duration to current time and returns result.
//
// Also triggers temporal effects.
func (t *Time) Travel(d time.Duration) time.Time {
	t.mux.Lock()
	now := t.now.Add(d)
	t.now = now
	t.tick()
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
	t.tick()
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
