// Command notify is the optional notifications plugin (compose profile: notify).
// Consumes Redis Streams, fans out to CRUD webhook channels, and accepts ingest.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/jjheffernan/BambuFarmer/packages/db"
	"github.com/jjheffernan/BambuFarmer/packages/events"
	"github.com/jjheffernan/BambuFarmer/packages/httpserver"
	"github.com/jjheffernan/BambuFarmer/services/notify/internal/httpapi"
	"github.com/jjheffernan/BambuFarmer/services/notify/internal/consume"
	"github.com/jjheffernan/BambuFarmer/services/notify/internal/export"
	"github.com/jjheffernan/BambuFarmer/services/notify/internal/store"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	service := "notify"
	siteID := getenv("SITE_ID", "default")
	databaseURL := os.Getenv("DATABASE_URL")
	redisURL := os.Getenv("REDIS_URL")

	if databaseURL == "" {
		slog.Error("DATABASE_URL required")
		os.Exit(1)
	}
	pool, err := db.Open(ctx, databaseURL)
	if err != nil {
		slog.Error("database open failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()
	if err := db.Migrate(ctx, pool); err != nil {
		slog.Error("migrate failed", "err", err)
		os.Exit(1)
	}

	var bus *events.Bus
	if redisURL != "" {
		bus, err = events.Dial(ctx, redisURL)
		if err != nil {
			slog.Warn("redis unavailable; ingest publish + stream fan-out disabled", "err", err)
		} else {
			defer bus.Close()
		}
	} else {
		slog.Warn("REDIS_URL unset; ingest publish + stream fan-out disabled")
	}

	st := &store.Store{Pool: pool}
	ex := &export.Client{Log: slog.Default()}
	if bus != nil {
		go consume.Loop(ctx, bus, st, ex, "notify-"+getenv("HOSTNAME", "1"))
	}

	mux := httpserver.NewMux(service)
	h := &httpapi.Handler{Store: st, Bus: bus, Export: ex, SiteID: siteID}
	h.Register(mux)

	if err := httpserver.ListenAndServe(httpserver.Options{
		Service: service,
		Addr:    getenv("HTTP_ADDR", ":8080"),
		Mux:     mux,
	}); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
