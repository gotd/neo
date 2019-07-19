// Package neo implements side effects simulation for testing.
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
}

// Now returns the current time.
func (t *Time) Now() time.Time {
	t.mux.Lock()
	now := t.now
	t.mux.Unlock()
	return now
}

// Set travels to specified time.
func (t *Time) Set(now time.Time) {
	t.mux.Lock()
	t.now = now
	t.mux.Unlock()
}

// Travel adds duration to current time and returns result.
func (t *Time) Travel(d time.Duration) time.Time {
	t.mux.Lock()
	now := t.now.Add(d)
	t.now = now
	t.mux.Unlock()
	return now
}

// TravelDate applies AddDate to current time and returns result.
func (t *Time) TravelDate(years int, months int, days int) time.Time {
	t.mux.Lock()
	now := t.now.AddDate(years, months, days)
	t.now = now
	t.mux.Unlock()
	return now
}
