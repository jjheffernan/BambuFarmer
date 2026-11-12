package tests

import (
	"testing"

	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/db"
	"github.com/jjheffernan/BambuFarmer/packages/events"
	"github.com/jjheffernan/BambuFarmer/services/core/internal/jobs"
)

func TestHandleJobEventIgnoresCoreTypes(t *testing.T) {
	err := jobs.HandleJobEvent(t.Context(), nil, "default", events.NewEnvelope(events.TypeJobQueued, "core", "default", map[string]any{
		"job_id": "x",
	}))
	if err != nil {
		t.Fatal(err)
	}
	err = jobs.HandleJobEvent(t.Context(), nil, "default", events.NewEnvelope(events.TypeJobAssigned, "core", "default", nil))
	if err != nil {
		t.Fatal(err)
	}
}

func TestHandleJobEventMissingJobID(t *testing.T) {
	err := jobs.HandleJobEvent(t.Context(), nil, "default", events.NewEnvelope(events.TypeJobFinished, "edge", "default", map[string]any{}))
	if err != nil {
		t.Fatal(err)
	}
}

func TestCommandStartPrintConstant(t *testing.T) {
	if events.TypeCommandStartPrint != "command.start_print" {
		t.Fatal(events.TypeCommandStartPrint)
	}
	if events.TypeJobAssigned != "job.assigned" {
		t.Fatal(events.TypeJobAssigned)
	}
}

func TestTransitionTableMatchesDispatchPath(t *testing.T) {
	// Documented happy path used by scheduler + edge acks.
	path := []string{db.JobQueued, db.JobAssigned, db.JobDispatched, db.JobRunning, db.JobSucceeded}
	for i := 0; i < len(path)-1; i++ {
		if !db.CanTransitionJob(path[i], path[i+1]) {
			t.Fatalf("break at %s → %s", path[i], path[i+1])
		}
	}
}

func TestStartPrintPayloadShape(t *testing.T) {
	// Compile-time / API contract for edge: fields edge.bus.ConsumeCommands reads.
	p := jobs.StartPrintPayload{
		JobID: "j1", ExternalSN: "SN1", SN: "SN1", ArtifactID: "art", FileID: "art",
	}
	if p.FileID == "" || p.JobID == "" {
		t.Fatal("required fields")
	}
	_ = api.Job{} // keep api import for future HTTP fixture tests
}
