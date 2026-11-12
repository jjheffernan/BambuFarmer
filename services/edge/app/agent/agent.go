// Package agent is the edge process runtime: poll a farm.ControlPlane, track state, publish events, run commands.
package agent

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/events"
	"github.com/jjheffernan/BambuFarmer/packages/farm"
	"github.com/jjheffernan/BambuFarmer/packages/printerstate"
	"github.com/jjheffernan/BambuFarmer/services/edge/app/bus"
	"github.com/jjheffernan/BambuFarmer/services/edge/app/config"
	"github.com/jjheffernan/BambuFarmer/services/edge/app/delta"
	"github.com/jjheffernan/BambuFarmer/services/edge/app/tracker"
)

// activeJob correlates an in-flight BambuFarmer job with a printer SN.
type activeJob struct {
	JobID      string
	TaskID     string
	SubtaskID  string
	LastState  string
	LastPct    int
}

// Agent owns the farm control plane and Redis workers for one site.
type Agent struct {
	cfg   config.Config
	plane farm.ControlPlane
	track *tracker.Tracker
	delta *delta.Filter
	pub   *bus.Publisher
	log   *slog.Logger

	mu       sync.RWMutex
	captain  api.CaptainInfo
	lastPoll time.Time
	lastErr  string
	started  bool
	jobs     map[string]activeJob // external_sn → job
}

// Options configures New.
type Options struct {
	Config config.Config
	Plane  farm.ControlPlane
	Bus    *events.Bus
	Log    *slog.Logger
}

// New builds an agent around an already-selected ControlPlane.
func New(opts Options) *Agent {
	log := opts.Log
	if log == nil {
		log = slog.Default()
	}
	a := &Agent{
		cfg:   opts.Config,
		plane: opts.Plane,
		track: tracker.New(opts.Config.SiteID),
		delta: delta.New(),
		log:   log,
		jobs:  make(map[string]activeJob),
	}
	if opts.Bus != nil {
		a.pub = &bus.Publisher{
			Bus:          opts.Bus,
			SiteID:       opts.Config.SiteID,
			ArtifactsDir: opts.Config.ArtifactsDir,
			Log:          log,
		}
	}
	return a
}

// Plane returns the underlying Bambu integration.
func (a *Agent) Plane() farm.ControlPlane { return a.plane }

// Connect establishes the farm session.
func (a *Agent) Connect(ctx context.Context) error {
	if a.plane == nil {
		return farm.ErrNotConfigured
	}
	if err := a.plane.Connect(ctx); err != nil {
		a.setErr(err)
		return err
	}
	if cap, err := a.plane.Captain(ctx); err == nil {
		a.mu.Lock()
		a.captain = cap
		a.lastErr = ""
		a.mu.Unlock()
		a.log.Info("farm connected", "kind", a.plane.Kind(), "server_id", cap.ServerID, "release", cap.Release)
	}
	return nil
}

// Start begins poll + command consumer until ctx cancel.
func (a *Agent) Start(ctx context.Context) {
	if a.pub != nil {
		go a.pub.ConsumeCommands(ctx, a)
	}
	if a.plane == nil || a.plane.Kind() == farm.KindNone {
		return
	}
	a.mu.Lock()
	if a.started {
		a.mu.Unlock()
		return
	}
	a.started = true
	a.mu.Unlock()

	go func() {
		t := time.NewTicker(a.cfg.PollInterval)
		defer t.Stop()
		a.pollOnce(ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				a.pollOnce(ctx)
			}
		}
	}()
}

func (a *Agent) pollOnce(ctx context.Context) {
	devices, err := a.plane.ListDevices(ctx)
	if err != nil {
		a.setErr(err)
		a.log.Warn("list devices", "kind", a.plane.Kind(), "err", err)
		_ = a.plane.Connect(ctx) // best-effort reconnect
		return
	}
	snaps, err := a.track.ObserveAll(devices)
	if err != nil {
		a.setErr(err)
		return
	}
	changed := 0
	for _, s := range snaps {
		if a.delta.Changed(s) {
			changed++
			if a.pub != nil {
				a.pub.PublishPrinterDelta(ctx, s)
			}
		}
		a.emitJobLifecycle(ctx, s)
	}
	a.mu.Lock()
	a.lastPoll = time.Now().UTC()
	a.lastErr = ""
	a.mu.Unlock()
	if changed > 0 {
		a.log.Info("poll", "kind", a.plane.Kind(), "devices", len(snaps), "changed", changed)
	}
}

// Control implements bus.Commander.
func (a *Agent) Control(ctx context.Context, sn string, opt api.ControlOpt) error {
	if err := a.plane.Control(ctx, sn, opt); err != nil {
		return err
	}
	// Immediate SSE update (do not wait for the next poll tick).
	a.publishSN(ctx, sn)
	return nil
}

// StartPrint implements bus.Commander.
func (a *Agent) StartPrint(ctx context.Context, req farm.PrintRequest) (farm.ExternalJobRef, error) {
	ref, err := a.plane.StartPrint(ctx, req)
	if err != nil {
		return ref, err
	}
	// Immediate SSE update so fleet/jobs UIs see RUNNING without waiting for poll.
	a.publishSN(ctx, req.ExternalSN)
	return ref, nil
}

