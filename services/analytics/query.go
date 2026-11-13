package main

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func queryUtilization(ctx context.Context, pool *pgxpool.Pool) []map[string]any {
	series := []map[string]any{}
	rows, err := pool.Query(ctx, `
		SELECT day::text, site_id, external_sn, idle_s, run_s, fail_s, samples
		FROM agg_utilization_daily ORDER BY day DESC LIMIT 30`)
	if err != nil {
		return series
	}
	defer rows.Close()
	for rows.Next() {
		var day, site, sn string
		var idle, run, failSec, samples int
		if err := rows.Scan(&day, &site, &sn, &idle, &run, &failSec, &samples); err == nil {
			series = append(series, map[string]any{
				"day": day, "site_id": site, "external_sn": sn,
				"idle_s": idle, "run_s": run, "fail_s": failSec, "samples": samples,
			})
		}
	}
	return series
}

func queryCosts(ctx context.Context, pool *pgxpool.Pool) []map[string]any {
	series := []map[string]any{}
	rows, err := pool.Query(ctx, `
		SELECT day::text, site_id, material_cost, machine_cost, labor_cost, jobs
		FROM agg_cost_daily ORDER BY day DESC LIMIT 30`)
	if err != nil {
		return series
	}
	defer rows.Close()
	for rows.Next() {
		var day, site string
		var material, machine, labor float64
		var jobs int
		if err := rows.Scan(&day, &site, &material, &machine, &labor, &jobs); err == nil {
			series = append(series, map[string]any{
				"day": day, "site_id": site,
				"material_cost": material, "machine_cost": machine, "labor_cost": labor, "jobs": jobs,
			})
		}
	}
	return series
}

func queryFailures(ctx context.Context, pool *pgxpool.Pool) []map[string]any {
	series := []map[string]any{}
	rows, err := pool.Query(ctx, `
		SELECT day::text, site_id, code, count
		FROM agg_hms_daily ORDER BY day DESC, count DESC LIMIT 100`)
	if err != nil {
		return series
	}
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
	return series
}
