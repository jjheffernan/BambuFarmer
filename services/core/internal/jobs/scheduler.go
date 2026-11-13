package jobs

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/db"
	"github.com/jjheffernan/BambuFarmer/packages/events"
)

// Scheduler periodically assigns queued jobs to idle printers and emits start_print.
type Scheduler struct {
	Pool             *pgxpool.Pool
	Bus              *events.Bus
	SiteID           string
	Interval         time.Duration
	Stagger          time.Duration // when set, at most one start per tick; tick period ≥ Stagger
	MaxStartsPerTick int           // 0 = unlimited (unless Stagger set → 1)
	CircuitMax       int           // 0 = no circuit tag limit
	Log              *slog.Logger
}

// startsCap is how many jobs may start in one Tick (0 = unlimited).
func (s *Scheduler) startsCap() int {
	if s == nil {
		return 0
	}
	if s.MaxStartsPerTick > 0 {
		return s.MaxStartsPerTick
	}
	if s.Stagger > 0 {
		return 1
	}
	return 0
}

// tickInterval is the poll period, never shorter than Stagger when set.
func (s *Scheduler) tickInterval() time.Duration {
	interval := time.Duration(0)
	if s != nil {
		interval = s.Interval
	}
	if interval <= 0 {
		interval = 3 * time.Second
	}
	if s != nil && s.Stagger > interval {
		return s.Stagger
	}
	return interval
}

// Run ticks until ctx cancel.
func (s *Scheduler) Run(ctx context.Context) {
	if s == nil || s.Pool == nil {
		return
	}
	interval := s.tickInterval()
	log := s.Log
	if log == nil {
		log = slog.Default()
	}
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		if err := s.Tick(ctx); err != nil && ctx.Err() == nil {
			log.Warn("scheduler tick", "err", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}

// Tick assigns queued jobs to idle printers (optionally capped for stagger).
func (s *Scheduler) Tick(ctx context.Context) error {
	queued, err := db.ListQueuedJobs(ctx, s.Pool, s.SiteID, 50)
	if err != nil {
		return err
	}
	if len(queued) == 0 {
		return nil
	}
	idle, err := db.ListIdlePrinters(ctx, s.Pool, s.SiteID)
	if err != nil {
		return err
	}
	known, err := db.ListPrinters(ctx, s.Pool, s.SiteID)
	if err != nil {
		return err
	}
	idleSet := map[string]bool{}
	for _, p := range idle {
		idleSet[p.ExternalSN] = true
	}
	used := map[string]bool{}
	for _, p := range known {
		if p.Online && !idleSet[p.ExternalSN] {
			used[p.ExternalSN] = true
		}
	}

	capN := s.startsCap()
	started := 0
	for _, job := range queued {
		if capN > 0 && started >= capN {
			break
		}
		prevState := ""
		if job.AfterJobID != "" {
			if prev, err := db.GetJob(ctx, s.Pool, s.SiteID, job.AfterJobID); err == nil {
				prevState = prev.State
			}
		}
		if WaitingOnPrevious(job.AfterJobID, prevState) {
			continue
		}
		sn := PickPrinter(job, idle, known, used, s.CircuitMax)
		if sn == "" {
			continue
		}
		assigned, err := db.AssignJob(ctx, s.Pool, s.SiteID, job.ID, sn, "scheduler", "auto-assign")
		if err != nil {
			continue
		}
		used[sn] = true
		started++
		if err := PublishStartPrintWithPool(ctx, s.Bus, s.Pool, s.SiteID, assigned); err != nil {
			slog.Warn("publish start_print", "err", err, "job_id", assigned.ID)
		}
	}
	return nil
}

// DispatchOne assigns one job (API path) and publishes start_print.
// preferredSN overrides targeting when non-empty.
func DispatchOne(ctx context.Context, pool *pgxpool.Pool, bus *events.Bus, siteID, jobID, preferredSN string) (api.Job, error) {
	job, err := db.GetJob(ctx, pool, siteID, jobID)
	if err != nil {
		return job, err
	}
	// Already past assign: do not re-emit start_print.
	switch job.State {
	case db.JobAssigned, db.JobDispatched, db.JobRunning, db.JobPaused:
		return job, nil
	case db.JobSucceeded, db.JobFailed, db.JobCancelled:
		return job, db.InvalidJobTransition(job.State, db.JobAssigned)
	}
	sn := preferredSN
	if sn == "" {
		sn = job.ExternalSN
	}
	if sn == "" {
		idle, err := db.ListIdlePrinters(ctx, pool, siteID)
		if err != nil {
			return job, err
		}
		if len(idle) == 0 {
			return job, db.Validation("no idle printer available")
		}
		sn = idle[0].ExternalSN
	}
	assigned, err := db.AssignJob(ctx, pool, siteID, jobID, sn, "api", "dispatch")
	if err != nil {
		return assigned, err
	}
	if err := PublishStartPrintWithPool(ctx, bus, pool, siteID, assigned); err != nil {
		return assigned, err
	}
	return assigned, nil
}