// publishSN lists devices and emits a printer delta (+ job lifecycle) for one SN.
func (a *Agent) publishSN(ctx context.Context, sn string) {
	if sn == "" || a.plane == nil {
		return
	}
	devices, err := a.plane.ListDevices(ctx)
	if err != nil {
		a.log.Debug("publishSN list", "sn", sn, "err", err)
		return
	}
	snaps, err := a.track.ObserveAll(devices)
	if err != nil {
		return
	}
	for _, s := range snaps {
		if s.ExternalSN != sn {
			continue
		}
		if a.pub != nil && a.delta.Changed(s) {
			a.pub.PublishPrinterDelta(ctx, s)
		}
		a.emitJobLifecycle(ctx, s)
		return
	}
}

// Upload3MF implements bus.Uploader for artifact → FM staging before StartPrint.
func (a *Agent) Upload3MF(ctx context.Context, name string, r io.Reader) (farm.FileRef, error) {
	return a.plane.Upload3MF(ctx, name, r)
}

// BindJob implements bus.JobBinder after a successful start_print command.
func (a *Agent) BindJob(sn, jobID string, ref farm.ExternalJobRef) {
	if sn == "" {
		return
	}
	a.mu.Lock()
	a.jobs[sn] = activeJob{
		JobID:     jobID,
		TaskID:    ref.TaskID,
		SubtaskID: ref.SubtaskID,
	}
	a.mu.Unlock()
	a.track.SetPending(sn, true)
}

// emitJobLifecycle publishes job.progress / finished / failed when a bound SN changes.
func (a *Agent) emitJobLifecycle(ctx context.Context, snap api.PrinterSnapshot) {
	a.mu.Lock()
	job, ok := a.jobs[snap.ExternalSN]
	if !ok {
		a.mu.Unlock()
		return
	}
	stateChanged := snap.GcodeState != job.LastState
	pctChanged := snap.MCPercent != job.LastPct
	if !stateChanged && !pctChanged {
		a.mu.Unlock()
		return
	}
	job.LastState = snap.GcodeState
	job.LastPct = snap.MCPercent
	a.jobs[snap.ExternalSN] = job
	jobCopy := job
	a.mu.Unlock()

	payload := map[string]any{
		"job_id":           jobCopy.JobID,
		"external_sn":      snap.ExternalSN,
		"external_task":    jobCopy.TaskID,
		"task_id":          jobCopy.TaskID,
		"subtask_id":       jobCopy.SubtaskID,
		"gcode_state":      snap.GcodeState,
		"mc_percent":       snap.MCPercent,
		"mc_remaining_time": snap.RemainingS,
		"layer_num":        snap.LayerNum,
		"total_layer_num":  snap.TotalLayerNum,
	}

	switch printerstate.GcodeState(snap.GcodeState) {
	case printerstate.Finish:
		if a.pub != nil {
			a.pub.PublishJob(ctx, events.TypeJobFinished, jobCopy.JobID, payload)
		}
		a.clearJob(snap.ExternalSN)
	case printerstate.Failed:
		payload["error"] = "printer reported FAILED"
		if a.pub != nil {
			a.pub.PublishJob(ctx, events.TypeJobFailed, jobCopy.JobID, payload)
		}
		a.clearJob(snap.ExternalSN)
	default:
		if a.pub != nil {
			a.pub.PublishJob(ctx, events.TypeJobProgress, jobCopy.JobID, payload)
		}
	}
}

func (a *Agent) clearJob(sn string) {
	a.mu.Lock()
	delete(a.jobs, sn)
	a.mu.Unlock()
	a.track.SetPending(sn, false)
}

// Status builds GET /api/v1/edge/status.
func (a *Agent) Status() api.EdgeStatus {
	a.mu.RLock()
	defer a.mu.RUnlock()
	kind := farm.KindNone
	if a.plane != nil {
		kind = a.plane.Kind()
	}
	connected := false
	if a.plane != nil && a.plane.Ready(context.Background()) == nil {
		connected = true
	}
	st := api.EdgeStatus{
		SiteID:         a.cfg.SiteID,
		FarmMode:       "farm_manager",
		FarmManagerURL: a.cfg.FarmManagerURL,
		Transport:      string(kind),
		Connected:      connected,
		ServerID:       a.captain.ServerID,
		Release:        a.captain.Release,
		DeviceCount:    a.track.Count(),
		LastError:      a.lastErr,
		Devices:        a.track.Summaries(),
	}
	if !a.lastPoll.IsZero() {
		t := a.lastPoll
		st.LastPollAt = &t
	}
	switch kind {
	case farm.KindFixture:
		st.Note = "fixture control plane — " + a.cfg.FixtureDir
	case farm.KindNone:
		st.Note = "no farm backend — set FARM_MANAGER_URL or EDGE_FIXTURE_DIR"
	case farm.KindCDP:
		st.Note = "CDP backend stub — not connected"
	case farm.KindNative:
		if !connected {
			st.Note = "fm_native awaiting session"
		}
	}
	return st
}

func (a *Agent) setErr(err error) {
	a.mu.Lock()
	a.lastErr = err.Error()
	a.mu.Unlock()
}
