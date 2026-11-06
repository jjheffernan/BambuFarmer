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

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/db"
	"github.com/jjheffernan/BambuFarmer/packages/events"
	"github.com/jjheffernan/BambuFarmer/packages/httpserver"
	"github.com/jjheffernan/BambuFarmer/packages/version"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	service := "core"
	siteID := getenv("SITE_ID", "default")
	databaseURL := os.Getenv("DATABASE_URL")
	redisURL := os.Getenv("REDIS_URL")

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
		slog.Warn("DATABASE_URL unset; APIs will return 503")
	}

	var bus *events.Bus
	if redisURL != "" {
		var err error
		bus, err = events.Dial(ctx, redisURL)
		if err != nil {
			slog.Warn("redis unavailable; printer projection consumer disabled", "err", err)
		} else {
			defer bus.Close()
			go consumePrinterEvents(ctx, bus, pool, siteID)
		}
	}

	mux := httpserver.NewMux(service)
	registerPresetRoutes(mux, pool, siteID)
	registerRateCardRoutes(mux, pool, siteID)
	registerJobRoutes(mux, pool, bus, siteID)
	registerPrinterRoutes(mux, pool, siteID)
	registerSetupRoutes(mux, pool, siteID)
	mux.HandleFunc("GET /api/v1/version", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, version.Get())
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

func registerPresetRoutes(mux *http.ServeMux, pool *pgxpool.Pool, siteID string) {
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
}

func registerRateCardRoutes(mux *http.ServeMux, pool *pgxpool.Pool, siteID string) {
	mux.HandleFunc("GET /api/v1/admin/rate-cards", func(w http.ResponseWriter, r *http.Request) {
		requireDB(w, pool, func() {
			rc, err := db.GetRateCard(r.Context(), pool, siteID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, rc)
		})
	})
	mux.HandleFunc("PUT /api/v1/admin/rate-cards", func(w http.ResponseWriter, r *http.Request) {
		requireDB(w, pool, func() {
			var rc api.RateCard
			if err := json.NewDecoder(r.Body).Decode(&rc); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			out, err := db.UpsertRateCard(r.Context(), pool, siteID, rc)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, out)
		})
	})
}

func registerJobRoutes(mux *http.ServeMux, pool *pgxpool.Pool, bus *events.Bus, siteID string) {
	mux.HandleFunc("GET /api/v1/jobs", func(w http.ResponseWriter, r *http.Request) {
		requireDB(w, pool, func() {
			items, err := db.ListJobs(r.Context(), pool, siteID, 50)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if items == nil {
				items = []api.Job{}
			}
			writeJSON(w, items)
		})
	})
	mux.HandleFunc("POST /api/v1/jobs", func(w http.ResponseWriter, r *http.Request) {
		requireDB(w, pool, func() {
			var req api.CreateJobRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			job, err := db.CreateJob(r.Context(), pool, siteID, req)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if rc, err := db.GetRateCard(r.Context(), pool, siteID); err == nil {
				est := db.EstimateJobCost(rc, job.EstSeconds, job.EstGrams, job.Copies)
				_ = db.ApplyJobCost(r.Context(), pool, siteID, job.ID, est, 0, rc.Currency)
				job.CostEstimate = est
				job.Currency = rc.Currency
			}
			if bus != nil {
				env := events.NewEnvelope(events.TypeJobQueued, "core", siteID, map[string]any{
					"job_id": job.ID, "name": job.Name, "external_sn": job.ExternalSN,
				})
				_, _ = bus.Publish(r.Context(), events.StreamJob, env)
				if job.ExternalSN != "" {
					cmd := events.NewEnvelope("command.start_print", "core", siteID, map[string]any{
						"job_id": job.ID, "external_sn": job.ExternalSN,
					})
					_, _ = bus.Publish(r.Context(), events.CommandStream(siteID), cmd)
				}
			}
			writeJSON(w, job)
		})
	})
	mux.HandleFunc("GET /api/v1/jobs/{id}", func(w http.ResponseWriter, r *http.Request) {
		requireDB(w, pool, func() {
			job, err := db.GetJob(r.Context(), pool, siteID, r.PathValue("id"))
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			writeJSON(w, job)
		})
	})
	mux.HandleFunc("POST /api/v1/jobs/{id}/transition", func(w http.ResponseWriter, r *http.Request) {
		requireDB(w, pool, func() {
			var body struct {
				State  string `json:"state"`
				Reason string `json:"reason"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			job, err := db.TransitionJob(r.Context(), pool, siteID, r.PathValue("id"), body.State, "api", body.Reason)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			writeJSON(w, job)
		})
	})
}

func registerPrinterRoutes(mux *http.ServeMux, pool *pgxpool.Pool, siteID string) {
	mux.HandleFunc("GET /api/v1/printers", func(w http.ResponseWriter, r *http.Request) {
		requireDB(w, pool, func() {
			items, err := db.ListPrinters(r.Context(), pool, siteID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if items == nil {
				items = []api.PrinterSummary{}
			}
			writeJSON(w, items)
		})
	})
}

func registerSetupRoutes(mux *http.ServeMux, pool *pgxpool.Pool, siteID string) {
	mux.HandleFunc("GET /api/v1/setup", func(w http.ResponseWriter, r *http.Request) {
		requireDB(w, pool, func() {
			st, err := db.GetInstallState(r.Context(), pool)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			rc, _ := db.GetRateCard(r.Context(), pool, siteID)
			writeJSON(w, map[string]any{
				"install":  st,
				"rate_card": rc,
				"version":  version.Get(),
				"site_id":  siteID,
			})
		})
	})
	mux.HandleFunc("POST /api/v1/setup", func(w http.ResponseWriter, r *http.Request) {
		requireDB(w, pool, func() {
			var req api.SetupRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			var (
				st  api.InstallState
				err error
			)
			if req.Complete {
				st, err = db.CompleteInstall(r.Context(), pool, siteID, version.Version, req)
			} else {
				st, err = db.SaveInstallDraft(r.Context(), pool, siteID, req)
			}
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, st)
		})
	})
	mux.HandleFunc("POST /api/v1/setup/reset", func(w http.ResponseWriter, r *http.Request) {
		requireDB(w, pool, func() {
			if err := db.ResetInstall(r.Context(), pool); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			st, err := db.GetInstallState(r.Context(), pool)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, st)
		})
	})
}

func consumePrinterEvents(ctx context.Context, bus *events.Bus, pool *pgxpool.Pool, siteID string) {
	if pool == nil {
		return
	}
	consumer := "core-" + getenv("HOSTNAME", "1")
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
		pct := intFrom(env.Payload["mc_percent"])
		remain := intFrom(env.Payload["mc_remaining_time"])
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

func intFrom(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	default:
		return 0
	}
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
