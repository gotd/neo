package neo

import (
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
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

func TestTime_After(t *testing.T) {
	now := time.Date(2049, 5, 6, 23, 55, 11, 1034, time.UTC)
	sim := NewTime(now)

	after := sim.After(time.Second)

	select {
	case <-after:
		t.Error("unexpected done")
	default:
	}

	sim.Travel(time.Second*1 + time.Microsecond)
	select {
	case <-after:
	default:
		t.Error("unexpected state")
	}
}

func TestTime_Observe(t *testing.T) {
	now := time.Date(2049, 5, 6, 23, 55, 11, 1034, time.UTC)
	sim := NewTime(now)

	var g errgroup.Group
	observe := sim.Observe()
	g.Go(func() error {
		<-observe
		sim.Travel(time.Second * 2)
		return nil
	})
	g.Go(func() error {
		<-sim.After(time.Second)
		return nil
	})

	if err := g.Wait(); err != nil {
		t.Error(err)
	}
}
