package alerts

import (
	"context"
	"log/slog"
	"os"

	"github.com/jjheffernan/BambuFarmer/packages/events"
	"github.com/jjheffernan/BambuFarmer/services/core/internal/notify"
)

// ConsumeAlertEvents POSTs alert.raised payloads to NOTIFY_URL when set.
func ConsumeAlertEvents(ctx context.Context, bus *events.Bus, siteID, consumer string) {
	if bus == nil {
		return
	}
	endpoint := os.Getenv("NOTIFY_URL")
	if consumer == "" {
		consumer = "core-1"
	}
	err := bus.ConsumeLoop(ctx, events.StreamAlert, events.GroupCore, consumer, "$", func(ctx context.Context, _ string, env events.Envelope) error {
		if env.Type != events.TypeAlertRaised {
			return nil
		}
		sn, _ := env.Payload["external_sn"].(string)
		if sn == "" {
			return nil
		}
		name, _ := env.Payload["name"].(string)
		model, _ := env.Payload["model"].(string)
		state, _ := env.Payload["gcode_state"].(string)
		online, _ := env.Payload["online"].(bool)
		sid := env.SiteID
		if sid == "" {
			sid = siteID
		}
		body := notify.Alert{
			Type:       events.TypeAlertRaised,
			ExternalSN: sn,
			Name:       name,
			Model:      model,
			Online:     online,
			GcodeState: state,
			Reasons:    stringSlice(env.Payload["reasons"]),
			HMS:        stringSlice(env.Payload["hms"]),
			SiteID:     sid,
		}
		if err := notify.PostAlert(ctx, endpoint, nil, body); err != nil {
			slog.Warn("alert notify", "err", err, "sn", sn)
		}
		return nil // always ACK — notify failure should not block the stream
	})
	if err != nil && ctx.Err() == nil {
		slog.Warn("alert consumer stopped", "err", err)
	}
}

func stringSlice(v any) []string {
	switch x := v.(type) {
	case []string:
		return x
	case []any:
		out := make([]string, 0, len(x))
		for _, e := range x {
			if s, ok := e.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}
