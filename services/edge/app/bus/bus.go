package bus

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/events"
	"github.com/jjheffernan/BambuFarmer/packages/farm"
)

// Commander is the subset of farm.ControlPlane the command consumer needs.
type Commander interface {
	Control(ctx context.Context, sn string, opt api.ControlOpt) error
	StartPrint(ctx context.Context, req farm.PrintRequest) (farm.ExternalJobRef, error)
}

// Uploader is optionally implemented by Commander for artifact → FM Upload3MF.
type Uploader interface {
	Upload3MF(ctx context.Context, name string, r io.Reader) (farm.FileRef, error)
}

// Publisher wraps events.Bus for edge printer deltas + command consume + job events.
type Publisher struct {
	Bus          *events.Bus
	SiteID       string
	ArtifactsDir string // shared volume with core; object_key is relative
	Log          *slog.Logger
}

// PublishPrinterDelta emits printer.state_changed (and offline heuristic).
func (p *Publisher) PublishPrinterDelta(ctx context.Context, snap api.PrinterSnapshot) {
	if p == nil || p.Bus == nil {
		return
	}
	payload := map[string]any{
		"external_sn":       snap.ExternalSN,
		"name":              snap.Name,
		"model":             snap.Model,
		"online":            snap.Online,
		"gcode_state":       snap.GcodeState,
		"mc_percent":        snap.MCPercent,
		"mc_remaining_time": snap.RemainingS,
		"bucket":            snap.Bucket,
		"state_forced":      snap.StateForced,
		"nozzle_temper":     snap.NozzleTemper,
		"bed_temper":        snap.BedTemper,
		"print_error":       snap.PrintError,
		"hms":               snap.HMS,
		"ams":               snap.AMS,
		"tags":              snap.Tags,
	}
	if snap.Task != nil {
		payload["task"] = snap.Task
	}
	typ := events.TypePrinterStateChanged
	if !snap.Online {
		typ = events.TypePrinterOffline
	}
	env := events.NewEnvelope(typ, "edge", p.SiteID, payload)
	id, err := p.Bus.Publish(ctx, events.StreamPrinter, env)
	if err != nil {
		p.Log.Warn("publish printer delta", "err", err)
		return
	}
	p.Log.Debug("published", "stream", events.StreamPrinter, "id", id, "sn", snap.ExternalSN, "state", snap.GcodeState)
}

// PublishAlertRaised emits alert.raised for operator attention (PAUSE/FAILED/offline/HMS).
func (p *Publisher) PublishAlertRaised(ctx context.Context, snap api.PrinterSnapshot) {
	if p == nil || p.Bus == nil {
		return
	}
	reasons := make([]string, 0, 3)
	if !snap.Online {
		reasons = append(reasons, "offline")
	}
	switch snap.GcodeState {
	case "PAUSE", "FAILED":
		reasons = append(reasons, snap.GcodeState)
	}
	if len(snap.HMS) > 0 {
		reasons = append(reasons, "hms")
	}
	payload := map[string]any{
		"external_sn": snap.ExternalSN,
		"name":        snap.Name,
		"model":       snap.Model,
		"online":      snap.Online,
		"gcode_state": snap.GcodeState,
		"hms":         snap.HMS,
		"reasons":     reasons,
		"bucket":      snap.Bucket,
	}
	env := events.NewEnvelope(events.TypeAlertRaised, "edge", p.SiteID, payload)
	id, err := p.Bus.Publish(ctx, events.StreamAlert, env)
	if err != nil {
		p.Log.Warn("publish alert.raised", "err", err)
		return
	}
	p.Log.Debug("published", "stream", events.StreamAlert, "id", id, "sn", snap.ExternalSN, "type", events.TypeAlertRaised)
}

