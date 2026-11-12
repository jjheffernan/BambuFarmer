package jobs

import (
	"context"

	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/events"
)

// StartPrintPayload is the bf.commands.edge.{site} body for command.start_print.
// Edge today reads job_id, external_sn|sn, file_id; artifact_id is our object-store key.
type StartPrintPayload struct {
	JobID      string `json:"job_id"`
	ExternalSN string `json:"external_sn"`
	SN         string `json:"sn"` // alias for older edge consumers
	ArtifactID string `json:"artifact_id,omitempty"`
	FileID     string `json:"file_id,omitempty"` // FM file id when known; MVP = artifact_id
}

// PublishStartPrint publishes command.start_print and job.assigned for an assigned job.
func PublishStartPrint(ctx context.Context, bus *events.Bus, siteID string, job api.Job) error {
	if bus == nil {
		return nil
	}
	payload := map[string]any{
		"job_id":      job.ID,
		"external_sn": job.ExternalSN,
		"sn":          job.ExternalSN,
		"artifact_id": job.ArtifactID,
		// file_id/f3mf_id: FM plate id when known; MVP uses artifact_id until upload maps it.
		"file_id":     job.ArtifactID,
		"f3mf_id":     job.ArtifactID,
		"task_name":   job.Name,
		"name":        job.Name,
		"copies":      job.Copies,
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
