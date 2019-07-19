package neo

import (
	"testing"
	"time"
)

func TestTime(t *testing.T) {
	now := time.Date(2049, 5, 6, 23, 55, 11, 1034, time.UTC)
	sim := NewTime(now)
	sim.TravelDate(1, 0, 0)
	if sim.Now().Year() != 2050 {
		t.Fatal("temporal anomaly detected")
	}
	sim.Set(now)
	if !sim.Now().Equal(now) {
		t.Fatal("failed to go back in time")
	}
	sim.Travel(time.Second * 10)
	if sim.Now().Second() != 21 {
		t.Fatal("10 second shift failed")
	}
}
