package main

import (
	"os"
	"testing"
	"time"
)

func TestCircuitMaxEnv(t *testing.T) {
	t.Setenv("CIRCUIT_MAX", "")
	if circuitMax() != 0 {
		t.Fatal(circuitMax())
	}
	t.Setenv("CIRCUIT_MAX", "3")
	if circuitMax() != 3 {
		t.Fatal(circuitMax())
	}
	t.Setenv("CIRCUIT_MAX", "nope")
	if circuitMax() != 0 {
		t.Fatal(circuitMax())
	}
}

func TestSchedulerStaggerEnv(t *testing.T) {
	t.Setenv("SCHEDULER_STAGGER", "")
	if schedulerStagger() != 0 {
		t.Fatal(schedulerStagger())
	}
	t.Setenv("SCHEDULER_STAGGER", "5s")
	if schedulerStagger() != 5*time.Second {
		t.Fatal(schedulerStagger())
	}
}

func TestSchedulerMaxStartsEnv(t *testing.T) {
	os.Unsetenv("SCHEDULER_MAX_STARTS_PER_TICK")
	if schedulerMaxStarts() != 0 {
		t.Fatal(schedulerMaxStarts())
	}
	t.Setenv("SCHEDULER_MAX_STARTS_PER_TICK", "2")
	if schedulerMaxStarts() != 2 {
		t.Fatal(schedulerMaxStarts())
	}
}
