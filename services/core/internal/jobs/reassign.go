package jobs

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/db"
)

// ReassignOne retargets a job onto external_sn.
// queued/assigned: update SN (assigned stays assigned, no new start_print).
// failed: requeue then set SN so the scheduler can pick it up.
func ReassignOne(ctx context.Context, pool *pgxpool.Pool, siteID, jobID, sn, actor string) (api.Job, error) {
	sn = strings.TrimSpace(sn)
	if sn == "" {
		return api.Job{}, db.Validation("external_sn required")
	}
	job, err := db.GetJob(ctx, pool, siteID, jobID)
	if err != nil {
		return job, err
	}
	switch job.State {
	case db.JobQueued, db.JobAssigned:
		return setJobSN(ctx, pool, siteID, job, sn, actor, "reassign")
	case db.JobFailed:
		queued, err := db.TransitionJob(ctx, pool, siteID, jobID, db.JobQueued, actor, "reassign requeue")
		if err != nil {
			return queued, err
		}
		return setJobSN(ctx, pool, siteID, queued, sn, actor, "reassign")
	default:
		return job, db.Validation("reassign only for queued, assigned, or failed jobs")
	}
}

func setJobSN(ctx context.Context, pool *pgxpool.Pool, siteID string, job api.Job, sn, actor, reason string) (api.Job, error) {
	if job.ExternalSN == sn {
		return job, nil
	}
	var exists bool
	if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM printers WHERE site_id=$1 AND external_sn=$2)`, siteID, sn).Scan(&exists); err != nil {
		return job, err
	}
	if !exists {
		return job, api.NotFound("printer not found")
	}
	printerID := siteID + ":" + sn
	now := time.Now().UTC()
	_, err := pool.Exec(ctx, `
		UPDATE jobs SET external_sn=$1, printer_id=$2, updated_at=$3, error=''
		WHERE id=$4 AND site_id=$5`, sn, printerID, now, job.ID, siteID)
	if err != nil {
		return job, err
	}
	from := job.ExternalSN
	if from == "" {
		from = "(none)"
	}
	_ = appendJobEventSafe(ctx, pool, job.ID, job.State, job.State, actor, reason+": "+from+" → "+sn)
	return db.GetJob(ctx, pool, siteID, job.ID)
}

func appendJobEventSafe(ctx context.Context, pool *pgxpool.Pool, jobID, from, to, actor, reason string) error {
	if actor == "" {
		actor = "api"
	}
	_, err := pool.Exec(ctx, `
		INSERT INTO job_events (job_id, from_state, to_state, actor, reason) VALUES ($1,$2,$3,$4,$5)`,
		jobID, from, to, actor, reason)
	return err
}

// InMaintenance is true when the printer is flagged or tagged for maintenance.
func InMaintenance(p api.PrinterSummary) bool {
	if p.Maintenance {
		return true
	}
	for _, t := range p.Tags {
		if strings.EqualFold(strings.TrimSpace(t), "maintenance") {
			return true
		}
	}
	return false
}
