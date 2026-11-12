package jobs

import (
	"context"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/db"
	"github.com/jjheffernan/BambuFarmer/packages/events"
	"github.com/jjheffernan/BambuFarmer/services/core/internal/notify"
)

// ConsumeJobEvents advances the job state machine from bf.events.job (edge acks).
func ConsumeJobEvents(ctx context.Context, bus *events.Bus, pool *pgxpool.Pool, siteID, consumer string) {
	if pool == nil || bus == nil {
		return
	}
	if consumer == "" {
		consumer = "core-jobs-1"
	}
	err := bus.ConsumeLoop(ctx, events.StreamJob, events.GroupCore, consumer, "$", func(ctx context.Context, _ string, env events.Envelope) error {
		return HandleJobEvent(ctx, pool, siteID, env, bus)
	})
	if err != nil && ctx.Err() == nil {
		slog.Warn("job consumer stopped", "err", err)
	}
}

// HandleJobEvent applies one inbound job.* envelope (idempotent transitions).
func HandleJobEvent(ctx context.Context, pool *pgxpool.Pool, siteID string, env events.Envelope, bus *events.Bus) error {
	// Ignore our own publishes (queued/assigned) — only edge lifecycle advances DB here.
	switch env.Type {
	case events.TypeJobQueued, events.TypeJobAssigned, events.TypeJobCancelled:
		return nil
	}

	jobID := db.JobIDFromPayload(env.Payload)
	if jobID == "" {
		jobID = env.CorrelationID
	}
	if jobID == "" {
		return nil
	}
	sid := env.SiteID
	if sid == "" {
		sid = siteID
	}

	switch env.Type {
	case events.TypeJobDispatched:
		_, err := db.TransitionJob(ctx, pool, sid, jobID, db.JobDispatched, "edge", "dispatched")
		return ignoreInvalid(err)
	case events.TypeJobProgress:
		return applyProgress(ctx, pool, sid, jobID, env.Payload)
	case events.TypeJobFinished:
		return applyFinished(ctx, pool, bus, sid, jobID, env.Payload)
	case events.TypeJobFailed:
		msg, _ := env.Payload["error"].(string)
		if msg == "" {
			msg = "failed"
		}
		_ = db.SetJobError(ctx, pool, sid, jobID, msg)
		_, err := db.TransitionJob(ctx, pool, sid, jobID, db.JobFailed, "edge", msg)
		return ignoreInvalid(err)
	default:
		return nil
	}
}

func applyProgress(ctx context.Context, pool *pgxpool.Pool, siteID, jobID string, payload map[string]any) error {
	gcode := db.NormalizeGcodeState(strPayload(payload, "gcode_state"))
	to := db.JobRunning
	if gcode == "PAUSE" || gcode == "PAUSED" {
		to = db.JobPaused
	}
	_, err := db.TransitionJob(ctx, pool, siteID, jobID, to, "edge", "progress")
	return ignoreInvalid(err)
}

func applyFinished(ctx context.Context, pool *pgxpool.Pool, bus *events.Bus, siteID, jobID string, payload map[string]any) error {
	job, err := db.GetJob(ctx, pool, siteID, jobID)
	if err != nil {
		return err
	}
	job, err = db.TransitionJob(ctx, pool, siteID, jobID, db.JobSucceeded, "edge", "finished")
	if err != nil {
		return ignoreInvalid(err)
	}

	seconds := intPayload(payload, "duration_s")
	if seconds <= 0 {
		seconds = job.EstSeconds
	}
	grams := floatPayload(payload, "grams")
	if grams <= 0 {
		grams = job.EstGrams
	}
	if rc, err := db.GetRateCard(ctx, pool, siteID); err == nil {
		actual := db.EstimateJobCost(rc, seconds, grams, job.Copies)
		_ = db.ApplyJobCost(ctx, pool, siteID, jobID, job.CostEstimate, actual, rc.Currency)
	}
	spoolID := strPayload(payload, "spool_id")
	if spoolID == "" {
		spoolID = job.SpoolID
	}
	if err := PublishFilamentUsed(ctx, bus, siteID, jobID, spoolID, grams); err != nil {
		return err
	}
	if err := notify.PostFinished(ctx, os.Getenv("NOTIFY_URL"), nil, notify.Finished{
		JobID: jobID, State: db.JobSucceeded, Grams: grams, SpoolID: spoolID,
	}); err != nil {
		slog.Warn("notify webhook", "err", err, "job_id", jobID)
	}
	return nil
}

// PublishFilamentUsed emits filament.consumed when a spool and grams are known.
// A nil bus, empty spool, or non-positive grams is a no-op.
func PublishFilamentUsed(ctx context.Context, bus *events.Bus, siteID, jobID, spoolID string, grams float64) error {
	if bus == nil || grams <= 0 || spoolID == "" {
		return nil
	}
	env := events.NewEnvelope(events.TypeFilamentConsumed, "core", siteID, map[string]any{
		"job_id": jobID, "spool_id": spoolID, "grams": grams,
	})
	env.CorrelationID = jobID
	_, err := bus.Publish(ctx, events.StreamFilament, env)
	return err
}

func ignoreInvalid(err error) error {
	if err == nil {
		return nil
	}
	// Illegal / late transitions are ack'd; other errors leave the message pending.
	if ae, ok := api.AsError(err); ok && ae.Code == api.CodeInvalidTransition {
		return nil
	}
	return err
}

func strPayload(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	s, _ := m[key].(string)
	return s
}

func intPayload(m map[string]any, key string) int {
	if m == nil {
		return 0
	}
	switch v := m[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	default:
		return 0
	}
}

func floatPayload(m map[string]any, key string) float64 {
	if m == nil {
		return 0
	}
	switch v := m[key].(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	default:
		return 0
	}
}
