package jobs

import (
	"testing"

	"github.com/jjheffernan/BambuFarmer/packages/db"
)

func TestCanCancelFromActiveStates(t *testing.T) {
	for _, from := range []string{db.JobQueued, db.JobAssigned, db.JobDispatched, db.JobRunning, db.JobPaused} {
		if !db.CanTransitionJob(from, db.JobCancelled) {
			t.Fatalf("%s → cancelled should be allowed", from)
		}
	}
	for _, from := range []string{db.JobSucceeded, db.JobFailed, db.JobCancelled} {
		if db.CanTransitionJob(from, db.JobCancelled) && from != db.JobCancelled {
			t.Fatalf("%s → cancelled should be rejected", from)
		}
	}
}
