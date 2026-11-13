package tests

import (
	"testing"

	"github.com/jjheffernan/BambuFarmer/packages/db"
)

func TestCanTransitionJobHappyPath(t *testing.T) {
	steps := []struct{ from, to string }{
		{db.JobQueued, db.JobAssigned},
		{db.JobAssigned, db.JobDispatched},
		{db.JobDispatched, db.JobRunning},
		{db.JobRunning, db.JobPaused},
		{db.JobPaused, db.JobRunning},
		{db.JobRunning, db.JobSucceeded},
	}
	for _, s := range steps {
		if !db.CanTransitionJob(s.from, s.to) {
			t.Fatalf("expected %s → %s", s.from, s.to)
		}
	}
}

func TestCanTransitionJobRejects(t *testing.T) {
	if db.CanTransitionJob(db.JobQueued, db.JobSucceeded) {
		t.Fatal("queued → succeeded should be illegal")
	}
	if db.CanTransitionJob(db.JobSucceeded, db.JobRunning) {
		t.Fatal("terminal → running should be illegal")
	}
}

func TestCanTransitionJobIdempotent(t *testing.T) {
	if !db.CanTransitionJob(db.JobRunning, db.JobRunning) {
		t.Fatal("same-state should be allowed")
	}
}

func TestCanTransitionJobEdgeShortcuts(t *testing.T) {
	// Edge may emit progress/finish before an explicit dispatched ack.
	if !db.CanTransitionJob(db.JobAssigned, db.JobRunning) {
		t.Fatal("assigned → running")
	}
	if !db.CanTransitionJob(db.JobDispatched, db.JobSucceeded) {
		t.Fatal("dispatched → succeeded")
	}
}

func TestJobIDFromPayload(t *testing.T) {
	if db.JobIDFromPayload(nil) != "" {
		t.Fatal("nil")
	}
	if got := db.JobIDFromPayload(map[string]any{"job_id": "abc"}); got != "abc" {
		t.Fatalf("got %q", got)
	}
}
