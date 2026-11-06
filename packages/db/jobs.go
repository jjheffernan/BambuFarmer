package db

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jjheffernan/BambuFarmer/packages/api"
)

// Job states (MVP subset of DESIGN_SPEC §7).
const (
	JobQueued     = "queued"
	JobDispatched = "dispatched"
	JobRunning    = "running"
	JobPaused     = "paused"
	JobSucceeded  = "succeeded"
	JobFailed     = "failed"
	JobCancelled  = "cancelled"
)

var jobTransitions = map[string]map[string]bool{
	JobQueued:     {JobDispatched: true, JobCancelled: true},
	JobDispatched: {JobRunning: true, JobFailed: true, JobCancelled: true},
	JobRunning:    {JobPaused: true, JobSucceeded: true, JobFailed: true, JobCancelled: true},
	JobPaused:     {JobRunning: true, JobFailed: true, JobCancelled: true},
}

// CreateJob inserts a queued job and event.
func CreateJob(ctx context.Context, pool *pgxpool.Pool, siteID string, req api.CreateJobRequest) (api.Job, error) {
	id := req.ID
	if id == "" {
		id = newID()
	}
	copies := req.Copies
	if copies < 1 {
		copies = 1
	}
	name := req.Name
	if name == "" {
		name = "job-" + id[:8]
	}
	_, err := pool.Exec(ctx, `
		INSERT INTO jobs (id, site_id, name, state, external_sn, artifact_id, copies, est_seconds, est_grams, currency)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		id, siteID, name, JobQueued, req.ExternalSN, nullStr(req.ArtifactID), copies, req.EstSeconds, req.EstGrams, "USD",
	)
	if err != nil {
		return api.Job{}, err
	}
	_ = appendJobEvent(ctx, pool, id, "", JobQueued, "api", "created")
	return GetJob(ctx, pool, siteID, id)
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// GetJob loads one job.
func GetJob(ctx context.Context, pool *pgxpool.Pool, siteID, id string) (api.Job, error) {
	var j api.Job
	var started, finished *time.Time
	var artifactID *string
	err := pool.QueryRow(ctx, `
		SELECT id, site_id, name, state, COALESCE(external_sn,''), COALESCE(artifact_id::text,''),
		       copies, est_seconds, est_grams, cost_estimate, cost_actual, currency, error,
		       created_at, updated_at, started_at, finished_at
		FROM jobs WHERE id=$1 AND site_id=$2`, id, siteID).Scan(
		&j.ID, &j.SiteID, &j.Name, &j.State, &j.ExternalSN, &j.ArtifactID,
		&j.Copies, &j.EstSeconds, &j.EstGrams, &j.CostEstimate, &j.CostActual, &j.Currency, &j.Error,
		&j.CreatedAt, &j.UpdatedAt, &started, &finished,
	)
	if err != nil {
		return j, err
	}
	_ = artifactID
	j.StartedAt = started
	j.FinishedAt = finished
	return j, nil
}

// ListJobs returns recent jobs for a site.
func ListJobs(ctx context.Context, pool *pgxpool.Pool, siteID string, limit int) ([]api.Job, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := pool.Query(ctx, `
		SELECT id, site_id, name, state, COALESCE(external_sn,''), COALESCE(artifact_id::text,''),
		       copies, est_seconds, est_grams, cost_estimate, cost_actual, currency, error,
		       created_at, updated_at, started_at, finished_at
		FROM jobs WHERE site_id=$1 ORDER BY created_at DESC LIMIT $2`, siteID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []api.Job
	for rows.Next() {
		var j api.Job
		var started, finished *time.Time
		if err := rows.Scan(
			&j.ID, &j.SiteID, &j.Name, &j.State, &j.ExternalSN, &j.ArtifactID,
			&j.Copies, &j.EstSeconds, &j.EstGrams, &j.CostEstimate, &j.CostActual, &j.Currency, &j.Error,
			&j.CreatedAt, &j.UpdatedAt, &started, &finished,
		); err != nil {
			return nil, err
		}
		j.StartedAt = started
		j.FinishedAt = finished
		out = append(out, j)
	}
	return out, rows.Err()
}

// TransitionJob moves job state if allowed and records an event.
func TransitionJob(ctx context.Context, pool *pgxpool.Pool, siteID, id, to, actor, reason string) (api.Job, error) {
	j, err := GetJob(ctx, pool, siteID, id)
	if err != nil {
		return j, err
	}
	if j.State == to {
		return j, nil
	}
	if !jobTransitions[j.State][to] {
		return j, fmt.Errorf("invalid job transition %s → %s", j.State, to)
	}
	now := time.Now().UTC()
	_, err = pool.Exec(ctx, `
		UPDATE jobs SET
			state=$1,
			updated_at=$2,
			started_at=CASE WHEN $1='running' AND started_at IS NULL THEN $2 ELSE started_at END,
			finished_at=CASE WHEN $1 IN ('succeeded','failed','cancelled') THEN $2 ELSE finished_at END
		WHERE id=$3 AND site_id=$4`, to, now, id, siteID)
	if err != nil {
		return j, err
	}
	_ = appendJobEvent(ctx, pool, id, j.State, to, actor, reason)
	return GetJob(ctx, pool, siteID, id)
}

// ApplyJobCost writes estimate or actual cost fields.
func ApplyJobCost(ctx context.Context, pool *pgxpool.Pool, siteID, id string, estimate, actual float64, currency string) error {
	if currency == "" {
		currency = "USD"
	}
	_, err := pool.Exec(ctx, `
		UPDATE jobs SET cost_estimate=$1, cost_actual=$2, currency=$3, updated_at=now()
		WHERE id=$4 AND site_id=$5`, estimate, actual, currency, id, siteID)
	return err
}

func appendJobEvent(ctx context.Context, pool *pgxpool.Pool, jobID, from, to, actor, reason string) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO job_events (job_id, from_state, to_state, actor, reason) VALUES ($1,$2,$3,$4,$5)`,
		jobID, from, to, actor, reason)
	return err
}

