package db

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jjheffernan/BambuFarmer/packages/api"
)

// Job states (MVP subset of DESIGN_SPEC §7).
const (
	JobQueued     = "queued"
	JobAssigned   = "assigned"
	JobDispatched = "dispatched"
	JobRunning    = "running"
	JobPaused     = "paused"
	JobSucceeded  = "succeeded"
	JobFailed     = "failed"
	JobCancelled  = "cancelled"
)

// Active print states occupy a printer (not eligible for new assign).
var jobBusyStates = []string{JobAssigned, JobDispatched, JobRunning, JobPaused}

var jobTransitions = map[string]map[string]bool{
	JobQueued:     {JobAssigned: true, JobCancelled: true},
	JobAssigned:   {JobDispatched: true, JobRunning: true, JobFailed: true, JobCancelled: true, JobQueued: true},
	JobDispatched: {JobRunning: true, JobPaused: true, JobSucceeded: true, JobFailed: true, JobCancelled: true},
	JobRunning:    {JobPaused: true, JobSucceeded: true, JobFailed: true, JobCancelled: true},
	JobPaused:     {JobRunning: true, JobSucceeded: true, JobFailed: true, JobCancelled: true},
}

// CanTransitionJob reports whether from→to is a legal job state change.
func CanTransitionJob(from, to string) bool {
	if from == to {
		return true
	}
	return jobTransitions[from][to]
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
		return j, MapQueryError(err)
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

// ListQueuedJobs returns oldest queued jobs (scheduler input).
func ListQueuedJobs(ctx context.Context, pool *pgxpool.Pool, siteID string, limit int) ([]api.Job, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := pool.Query(ctx, `
		SELECT id, site_id, name, state, COALESCE(external_sn,''), COALESCE(artifact_id::text,''),
		       copies, est_seconds, est_grams, cost_estimate, cost_actual, currency, error,
		       created_at, updated_at, started_at, finished_at
		FROM jobs WHERE site_id=$1 AND state=$2
		ORDER BY created_at ASC LIMIT $3`, siteID, JobQueued, limit)
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

// ListIdlePrinters returns online IDLE printers with no busy job.
func ListIdlePrinters(ctx context.Context, pool *pgxpool.Pool, siteID string) ([]api.PrinterSummary, error) {
	rows, err := pool.Query(ctx, `
		SELECT p.id, p.site_id, p.external_sn, p.name, p.model, p.online, p.gcode_state, p.mc_percent, p.mc_remaining_time
		FROM printers p
		WHERE p.site_id=$1
		  AND p.online = TRUE
		  AND (p.gcode_state = '' OR upper(p.gcode_state) = 'IDLE')
		  AND NOT EXISTS (
			SELECT 1 FROM jobs j
			WHERE j.site_id = p.site_id
			  AND j.external_sn = p.external_sn
			  AND j.state = ANY($2)
		  )
		ORDER BY p.name, p.external_sn`, siteID, jobBusyStates)
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

// AssignJob claims a queued job onto a printer SN and moves state to assigned.
// preferredSN may be empty to use job.ExternalSN; if both empty, returns Validation.
func AssignJob(ctx context.Context, pool *pgxpool.Pool, siteID, jobID, preferredSN, actor, reason string) (api.Job, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return api.Job{}, err
	}
	defer tx.Rollback(ctx)

	var j api.Job
	var started, finished *time.Time
	err = tx.QueryRow(ctx, `
		SELECT id, site_id, name, state, COALESCE(external_sn,''), COALESCE(artifact_id::text,''),
		       copies, est_seconds, est_grams, cost_estimate, cost_actual, currency, error,
		       created_at, updated_at, started_at, finished_at
		FROM jobs WHERE id=$1 AND site_id=$2 FOR UPDATE`, jobID, siteID).Scan(
		&j.ID, &j.SiteID, &j.Name, &j.State, &j.ExternalSN, &j.ArtifactID,
		&j.Copies, &j.EstSeconds, &j.EstGrams, &j.CostEstimate, &j.CostActual, &j.Currency, &j.Error,
		&j.CreatedAt, &j.UpdatedAt, &started, &finished,
	)
	if err != nil {
		return api.Job{}, MapQueryError(err)
	}
	j.StartedAt = started
	j.FinishedAt = finished

	if j.State == JobAssigned && (preferredSN == "" || preferredSN == j.ExternalSN) {
		return j, tx.Commit(ctx)
	}
	if j.State != JobQueued {
		return j, InvalidJobTransition(j.State, JobAssigned)
	}

	sn := preferredSN
	if sn == "" {
		sn = j.ExternalSN
	}
	if sn == "" {
		return j, Validation("external_sn required to assign job")
	}

	now := time.Now().UTC()
	var printerID any
	var exists bool
	_ = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM printers WHERE site_id=$1 AND external_sn=$2)`, siteID, sn).Scan(&exists)
	if exists {
		printerID = siteID + ":" + sn
	}
	_, err = tx.Exec(ctx, `
		UPDATE jobs SET
			state=$1, external_sn=$2, printer_id=$3, updated_at=$4, error=''
		WHERE id=$5 AND site_id=$6`,
		JobAssigned, sn, printerID, now, jobID, siteID)
	if err != nil {
		return j, err
	}
	if reason == "" {
		reason = "assigned"
	}
	if actor == "" {
		actor = "scheduler"
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO job_events (job_id, from_state, to_state, actor, reason) VALUES ($1,$2,$3,$4,$5)`,
		jobID, JobQueued, JobAssigned, actor, reason)
	if err != nil {
		return j, err
	}
	if err := tx.Commit(ctx); err != nil {
		return j, err
	}
	return GetJob(ctx, pool, siteID, jobID)
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
	if !CanTransitionJob(j.State, to) {
		return j, InvalidJobTransition(j.State, to)
	}
	now := time.Now().UTC()
	_, err = pool.Exec(ctx, `
		UPDATE jobs SET
			state=$1,
			updated_at=$2,
			started_at=CASE WHEN $1 IN ('running','dispatched') AND started_at IS NULL THEN $2 ELSE started_at END,
			finished_at=CASE WHEN $1 IN ('succeeded','failed','cancelled') THEN $2 ELSE finished_at END
		WHERE id=$3 AND site_id=$4`, to, now, id, siteID)
	if err != nil {
		return j, err
	}
	_ = appendJobEvent(ctx, pool, id, j.State, to, actor, reason)
	return GetJob(ctx, pool, siteID, id)
}

// SetJobError stores a failure message without changing state.
func SetJobError(ctx context.Context, pool *pgxpool.Pool, siteID, id, msg string) error {
	_, err := pool.Exec(ctx, `
		UPDATE jobs SET error=$1, updated_at=now() WHERE id=$2 AND site_id=$3`, msg, id, siteID)
	return err
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

// JobIDFromPayload extracts job_id from an event payload (string or empty).
func JobIDFromPayload(payload map[string]any) string {
	if payload == nil {
		return ""
	}
	if s, ok := payload["job_id"].(string); ok && s != "" {
		return s
	}
	return ""
}

// NormalizeGcodeState uppercases a telemetry state string for comparisons.
func NormalizeGcodeState(s string) string {
	return strings.ToUpper(strings.TrimSpace(s))
}