// PublishAMSChanged emits printer.ams_changed with the current AMS tray map.
func (p *Publisher) PublishAMSChanged(ctx context.Context, snap api.PrinterSnapshot) {
	if p == nil || p.Bus == nil {
		return
	}
	payload := map[string]any{
		"external_sn": snap.ExternalSN,
		"name":        snap.Name,
		"model":       snap.Model,
		"ams":         snap.AMS,
	}
	env := events.NewEnvelope(events.TypePrinterAMSChanged, "edge", p.SiteID, payload)
	id, err := p.Bus.Publish(ctx, events.StreamPrinter, env)
	if err != nil {
		p.Log.Warn("publish ams_changed", "err", err)
		return
	}
	p.Log.Debug("published", "stream", events.StreamPrinter, "id", id, "sn", snap.ExternalSN, "type", events.TypePrinterAMSChanged)
}

// PublishJob emits a job lifecycle event on bf.events.job.
func (p *Publisher) PublishJob(ctx context.Context, typ, correlationID string, payload map[string]any) {
	if p == nil || p.Bus == nil {
		return
	}
	if payload == nil {
		payload = map[string]any{}
	}
	env := events.NewEnvelope(typ, "edge", p.SiteID, payload)
	env.CorrelationID = correlationID
	id, err := p.Bus.Publish(ctx, events.StreamJob, env)
	if err != nil {
		p.Log.Warn("publish job event", "type", typ, "err", err)
		return
	}
	p.Log.Debug("published", "stream", events.StreamJob, "id", id, "type", typ)
}

// ConsumeCommands listens on bf.commands.edge.{site} and dispatches to Commander.
func (p *Publisher) ConsumeCommands(ctx context.Context, cmd Commander) {
	if p == nil || p.Bus == nil || cmd == nil {
		return
	}
	stream := events.CommandStream(p.SiteID)
	consumer := "edge-1"
	err := p.Bus.ConsumeLoop(ctx, stream, "edge", consumer, "$", func(ctx context.Context, _ string, env events.Envelope) error {
		return p.HandleCommand(ctx, cmd, env)
	})
	if err != nil && ctx.Err() == nil {
		p.Log.Warn("command consumer stopped", "err", err)
	}
}

// HandleCommand dispatches one command envelope. Exported for unit tests.
func (p *Publisher) HandleCommand(ctx context.Context, cmd Commander, env events.Envelope) error {
	sn := payloadString(env.Payload, "external_sn", "sn")
	switch env.Type {
	case "command.control", "command.pause", "command.resume", "command.stop", "command.bed_clean":
		opt := api.ControlPause
		if v, ok := env.Payload["opt"].(string); ok && v != "" {
			opt = api.ControlOpt(v)
		} else if env.Type == "command.resume" {
			opt = api.ControlResume
		} else if env.Type == "command.stop" {
			opt = api.ControlStop
		} else if env.Type == "command.bed_clean" {
			opt = api.ControlBedClean
		}
		if sn == "" {
			return nil
		}
		return cmd.Control(ctx, sn, opt)
	case "command.start_print":
		return p.handleStartPrint(ctx, cmd, env, sn)
	default:
		if p.Log != nil {
			p.Log.Debug("ignored command", "type", env.Type)
		}
		return nil
	}
}

