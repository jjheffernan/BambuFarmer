package bus

import (
	"context"
	"log/slog"

	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/events"
	"github.com/jjheffernan/BambuFarmer/packages/farm"
)

// Commander is the subset of farm.ControlPlane the command consumer needs.
type Commander interface {
	Control(ctx context.Context, sn string, opt api.ControlOpt) error
	StartPrint(ctx context.Context, req farm.PrintRequest) (farm.ExternalJobRef, error)
}

// Publisher wraps events.Bus for edge printer deltas + command consume.
type Publisher struct {
	Bus    *events.Bus
	SiteID string
	Log    *slog.Logger
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

// ConsumeCommands listens on bf.commands.edge.{site} and dispatches to Commander.
func (p *Publisher) ConsumeCommands(ctx context.Context, cmd Commander) {
	if p == nil || p.Bus == nil || cmd == nil {
		return
	}
	stream := events.CommandStream(p.SiteID)
	consumer := "edge-1"
	err := p.Bus.ConsumeLoop(ctx, stream, "edge", consumer, "$", func(ctx context.Context, _ string, env events.Envelope) error {
		sn, _ := env.Payload["external_sn"].(string)
		if sn == "" {
			sn, _ = env.Payload["sn"].(string)
		}
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
			jobID, _ := env.Payload["job_id"].(string)
			fileID, _ := env.Payload["file_id"].(string)
			ref, err := cmd.StartPrint(ctx, farm.PrintRequest{
				JobID:      jobID,
				ExternalSN: sn,
				FileRef:    farm.FileRef{ID: fileID},
			})
			if err != nil {
				p.Log.Warn("start_print", "err", err, "sn", sn, "job_id", jobID)
				return err
			}
			p.Log.Info("start_print ok", "sn", sn, "job_id", jobID, "task_id", ref.TaskID)
			return nil
		default:
			p.Log.Debug("ignored command", "type", env.Type)
			return nil
		}
	})
	if err != nil && ctx.Err() == nil {
		p.Log.Warn("command consumer stopped", "err", err)
	}
}
