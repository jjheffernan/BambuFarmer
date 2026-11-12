package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jjheffernan/BambuFarmer/packages/db"
	"github.com/jjheffernan/BambuFarmer/packages/events"
	"github.com/jjheffernan/BambuFarmer/packages/httpserver"
	"github.com/jjheffernan/BambuFarmer/packages/version"
	"github.com/jjheffernan/BambuFarmer/services/core/internal/config"
	"github.com/jjheffernan/BambuFarmer/services/core/internal/httpx"
	"github.com/jjheffernan/BambuFarmer/services/core/internal/jobs"
	"github.com/jjheffernan/BambuFarmer/services/core/internal/presets"
	"github.com/jjheffernan/BambuFarmer/services/core/internal/printers"
	"github.com/jjheffernan/BambuFarmer/services/core/internal/rates"
	"github.com/jjheffernan/BambuFarmer/services/core/internal/setup"
	"github.com/jjheffernan/BambuFarmer/services/core/internal/users"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()

	var pool *pgxpool.Pool
	if cfg.DatabaseURL != "" {
		var err error
		pool, err = db.Open(ctx, cfg.DatabaseURL)
		if err != nil {
			slog.Error("database open failed", "err", err)
			os.Exit(1)
		}
		defer pool.Close()
		if err := db.Migrate(ctx, pool); err != nil {
			slog.Error("migrate failed", "err", err)
			os.Exit(1)
		}
		slog.Info("database ready", "site_id", cfg.SiteID)
	} else {
		slog.Warn("DATABASE_URL unset; APIs will return 503")
	}

	var bus *events.Bus
	if cfg.RedisURL != "" {
		var err error
		bus, err = events.Dial(ctx, cfg.RedisURL)
		if err != nil {
			slog.Warn("redis unavailable; consumers/scheduler command publish disabled", "err", err)
		} else {
			defer bus.Close()
			go printers.ConsumePrinterEvents(ctx, bus, pool, cfg.SiteID, "core-printers-"+cfg.Hostname)
			go jobs.ConsumeJobEvents(ctx, bus, pool, cfg.SiteID, "core-jobs-"+cfg.Hostname)
		}
	}
	if pool != nil {
		sched := &jobs.Scheduler{Pool: pool, Bus: bus, SiteID: cfg.SiteID}
		go sched.Run(ctx)
	}

	mux := httpserver.NewMux("core")
	presets.Register(mux, pool, cfg.SiteID)
	rates.Register(mux, pool, cfg.SiteID)
	jobs.Register(mux, pool, bus, cfg.SiteID)
	printers.Register(mux, pool, cfg.SiteID)
	setup.Register(mux, pool, cfg.SiteID)
	users.Register(mux, pool)
	mux.HandleFunc("GET /api/v1/version", func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteJSON(w, version.Get())
	})

	addr := cfg.HTTPAddr
	if addr == "" {
		addr = httpserver.DefaultAddr()
	}
	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		slog.Info("listening", "service", "core", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server stopped", "err", err)
			os.Exit(1)
		}
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}
