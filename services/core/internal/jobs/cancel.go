package jobs

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/db"
	"github.com/jjheffernan/BambuFarmer/packages/events"
	"github.com/jjheffernan/BambuFarmer/services/core/internal/printers"
)

// CancelOne transitions a job to cancelled, emits job.cancelled, and stops the
// printer when the job was already on a machine.
func CancelOne(ctx context.Context, pool *pgxpool.Pool, bus *events.Bus, siteID, id, reason string) (api.Job, error) {
	before, err := db.GetJob(ctx, pool, siteID, id)
	if err != nil {
		return before, err
	}
	if reason == "" {
		reason = "operator cancel"
	}
	job, err := db.TransitionJob(ctx, pool, siteID, id, db.JobCancelled, "api", reason)
	if err != nil {
		return job, err
	}
	publishJobCancelled(ctx, bus, siteID, job)
	switch before.State {
	case db.JobAssigned, db.JobDispatched, db.JobRunning, db.JobPaused:
		if sn := job.ExternalSN; sn != "" && bus != nil {
			_, _ = printers.PublishControl(ctx, bus, siteID, api.ControlStop, []string{sn})
		}
	}
	return job, nil
}

func publishJobCancelled(ctx context.Context, bus *events.Bus, siteID string, job api.Job) {
	if bus == nil {
		return
	}
	env := events.NewEnvelope(events.TypeJobCancelled, "core", siteID, map[string]any{
		"job_id": job.ID, "name": job.Name, "external_sn": job.ExternalSN, "state": job.State,
	})
	env.CorrelationID = job.ID
	_, _ = bus.Publish(ctx, events.StreamJob, env)
}
