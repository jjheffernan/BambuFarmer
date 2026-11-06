package printers

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jjheffernan/BambuFarmer/packages/db"
	"github.com/jjheffernan/BambuFarmer/packages/events"
)

// ConsumePrinterEvents projects bf.events.printer into the printers table.
func ConsumePrinterEvents(ctx context.Context, bus *events.Bus, pool *pgxpool.Pool, siteID, consumer string) {
	if pool == nil || bus == nil {
		return
	}
	if consumer == "" {
		consumer = "core-1"
	}
	err := bus.ConsumeLoop(ctx, events.StreamPrinter, events.GroupCore, consumer, "$", func(ctx context.Context, _ string, env events.Envelope) error {
		sn, _ := env.Payload["external_sn"].(string)
		if sn == "" {
			sn, _ = env.Payload["dev_id"].(string)
		}
		if sn == "" {
			return nil
		}
		name, _ := env.Payload["name"].(string)
		model, _ := env.Payload["model"].(string)
		state, _ := env.Payload["gcode_state"].(string)
		online, _ := env.Payload["online"].(bool)
		pct := IntFrom(env.Payload["mc_percent"])
		remain := IntFrom(env.Payload["mc_remaining_time"])
		raw, _ := json.Marshal(env.Payload)
		sid := env.SiteID
		if sid == "" {
			sid = siteID
		}
		return db.UpsertPrinterProjection(ctx, pool, sid, sn, name, model, state, online, pct, remain, raw)
	})
	if err != nil && ctx.Err() == nil {
		slog.Warn("printer consumer stopped", "err", err)
	}
}
