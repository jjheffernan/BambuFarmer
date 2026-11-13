package printers

import (
	"context"
	"strings"

	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/events"
)

// NormalizeSerials returns unique non-empty printer serials from a bulk request.
func NormalizeSerials(req api.PrinterActionsRequest) []string {
	raw := append([]string{}, req.Serials...)
	raw = append(raw, req.ExternalSNs...)
	seen := make(map[string]struct{}, len(raw))
	out := make([]string, 0, len(raw))
	for _, sn := range raw {
		sn = strings.TrimSpace(sn)
		if sn == "" {
			continue
		}
		if _, ok := seen[sn]; ok {
			continue
		}
		seen[sn] = struct{}{}
		out = append(out, sn)
	}
	return out
}

// PublishControl enqueues one command.control for each serial on the edge command stream.
func PublishControl(ctx context.Context, bus *events.Bus, siteID string, action api.ControlOpt, serials []string) (api.PrinterActionsResult, error) {
	res := api.PrinterActionsResult{Action: action, Accepted: []string{}, Skipped: []string{}}
	if !api.ValidControlOpt(action) {
		return res, api.BadRequest("unsupported action")
	}
	if len(serials) == 0 {
		return res, api.BadRequest("serials required")
	}
	if bus == nil {
		return res, api.NotConfigured("redis not configured")
	}
	stream := events.CommandStream(siteID)
	for _, sn := range serials {
		env := events.NewEnvelope(events.TypeCommandControl, "core", siteID, map[string]any{
			"external_sn": sn,
			"opt":         string(action),
		})
		if _, err := bus.Publish(ctx, stream, env); err != nil {
			res.Skipped = append(res.Skipped, sn)
			continue
		}
		res.Accepted = append(res.Accepted, sn)
	}
	return res, nil
}
