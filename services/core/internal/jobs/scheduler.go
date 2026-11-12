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
	Pool     *pgxpool.Pool
	Bus      *events.Bus
	SiteID   string
	Interval time.Duration
	Log      *slog.Logger
}

// Run ticks until ctx cancel.
func (s *Scheduler) Run(ctx context.Context) {
	if s == nil || s.Pool == nil {
		return
	}
	interval := s.Interval
	if interval <= 0 {
		interval = 3 * time.Second
	}
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

// Tick assigns as many queued jobs as idle printers allow.
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
	used := map[string]bool{}

	for _, job := range queued {
		prevState := ""
		if job.AfterJobID != "" {
			if prev, err := db.GetJob(ctx, s.Pool, s.SiteID, job.AfterJobID); err == nil {
				prevState = prev.State
			}
		}
		if WaitingOnPrevious(job.AfterJobID, prevState) {
			continue
		}
		sn := PickPrinter(job, idle, used)
		if sn == "" {
			continue
		}
		assigned, err := db.AssignJob(ctx, s.Pool, s.SiteID, job.ID, sn, "scheduler", "auto-assign")
		if err != nil {
			continue
		}
		used[sn] = true
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
