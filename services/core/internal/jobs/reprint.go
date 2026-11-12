package jobs

import (
	"strings"

	"github.com/jjheffernan/BambuFarmer/packages/api"
)

// ReprintRequest copies a retained plate into a new queued job.
// Same artifact or FM file id, copies, estimates, and printer target.
// Cost is re-estimated on insert; no schema change.
func ReprintRequest(src api.Job) (api.CreateJobRequest, error) {
	if strings.TrimSpace(src.ArtifactID) == "" && strings.TrimSpace(src.ExternalFileID) == "" {
		return api.CreateJobRequest{}, api.BadRequest("job has no artifact or file to reprint")
	}
	name := strings.TrimSpace(src.Name)
	if name == "" {
		name = "reprint"
	}
	const suffix = " (reprint)"
	if !strings.HasSuffix(name, suffix) {
		name += suffix
	}
	copies := src.Copies
	if copies < 1 {
		copies = 1
	}
	return api.CreateJobRequest{
		Name:        name,
		ExternalSN:  src.ExternalSN,
		ArtifactID:  src.ArtifactID,
		FileID:      src.ExternalFileID,
		Copies:      copies,
		EstSeconds:  src.EstSeconds,
		EstGrams:    src.EstGrams,
		TargetModel: src.TargetModel,
	}, nil
}