func (p *Publisher) handleStartPrint(ctx context.Context, cmd Commander, env events.Envelope, sn string) error {
	jobID := payloadString(env.Payload, "job_id")
	fileID := payloadString(env.Payload, "file_id", "f3mf_id")
	name := payloadString(env.Payload, "task_name", "name", "filename")
	objectKey := payloadString(env.Payload, "object_key")
	artifactID := payloadString(env.Payload, "artifact_id")
	plateIndex := payloadInt(env.Payload, "plate_index")
	copies := payloadInt(env.Payload, "copies")

	if fileID == "" {
		key := objectKey
		if key == "" {
			key = artifactID
		}
		uploaded, upName, err := p.uploadArtifact(ctx, cmd, key, name)
		if err != nil {
			if p.Log != nil {
				p.Log.Warn("upload before start_print", "err", err, "job_id", jobID, "key", key)
			}
			p.PublishJob(ctx, events.TypeJobFailed, jobID, map[string]any{
				"job_id": jobID, "external_sn": sn, "error": err.Error(), "artifact_id": artifactID,
			})
			return nil
		}
		fileID = uploaded
		if name == "" {
			name = upName
		}
	}

	req := farm.PrintRequest{
		JobID:      jobID,
		ExternalSN: sn,
		FileRef:    farm.FileRef{ID: fileID, Name: name},
		Name:       name,
		PlateIndex: plateIndex,
		Copies:     copies,
	}
	ref, err := cmd.StartPrint(ctx, req)
	if err != nil {
		if p.Log != nil {
			p.Log.Warn("start_print", "err", err, "sn", sn, "job_id", jobID)
		}
		p.PublishJob(ctx, events.TypeJobFailed, jobID, map[string]any{
			"job_id":      jobID,
			"external_sn": sn,
			"error":       err.Error(),
			"file_id":     fileID,
		})
		// ACK after job.failed — avoid infinite retry on permanent errors (missing f3mf_id, etc.).
		return nil
	}
	if p.Log != nil {
		p.Log.Info("start_print ok", "sn", sn, "job_id", jobID, "task_id", ref.TaskID, "file_id", fileID)
	}
	p.PublishJob(ctx, events.TypeJobDispatched, jobID, map[string]any{
		"job_id":           jobID,
		"external_sn":      sn,
		"external_task":    ref.TaskID,
		"external_subtask": ref.SubtaskID,
		"file_id":          fileID,
		"task_id":          ref.TaskID,
		"subtask_id":       ref.SubtaskID,
	})
	if binder, ok := cmd.(JobBinder); ok {
		binder.BindJob(sn, jobID, ref)
	}
	return nil
}

func (p *Publisher) uploadArtifact(ctx context.Context, cmd Commander, objectKey, name string) (fileID, filename string, err error) {
	if objectKey == "" {
		return "", "", fmt.Errorf("start_print: file_id or artifact object_key required")
	}
	up, ok := cmd.(Uploader)
	if !ok {
		return "", "", fmt.Errorf("start_print: commander cannot Upload3MF")
	}
	dir := ""
	if p != nil {
		dir = p.ArtifactsDir
	}
	if dir == "" {
		return "", "", fmt.Errorf("start_print: ARTIFACTS_DIR unset and no file_id")
	}
	path, err := absArtifact(dir, objectKey)
	if err != nil {
		return "", "", err
	}
	f, err := os.Open(path)
	if err != nil {
		return "", "", fmt.Errorf("open artifact %s: %w", objectKey, err)
	}
	defer f.Close()
	if name == "" {
		name = filepath.Base(path)
	}
	ref, err := up.Upload3MF(ctx, name, f)
	if err != nil {
		return "", "", err
	}
	if ref.ID == "" {
		return "", "", fmt.Errorf("upload returned empty file id")
	}
	return ref.ID, firstNonEmpty(ref.Name, name), nil
}

func absArtifact(dir, objectKey string) (string, error) {
	key := strings.TrimSpace(objectKey)
	if key == "" || strings.Contains(key, "..") || strings.HasPrefix(key, "/") {
		return "", fmt.Errorf("invalid object_key")
	}
	abs := filepath.Join(dir, filepath.FromSlash(key))
	rel, err := filepath.Rel(dir, abs)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("artifact path escape")
	}
	return abs, nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// JobBinder is optionally implemented by Commander to correlate later progress.
type JobBinder interface {
	BindJob(sn, jobID string, ref farm.ExternalJobRef)
}

func payloadString(payload map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := payload[k].(string); ok && v != "" {
			return v
		}
	}
	return ""
}

func payloadInt(payload map[string]any, key string) int {
	v, ok := payload[key]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	default:
		return 0
	}
}
