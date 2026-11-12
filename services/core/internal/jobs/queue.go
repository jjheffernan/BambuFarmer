package jobs

import (
	"strings"

	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/db"
)

// WaitingOnPrevious is true until after_job_id's job has succeeded.
// A missing previous job (empty state) also waits.
func WaitingOnPrevious(afterJobID, prevState string) bool {
	if strings.TrimSpace(afterJobID) == "" {
		return false
	}
	return prevState != db.JobSucceeded
}

// ModelMatch is true when the job has no target, or the printer model equals it.
func ModelMatch(target, printerModel string) bool {
	target = strings.TrimSpace(target)
	if target == "" {
		return true
	}
	return strings.EqualFold(target, strings.TrimSpace(printerModel))
}

// PickPrinter chooses an idle SN for a queued job.
// A set external_sn stays on that printer. target_model filters candidates.
// used SNs are already taken in this tick.
func PickPrinter(job api.Job, idle []api.PrinterSummary, used map[string]bool) string {
	if job.ExternalSN != "" {
		for _, p := range idle {
			if p.ExternalSN == job.ExternalSN && !used[p.ExternalSN] && ModelMatch(job.TargetModel, p.Model) {
				return p.ExternalSN
			}
		}
		return ""
	}
	for _, p := range idle {
		if used[p.ExternalSN] || !ModelMatch(job.TargetModel, p.Model) {
			continue
		}
		return p.ExternalSN
	}
	return ""
}
