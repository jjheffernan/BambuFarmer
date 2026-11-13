package main

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func queryUtilization(ctx context.Context, pool *pgxpool.Pool, win dayWindow, gb string) []map[string]any {
	series := []map[string]any{}
	sql, args := utilizationSQL(win, gb)
	rows, err := pool.Query(ctx, sql, args...)
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

func utilizationSQL(win dayWindow, gb string) (string, []any) {
	switch gb {
	case "site":
		base := `
			SELECT MAX(day)::text AS day, site_id, '' AS external_sn,
				SUM(idle_s)::int, SUM(run_s)::int, SUM(fail_s)::int, SUM(samples)::int
			FROM agg_utilization_daily`
		where, args := dayWhere(win)
		return base + where + ` GROUP BY site_id ORDER BY day DESC LIMIT 100`, args
	case "day":
		base := `
			SELECT day::text, '' AS site_id, '' AS external_sn,
				SUM(idle_s)::int, SUM(run_s)::int, SUM(fail_s)::int, SUM(samples)::int
			FROM agg_utilization_daily`
		where, args := dayWhere(win)
		return base + where + ` GROUP BY day ORDER BY day DESC LIMIT 100`, args
	default: // printer
		base := `
			SELECT day::text, site_id, external_sn, idle_s, run_s, fail_s, samples
			FROM agg_utilization_daily`
		where, args := dayWhere(win)
		limit := " LIMIT 30"
		if win.Set {
			limit = " LIMIT 500"
		}
		return base + where + ` ORDER BY day DESC` + limit, args
	}
}

func queryCosts(ctx context.Context, pool *pgxpool.Pool, win dayWindow) []map[string]any {
	series := []map[string]any{}
	base := `
		SELECT day::text, site_id, material_cost, machine_cost, labor_cost, jobs
		FROM agg_cost_daily`
	where, args := dayWhere(win)
	limit := " LIMIT 30"
	if win.Set {
		limit = " LIMIT 500"
	}
	rows, err := pool.Query(ctx, base+where+` ORDER BY day DESC`+limit, args...)
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

func queryFailures(ctx context.Context, pool *pgxpool.Pool, win dayWindow) []map[string]any {
	series := []map[string]any{}
	base := `
		SELECT day::text, site_id, code, count
		FROM agg_hms_daily`
	where, args := dayWhere(win)
	limit := " LIMIT 100"
	if win.Set {
		limit = " LIMIT 500"
	}
	rows, err := pool.Query(ctx, base+where+` ORDER BY day DESC, count DESC`+limit, args...)
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

func dayWhere(win dayWindow) (string, []any) {
	if !win.Set {
		return "", nil
	}
	return " WHERE day >= $1::date AND day <= $2::date", []any{
		win.From.Format("2006-01-02"),
		win.To.Format("2006-01-02"),
	}
}
