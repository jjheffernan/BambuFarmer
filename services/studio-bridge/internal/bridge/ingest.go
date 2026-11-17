package bridge

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

// Policy controls post-upload behavior.
type Policy string

const (
	PolicyArchive Policy = "archive" // artifact only
	PolicyQueue   Policy = "queue"   // artifact + create job
)

// ParsePolicy defaults to queue.
func ParsePolicy(s string) Policy {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "archive":
		return PolicyArchive
	default:
		return PolicyQueue
	}
}

// Request is one plate ingest.
type Request struct {
	Name        string
	Filename    string
	Body        []byte
	TargetModel string
	Queue       *bool // nil = use default policy
}

// Result is ingest outcome.
type Result struct {
	Received      bool   `json:"received"`
	ArtifactID    string `json:"artifact_id,omitempty"`
	JobID         string `json:"job_id,omitempty"`
	CreateJobPath string `json:"create_job_path,omitempty"`
	Reason        string `json:"reason,omitempty"`
}

// Ingest uploads to core and optionally queues a job.
func Ingest(ctx context.Context, coreURL string, defaultPolicy Policy, req Request, do HTTPDo) (Result, error) {
	if len(req.Body) == 0 {
		return Result{}, fmt.Errorf("file required")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		base := filepath.Base(req.Filename)
		name = strings.TrimSuffix(base, filepath.Ext(base))
		name = strings.TrimSuffix(name, ".gcode")
		if name == "" || name == "." {
			name = "studio-bridge"
		}
	}
	id, err := UploadArtifact(ctx, coreURL, req.Filename, req.Body, do)
	if err != nil {
		return Result{}, err
	}
	if id == "" {
		return Result{Received: true, Reason: "CORE_URL unset — plate accepted but not registered"}, nil
	}
	out := Result{
		Received:      true,
		ArtifactID:    id,
		CreateJobPath: "/jobs?artifact=" + id,
	}
	wantQueue := defaultPolicy == PolicyQueue
	if req.Queue != nil {
		wantQueue = *req.Queue
	}
	if !wantQueue {
		out.Reason = "archived"
		return out, nil
	}
	jobID, err := CreateJob(ctx, coreURL, name, id, req.TargetModel, do)
	if err != nil {
		out.Reason = "artifact registered; job failed: " + err.Error()
		return out, nil
	}
	out.JobID = jobID
	if jobID != "" {
		out.CreateJobPath = "/jobs?job=" + jobID
	}
	return out, nil
}
