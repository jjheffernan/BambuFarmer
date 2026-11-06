package tests

import (
	"testing"

	"github.com/jjheffernan/BambuFarmer/packages/printerstate"
)

func TestParseGcodeState(t *testing.T) {
	cases := []struct {
		in   string
		want printerstate.GcodeState
		ok   bool
	}{
		{"IDLE", printerstate.Idle, true},
		{"RUNNING", printerstate.Running, true},
		{"PAUSED", printerstate.Pause, true},
		{"FINISHED", printerstate.Finish, true},
		{"", printerstate.Idle, true},
		{"BOGUS", "", false},
	}
	for _, c := range cases {
		got, err := printerstate.ParseGcodeState(c.in)
		if c.ok && err != nil {
			t.Fatalf("%q: %v", c.in, err)
		}
		if !c.ok && err == nil {
			t.Fatalf("%q: expected error", c.in)
		}
		if c.ok && got != c.want {
			t.Fatalf("%q: got %s want %s", c.in, got, c.want)
		}
	}
}

func TestHappyPath(t *testing.T) {
	m := printerstate.NewMachine("01P00A000000001")
	steps := []printerstate.GcodeState{
		printerstate.Prepare, printerstate.Running, printerstate.Pause,
		printerstate.Running, printerstate.Finish, printerstate.Idle,
	}
	for _, s := range steps {
		if err := m.Transition(s); err != nil {
			t.Fatalf("%s: %v", s, err)
		}
	}
	if m.State != printerstate.Idle {
		t.Fatalf("want IDLE got %s", m.State)
	}
}

func TestIllegalTransition(t *testing.T) {
	m := printerstate.NewMachine("x")
	if err := m.Transition(printerstate.Finish); err == nil {
		t.Fatal("IDLE→FINISH should fail")
	}
}

func TestObserveForceSkip(t *testing.T) {
	m := printerstate.NewMachine("x")
	from, to, forced, err := m.Observe("FINISH")
	if err != nil {
		t.Fatal(err)
	}
	if from != printerstate.Idle || to != printerstate.Finish || !forced {
		t.Fatalf("got from=%s to=%s forced=%v", from, to, forced)
	}
	if m.State != printerstate.Finish {
		t.Fatalf("state=%s", m.State)
	}
}

func TestBucketFor(t *testing.T) {
	if printerstate.BucketFor(false, printerstate.Idle, false) != printerstate.BucketAttentionNeeded {
		t.Fatal("offline")
	}
	if printerstate.BucketFor(true, printerstate.Running, false) != printerstate.BucketPrinting {
		t.Fatal("printing")
	}
	if printerstate.BucketFor(true, printerstate.Idle, true) != printerstate.BucketReadyToStart {
		t.Fatal("ready")
	}
	if printerstate.BucketFor(true, printerstate.Idle, false) != printerstate.BucketNotInTask {
		t.Fatal("not in task")
	}
	if printerstate.BucketFor(true, printerstate.Finish, false) != printerstate.BucketCompletedTask {
		t.Fatal("completed")
	}
	if printerstate.BucketFor(true, printerstate.Failed, false) != printerstate.BucketAttentionNeeded {
		t.Fatal("failed")
	}
}

func TestCanAcceptPrint(t *testing.T) {
	if !printerstate.CanAcceptPrint(true, printerstate.Idle) {
		t.Fatal("expected accept")
	}
	if printerstate.CanAcceptPrint(true, printerstate.Running) {
		t.Fatal("running should refuse")
	}
}
