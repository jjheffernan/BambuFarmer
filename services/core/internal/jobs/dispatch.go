package jobs

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/db"
	"github.com/jjheffernan/BambuFarmer/packages/events"
)

// StartPrintPayload is the bf.commands.edge.{site} body for command.start_print.
// file_id/f3mf_id are FM plate ids only — never aliased from our artifact_id.
type StartPrintPayload struct {
	JobID      string `json:"job_id"`
	ExternalSN string `json:"external_sn"`
	SN         string `json:"sn"` // alias for older edge consumers
	ArtifactID string `json:"artifact_id,omitempty"`
	ObjectKey  string `json:"object_key,omitempty"` // relative path under ARTIFACTS_DIR
	Filename   string `json:"filename,omitempty"`
	FileID     string `json:"file_id,omitempty"`  // FM f3mf_id when already known
	F3MFID     string `json:"f3mf_id,omitempty"` // alias of file_id
}

// PublishStartPrint publishes command.start_print and job.assigned for an assigned job.
func PublishStartPrint(ctx context.Context, bus *events.Bus, siteID string, job api.Job) error {
	return PublishStartPrintWithPool(ctx, bus, nil, siteID, job)
}

// PublishStartPrintWithPool enriches object_key/filename from artifacts when pool is set.
func PublishStartPrintWithPool(ctx context.Context, bus *events.Bus, pool *pgxpool.Pool, siteID string, job api.Job) error {
	if bus == nil {
		return nil
	}
	payload := map[string]any{
		"job_id":      job.ID,
		"external_sn": job.ExternalSN,
		"sn":          job.ExternalSN,
		"artifact_id": job.ArtifactID,
		"task_name":   job.Name,
		"name":        job.Name,
		"copies":      job.Copies,
	}
	if job.ExternalFileID != "" {
		payload["file_id"] = job.ExternalFileID
		payload["f3mf_id"] = job.ExternalFileID
	}
	if pool != nil && job.ArtifactID != "" {
		if art, err := db.GetArtifact(ctx, pool, siteID, job.ArtifactID); err == nil {
			payload["object_key"] = art.ObjectKey
			payload["filename"] = art.Filename
			if job.Name == "" && art.Filename != "" {
				payload["task_name"] = art.Filename
				payload["name"] = art.Filename
			}
		}
	}
	cmd := events.NewEnvelope(events.TypeCommandStartPrint, "core", siteID, payload)
	cmd.CorrelationID = job.ID
	if _, err := bus.Publish(ctx, events.CommandStream(siteID), cmd); err != nil {
		return err
	}
	assigned := events.NewEnvelope(events.TypeJobAssigned, "core", siteID, map[string]any{
		"job_id": job.ID, "external_sn": job.ExternalSN, "artifact_id": job.ArtifactID,
	})
	assigned.CorrelationID = job.ID
	_, err := bus.Publish(ctx, events.StreamJob, assigned)
	return err
}

// PublishJobQueued emits job.queued after create.
func PublishJobQueued(ctx context.Context, bus *events.Bus, siteID string, job api.Job) {
	if bus == nil {
		return
	}
	env := events.NewEnvelope(events.TypeJobQueued, "core", siteID, map[string]any{
		"job_id": job.ID, "name": job.Name, "external_sn": job.ExternalSN, "artifact_id": job.ArtifactID,
	})
	env.CorrelationID = job.ID
	_, _ = bus.Publish(ctx, events.StreamJob, env)
}
