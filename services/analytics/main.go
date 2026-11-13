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
	"github.com/jjheffernan/BambuFarmer/services/analytics/internal/fail"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	service := "analytics"
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
	} else {
		slog.Warn("DATABASE_URL unset; rollups disabled")
	}

	if redisURL != "" && pool != nil {
		bus, err := events.Dial(ctx, redisURL)
		if err != nil {
			slog.Warn("redis dial failed", "err", err)
		} else {
			defer bus.Close()
			go consume(ctx, bus, pool)
		}
	}

	mux := httpserver.NewMux(service)
	mux.HandleFunc("GET /api/v1/analytics/utilization", func(w http.ResponseWriter, r *http.Request) {
		win, err := parseDayWindow(r)
		if err != nil {
			httpserver.WriteError(w, api.BadRequest(err.Error()))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		series := []map[string]any{}
		if pool != nil {
			series = queryUtilization(r.Context(), pool, win, groupBy(r))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"series": series, "group_by": groupBy(r)})
	})
	mux.HandleFunc("GET /api/v1/analytics/costs", func(w http.ResponseWriter, r *http.Request) {
		win, err := parseDayWindow(r)
		if err != nil {
			httpserver.WriteError(w, api.BadRequest(err.Error()))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		series := []map[string]any{}
		if pool != nil {
			series = queryCosts(r.Context(), pool, win)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"series": series})
	})
	mux.HandleFunc("GET /api/v1/analytics/failures", func(w http.ResponseWriter, r *http.Request) {
		win, err := parseDayWindow(r)
		if err != nil {
			httpserver.WriteError(w, api.BadRequest(err.Error()))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		series := []map[string]any{}
		if pool != nil {
			series = queryFailures(r.Context(), pool, win)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"series": series})
	})
	mux.HandleFunc("GET /api/v1/analytics/counters", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		out := map[string]int64{}
		if pool != nil {
			rows, err := pool.Query(r.Context(), `SELECT key, value FROM analytics_counters`)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					var k string
					var v int64
					if err := rows.Scan(&k, &v); err == nil {
						out[k] = v
					}
				}
			}
		}
		_ = json.NewEncoder(w).Encode(out)
	})
	mux.HandleFunc("GET /api/v1/analytics/export", func(w http.ResponseWriter, r *http.Request) {
		win, err := parseDayWindow(r)
		if err != nil {
			httpserver.WriteError(w, api.BadRequest(err.Error()))
			return
		}
		bundle := exportBundle{
			Utilization: []map[string]any{},
			Costs:       []map[string]any{},
			Failures:    []map[string]any{},
		}
		if pool != nil {
			bundle.Utilization = queryUtilization(r.Context(), pool, win, groupBy(r))
			bundle.Costs = queryCosts(r.Context(), pool, win)
			bundle.Failures = queryFailures(r.Context(), pool, win)
		}
		if err := writeExport(w, exportFormat(r), exportKind(r), bundle); err != nil {
			httpserver.WriteError(w, api.Internal("export failed: "+err.Error()))
		}
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

func consume(ctx context.Context, bus *events.Bus, pool *pgxpool.Pool) {
	for _, stream := range []string{events.StreamPrinter, events.StreamJob, events.StreamSlice, events.StreamAlert} {
		s := stream
		go func() {
			consumer := "analytics-" + getenv("HOSTNAME", "1")
			_ = bus.ConsumeLoop(ctx, s, events.GroupAnalytics, consumer, "$", func(ctx context.Context, _ string, env events.Envelope) error {
				return bump(ctx, pool, env)
			})
		}()
	}
	<-ctx.Done()
}

func bump(ctx context.Context, pool *pgxpool.Pool, env events.Envelope) error {
	key := "events." + env.Type
	_, err := pool.Exec(ctx, `
		INSERT INTO analytics_counters(key, value, updated_at) VALUES ($1, 1, now())
		ON CONFLICT (key) DO UPDATE SET value = analytics_counters.value + 1, updated_at = now()`, key)
	if err != nil {
		return err
	}
	if env.Type == events.TypePrinterStateChanged || env.Type == events.TypePrinterOnline || env.Type == events.TypePrinterOffline {
		sn, _ := env.Payload["external_sn"].(string)
		site := env.SiteID
		if site == "" {
			site = "default"
		}
		state, _ := env.Payload["gcode_state"].(string)
		idle, run, failSec := 0, 0, 0
		switch state {
		case "RUNNING", "PREPARE", "PAUSE":
			run = 3 // poll interval approx
		case "FAILED":
			failSec = 3
		default:
			idle = 3
		}
		_, err = pool.Exec(ctx, `
			INSERT INTO agg_utilization_daily(day, site_id, external_sn, idle_s, run_s, fail_s, samples)
			VALUES (CURRENT_DATE, $1, $2, $3, $4, $5, 1)
			ON CONFLICT (day, site_id, external_sn) DO UPDATE SET
				idle_s = agg_utilization_daily.idle_s + EXCLUDED.idle_s,
				run_s = agg_utilization_daily.run_s + EXCLUDED.run_s,
				fail_s = agg_utilization_daily.fail_s + EXCLUDED.fail_s,
				samples = agg_utilization_daily.samples + 1`,
			site, sn, idle, run, failSec)
		for _, code := range fail.Codes(env.Payload) {
			_, err = pool.Exec(ctx, `
				INSERT INTO agg_hms_daily(day, site_id, code, count)
				VALUES (CURRENT_DATE, $1, $2, 1)
				ON CONFLICT (day, site_id, code) DO UPDATE SET
					count = agg_hms_daily.count + 1`,
				site, code)
			if err != nil {
				return err
			}
		}
	}
	if env.Type == events.TypeJobFinished || env.Type == events.TypeJobFailed {
		jobID, _ := env.Payload["job_id"].(string)
		site := env.SiteID
		if site == "" {
			site = "default"
		}
		sn, _ := env.Payload["external_sn"].(string)
		success := env.Type == events.TypeJobFinished
		_, err = pool.Exec(ctx, `
			INSERT INTO fact_job_completed(job_id, site_id, external_sn, success, completed_at)
			VALUES ($1,$2,$3,$4,now())
			ON CONFLICT (job_id) DO NOTHING`, jobID, site, sn, success)
	}
	if env.Type == events.TypeJobCosted {
		jobID, _ := env.Payload["job_id"].(string)
		site := env.SiteID
		if site == "" {
			site = "default"
		}
		sn, _ := env.Payload["external_sn"].(string)
		seconds := intNum(env.Payload["duration_s"])
		grams := floatNum(env.Payload["grams"])
		actual := floatNum(env.Payload["cost_actual"])
		material := floatNum(env.Payload["material_cost"])
		machine := floatNum(env.Payload["machine_cost"])
		labor := floatNum(env.Payload["labor_cost"])
		_, err = pool.Exec(ctx, `
			INSERT INTO fact_job_completed(job_id, site_id, external_sn, duration_s, success, cost_actual, grams, completed_at)
			VALUES ($1,$2,$3,$4,true,$5,$6,now())
			ON CONFLICT (job_id) DO UPDATE SET
				external_sn = EXCLUDED.external_sn,
				duration_s = EXCLUDED.duration_s,
				cost_actual = EXCLUDED.cost_actual,
				grams = EXCLUDED.grams,
				success = true`,
			jobID, site, sn, seconds, actual, grams)
		if err != nil {
			return err
		}
		_, err = pool.Exec(ctx, `
			INSERT INTO agg_cost_daily(day, site_id, material_cost, machine_cost, labor_cost, jobs)
			VALUES (CURRENT_DATE, $1, $2, $3, $4, 1)
			ON CONFLICT (day, site_id) DO UPDATE SET
				material_cost = agg_cost_daily.material_cost + EXCLUDED.material_cost,
				machine_cost = agg_cost_daily.machine_cost + EXCLUDED.machine_cost,
				labor_cost = agg_cost_daily.labor_cost + EXCLUDED.labor_cost,
				jobs = agg_cost_daily.jobs + 1`,
			site, material, machine, labor)
	}
	return err
}

func floatNum(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case int64:
		return float64(n)
	default:
		return 0
	}
}

func intNum(v any) int {
	return int(floatNum(v))
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
