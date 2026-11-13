package jobs

import (
	"testing"
	"time"
)

func TestStartsCap(t *testing.T) {
	if (&Scheduler{}).startsCap() != 0 {
		t.Fatal("default unlimited")
	}
	if (&Scheduler{MaxStartsPerTick: 3}).startsCap() != 3 {
		t.Fatal("explicit max")
	}
	if (&Scheduler{Stagger: 5 * time.Second}).startsCap() != 1 {
		t.Fatal("stagger implies one start per tick")
	}
	if (&Scheduler{Stagger: time.Second, MaxStartsPerTick: 4}).startsCap() != 4 {
		t.Fatal("explicit max wins over stagger")
	}
}

func TestTickIntervalRespectsStagger(t *testing.T) {
	s := &Scheduler{Interval: 3 * time.Second, Stagger: 10 * time.Second}
	if s.tickInterval() != 10*time.Second {
		t.Fatal(s.tickInterval())
	}
	s = &Scheduler{Interval: 5 * time.Second, Stagger: 2 * time.Second}
	if s.tickInterval() != 5*time.Second {
		t.Fatal(s.tickInterval())
	}
}