// UpsertPrinterProjection stores last-known printer state from edge events.
func UpsertPrinterProjection(ctx context.Context, pool *pgxpool.Pool, siteID string, sn, name, model, state string, online bool, pct, remain int, snapshot []byte) error {
	id := siteID + ":" + sn
	_, err := pool.Exec(ctx, `
		INSERT INTO printers (id, site_id, external_sn, name, model, online, gcode_state, mc_percent, mc_remaining_time, last_snapshot, last_seen_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10::jsonb, now(), now())
		ON CONFLICT (site_id, external_sn) DO UPDATE SET
			name=EXCLUDED.name, model=EXCLUDED.model, online=EXCLUDED.online,
			gcode_state=EXCLUDED.gcode_state, mc_percent=EXCLUDED.mc_percent,
			mc_remaining_time=EXCLUDED.mc_remaining_time, last_snapshot=EXCLUDED.last_snapshot,
			last_seen_at=now(), updated_at=now()`,
		id, siteID, sn, name, model, online, state, pct, remain, string(snapshot),
	)
	return err
}

func newID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

// ListPrinters returns projected printers.
func ListPrinters(ctx context.Context, pool *pgxpool.Pool, siteID string) ([]api.PrinterSummary, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, site_id, external_sn, name, model, online, gcode_state, mc_percent, mc_remaining_time
		FROM printers WHERE site_id=$1 ORDER BY name, external_sn`, siteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []api.PrinterSummary
	for rows.Next() {
		var p api.PrinterSummary
		if err := rows.Scan(&p.ID, &p.SiteID, &p.ExternalSN, &p.Name, &p.Model, &p.Online, &p.GcodeState, &p.MCPercent, &p.RemainingS); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
