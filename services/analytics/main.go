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
		w.Header().Set("Content-Type", "application/json")
		series := []map[string]any{}
		if pool != nil {
			rows, err := pool.Query(r.Context(), `
				SELECT day::text, site_id, external_sn, idle_s, run_s, fail_s, samples
				FROM agg_utilization_daily ORDER BY day DESC LIMIT 30`)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					var day, site, sn string
					var idle, run, fail, samples int
					if err := rows.Scan(&day, &site, &sn, &idle, &run, &fail, &samples); err == nil {
						series = append(series, map[string]any{
							"day": day, "site_id": site, "external_sn": sn,
							"idle_s": idle, "run_s": run, "fail_s": fail, "samples": samples,
						})
					}
				}
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"series": series})
	})
	mux.HandleFunc("GET /api/v1/analytics/failures", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		series := []map[string]any{}
		if pool != nil {
			rows, err := pool.Query(r.Context(), `
				SELECT day::text, site_id, code, count
				FROM agg_hms_daily ORDER BY day DESC, count DESC LIMIT 100`)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					var day, site, code string
					var n int
					if err := rows.Scan(&day, &site, &code, &n); err == nil {
						series = append(series, map[string]any{
							"day": day, "site_id": site, "code": code, "count": n,
						})
					}
				}
			}
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
	for _, stream := range []string{events.StreamPrinter, events.StreamJob} {
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
		if code := fail.Code(env.Payload); code != "" {
			_, err = pool.Exec(ctx, `
				INSERT INTO agg_hms_daily(day, site_id, code, count)
				VALUES (CURRENT_DATE, $1, $2, 1)
				ON CONFLICT (day, site_id, code) DO UPDATE SET
					count = agg_hms_daily.count + 1`,
				site, code)
		}
	}
	if env.Type == events.TypeJobFinished {
		jobID, _ := env.Payload["job_id"].(string)
		site := env.SiteID
		if site == "" {
			site = "default"
		}
		_, _ = pool.Exec(ctx, `
			INSERT INTO fact_job_completed(job_id, site_id, success, completed_at)
			VALUES ($1,$2,true,now())
			ON CONFLICT (job_id) DO NOTHING`, jobID, site)
	}
	return err
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
