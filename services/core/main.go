package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/db"
	"github.com/jjheffernan/BambuFarmer/packages/httpserver"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	service := "core"
	siteID := getenv("SITE_ID", "default")
	databaseURL := os.Getenv("DATABASE_URL")

	var pool *pgxpool.Pool
	if databaseURL != "" {
		var err error
		pool, err = db.Open(ctx, databaseURL)
		if err != nil {
			slog.Error("database open failed", "err", err)
			os.Exit(1)
		}
		defer pool.Close()
		if err := db.Migrate(ctx, pool); err != nil {
			slog.Error("migrate failed", "err", err)
			os.Exit(1)
		}
		slog.Info("database ready", "site_id", siteID)
	} else {
		slog.Warn("DATABASE_URL unset; preset APIs will return 503")
	}

	mux := httpserver.NewMux(service)
	mux.HandleFunc("GET /api/v1/presets/filament", func(w http.ResponseWriter, r *http.Request) {
		requireDB(w, pool, func() {
			items, err := db.ListFilamentPresets(r.Context(), pool, siteID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if items == nil {
				items = []api.FilamentPreset{}
			}
			writeJSON(w, items)
		})
	})
	mux.HandleFunc("POST /api/v1/presets/filament", func(w http.ResponseWriter, r *http.Request) {
		requireDB(w, pool, func() {
			var p api.FilamentPreset
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if p.ID == "" || p.Name == "" {
				http.Error(w, "id and name required", http.StatusBadRequest)
				return
			}
			out, err := db.UpsertFilamentPreset(r.Context(), pool, siteID, p)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, out)
		})
	})
	mux.HandleFunc("GET /api/v1/presets/print", func(w http.ResponseWriter, r *http.Request) {
		requireDB(w, pool, func() {
			items, err := db.ListPrintPresets(r.Context(), pool, siteID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if items == nil {
				items = []api.PrintPreset{}
			}
			writeJSON(w, items)
		})
	})
	mux.HandleFunc("POST /api/v1/presets/print", func(w http.ResponseWriter, r *http.Request) {
		requireDB(w, pool, func() {
			var p api.PrintPreset
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if p.ID == "" || p.Name == "" {
				http.Error(w, "id and name required", http.StatusBadRequest)
				return
			}
			out, err := db.UpsertPrintPreset(r.Context(), pool, siteID, p)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, out)
		})
	})
	mux.HandleFunc("GET /api/v1/presets/export", func(w http.ResponseWriter, r *http.Request) {
		requireDB(w, pool, func() {
			bundle, err := db.ExportPresetBundle(r.Context(), pool, siteID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, bundle)
		})
	})
	mux.HandleFunc("POST /api/v1/presets/import", func(w http.ResponseWriter, r *http.Request) {
		requireDB(w, pool, func() {
			var bundle api.PresetBundle
			if err := json.NewDecoder(r.Body).Decode(&bundle); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := db.ImportPresetBundle(r.Context(), pool, siteID, bundle); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})
	})

	srv := &http.Server{
		Addr:              httpserver.DefaultAddr(),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		slog.Info("listening", "service", service, "addr", srv.Addr)
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

func requireDB(w http.ResponseWriter, pool *pgxpool.Pool, next func()) {
	if pool == nil {
		http.Error(w, "database not configured", http.StatusServiceUnavailable)
		return
	}
	next()
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
