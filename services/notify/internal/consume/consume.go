package consume

import (
	"context"
	"log/slog"

	"github.com/jjheffernan/BambuFarmer/packages/events"
	"github.com/jjheffernan/BambuFarmer/services/notify/internal/export"
	"github.com/jjheffernan/BambuFarmer/services/notify/internal/store"
)

// RelevantTypes are fan-out by default when a channel lists them.
var RelevantTypes = map[string]bool{
	events.TypeJobFinished: true,
	events.TypeJobFailed:   true,
	events.TypeAlertRaised: true,
}

// Loop consumes job + alert streams and fans out to webhook channels.
func Loop(ctx context.Context, bus *events.Bus, st *store.Store, client *export.Client, consumer string) {
	if bus == nil || st == nil {
		return
	}
	if consumer == "" {
		consumer = "notify-1"
	}
	if client == nil {
		client = &export.Client{}
	}
	for _, stream := range []string{events.StreamJob, events.StreamAlert} {
		s := stream
		go func() {
			err := bus.ConsumeLoop(ctx, s, events.GroupNotify, consumer, "$", func(ctx context.Context, _ string, env events.Envelope) error {
				if !RelevantTypes[env.Type] {
					return nil
				}
				_ = client.FanOut(ctx, st, env)
				return nil // always ACK — delivery failures are logged, not retried in MVP
			})
			if err != nil && ctx.Err() == nil {
				slog.Warn("notify consumer stopped", "stream", s, "err", err)
			}
		}()
	}
	<-ctx.Done()
}
