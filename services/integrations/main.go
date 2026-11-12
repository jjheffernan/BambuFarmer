package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/jjheffernan/BambuFarmer/packages/events"
	"github.com/jjheffernan/BambuFarmer/packages/httpserver"
	"github.com/jjheffernan/BambuFarmer/services/integrations/app/config"
	"github.com/jjheffernan/BambuFarmer/services/integrations/app/httpapi"
	"github.com/jjheffernan/BambuFarmer/services/integrations/app/manyfold"
	"github.com/jjheffernan/BambuFarmer/services/integrations/app/spoolman"
)

func main() {
	cfg := config.Load()
	deps := httpapi.Deps{
		SpoolURL: cfg.SpoolmanURL,
		ManyURL:  cfg.ManyfoldURL,
	}
	if cfg.SpoolmanURL != "" {
		deps.Spoolman = spoolman.New(cfg.SpoolmanURL, cfg.SpoolmanToken, cfg.HTTPTimeout)
		slog.Info("spoolman configured", "url", cfg.SpoolmanURL)
	} else {
		slog.Info("spoolman disabled (SPOOLMAN_URL empty)")
	}
	if cfg.ManyfoldURL != "" {
		deps.Manyfold = manyfold.New(cfg.ManyfoldURL, cfg.ManyfoldToken, cfg.ManyfoldClientID, cfg.ManyfoldClientSecret, cfg.HTTPTimeout)
		slog.Info("manyfold configured", "url", cfg.ManyfoldURL)
	} else {
		slog.Info("manyfold disabled (MANYFOLD_URL empty)")
	}

	if cfg.RedisURL != "" {
		go consumeFilament(cfg.RedisURL, deps.Spoolman)
	}

	mux := httpserver.NewMux("integrations")
	httpapi.Register(mux, deps)

	if err := httpserver.ListenAndServe(httpserver.Options{
		Service: "integrations",
		Addr:    cfg.HTTPAddr,
		Mux:     mux,
	}); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

func consumeFilament(redisURL string, client *spoolman.Client) {
	ctx := context.Background()
	bus, err := events.Dial(ctx, redisURL)
	if err != nil {
		slog.Warn("filament consumer skipped", "err", err)
		return
	}
	defer bus.Close()
	err = bus.ConsumeLoop(ctx, events.StreamFilament, events.GroupIntegrations, "integrations-filament-1", "$", func(ctx context.Context, _ string, env events.Envelope) error {
		return spoolman.Consume(ctx, client, env)
	})
	if err != nil && ctx.Err() == nil {
		slog.Warn("filament consumer stopped", "err", err)
	}
}
