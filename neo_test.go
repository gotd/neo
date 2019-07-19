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
		t.Error("temporal anomaly detected")
	}
}
