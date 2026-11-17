package bridge

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// WatchOptions configures the drop-folder poller.
type WatchOptions struct {
	Dir      string
	CoreURL  string
	Policy   Policy
	Interval time.Duration
	Log      *slog.Logger
	Do       HTTPDo
}

// WatchDropDir polls for .3mf plates, ingests them, then moves to processed/.
// Empty Dir is a no-op. Cancels on ctx done.
func WatchDropDir(ctx context.Context, opt WatchOptions) {
	dir := strings.TrimSpace(opt.Dir)
	if dir == "" {
		return
	}
	log := opt.Log
	if log == nil {
		log = slog.Default()
	}
	interval := opt.Interval
	if interval <= 0 {
		interval = 5 * time.Second
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Warn("studio-bridge drop dir", "err", err, "dir", dir)
		return
	}
	processed := filepath.Join(dir, "processed")
	_ = os.MkdirAll(processed, 0o755)
	t := time.NewTicker(interval)
	defer t.Stop()
	scan := func() {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			lower := strings.ToLower(name)
			if !strings.HasSuffix(lower, ".3mf") {
				continue
			}
			path := filepath.Join(dir, name)
			body, err := os.ReadFile(path)
			if err != nil {
				log.Warn("read drop plate", "err", err, "path", path)
				continue
			}
			res, err := Ingest(ctx, opt.CoreURL, opt.Policy, Request{
				Filename: name,
				Body:     body,
			}, opt.Do)
			if err != nil {
				log.Warn("ingest drop plate", "err", err, "path", path)
				continue
			}
			dest := filepath.Join(processed, name)
			if err := os.Rename(path, dest); err != nil {
				log.Warn("move processed plate", "err", err, "from", path)
			}
			log.Info("studio-bridge drop ingested",
				"file", name, "artifact_id", res.ArtifactID, "job_id", res.JobID, "reason", res.Reason)
		}
	}
	scan()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			scan()
		}
	}
}
