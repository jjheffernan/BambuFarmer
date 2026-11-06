package printerstate

import "testing"

func TestParseGcodeState(t *testing.T) {
	cases := []struct {
		in   string
		want GcodeState
		ok   bool
	}{
		{"IDLE", Idle, true},
		{"RUNNING", Running, true},
		{"PAUSED", Pause, true},
		{"FINISHED", Finish, true},
		{"", Idle, true},
		{"BOGUS", "", false},
	}
	for _, c := range cases {
		got, err := ParseGcodeState(c.in)
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
	m := NewMachine("01P00A000000001")
	steps := []GcodeState{Prepare, Running, Pause, Running, Finish, Idle}
	for _, s := range steps {
		if err := m.Transition(s); err != nil {
			t.Fatalf("%s: %v", s, err)
		}
	}
	if m.State != Idle {
		t.Fatalf("want IDLE got %s", m.State)
	}
}

func TestIllegalTransition(t *testing.T) {
	m := NewMachine("x")
	if err := m.Transition(Finish); err == nil {
		t.Fatal("IDLE→FINISH should fail")
	}
}

func TestObserveForceSkip(t *testing.T) {
	m := NewMachine("x")
	from, to, forced, err := m.Observe("FINISH")
	if err != nil {
		t.Fatal(err)
	}
	if from != Idle || to != Finish || !forced {
		t.Fatalf("got from=%s to=%s forced=%v", from, to, forced)
	}
	if m.State != Finish {
		t.Fatalf("state=%s", m.State)
	}
}

func TestBucketFor(t *testing.T) {
	if BucketFor(false, Idle, false) != BucketAttentionNeeded {
		t.Fatal("offline")
	}
	if BucketFor(true, Running, false) != BucketPrinting {
		t.Fatal("printing")
	}
	if BucketFor(true, Idle, true) != BucketReadyToStart {
		t.Fatal("ready")
	}
	if BucketFor(true, Idle, false) != BucketNotInTask {
		t.Fatal("not in task")
	}
	if BucketFor(true, Finish, false) != BucketCompletedTask {
		t.Fatal("completed")
	}
	if BucketFor(true, Failed, false) != BucketAttentionNeeded {
		t.Fatal("failed")
	}
}

func TestCanAcceptPrint(t *testing.T) {
	if !CanAcceptPrint(true, Idle) {
		t.Fatal("expected accept")
	}
	if CanAcceptPrint(true, Running) {
		t.Fatal("running should refuse")
	}
}
