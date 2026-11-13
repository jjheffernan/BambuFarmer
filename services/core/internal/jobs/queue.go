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

// CircuitOK is true when circuitMax is unset, or every tag on p still has room.
// used is busy SNs plus claims from this tick. known supplies tags for those SNs.
func CircuitOK(p api.PrinterSummary, known []api.PrinterSummary, used map[string]bool, circuitMax int) bool {
	if circuitMax <= 0 || len(p.Tags) == 0 {
		return true
	}
	bySN := map[string]api.PrinterSummary{}
	for _, x := range known {
		bySN[x.ExternalSN] = x
	}
	bySN[p.ExternalSN] = p
	for _, tag := range p.Tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		n := 0
		for sn := range used {
			peer, ok := bySN[sn]
			if !ok {
				continue
			}
			for _, t := range peer.Tags {
				if strings.EqualFold(strings.TrimSpace(t), tag) {
					n++
					break
				}
			}
		}
		if n >= circuitMax {
			return false
		}
	}
	return true
}

// PickPrinter chooses an idle SN for a queued job.
// A set external_sn stays on that printer. target_model filters candidates.
// used is busy SNs plus claims from this tick. known supplies tags for circuit counts.
// circuitMax 0 means no circuit limit.
func PickPrinter(job api.Job, idle, known []api.PrinterSummary, used map[string]bool, circuitMax int) string {
	if known == nil {
		known = idle
	}
	if job.ExternalSN != "" {
		for _, p := range idle {
			if p.ExternalSN == job.ExternalSN && !used[p.ExternalSN] && ModelMatch(job.TargetModel, p.Model) && CircuitOK(p, known, used, circuitMax) {
				return p.ExternalSN
			}
		}
		return ""
	}
	for _, p := range idle {
		if used[p.ExternalSN] || !ModelMatch(job.TargetModel, p.Model) || !CircuitOK(p, known, used, circuitMax) {
			continue
		}
		return p.ExternalSN
	}
	return ""
}
