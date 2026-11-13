# Service shape: `bambufarmer-analytics`

**Language:** Go  
**Container:** `localhost/bambufarmer-analytics`  
**Role:** Dedicated read models / rollups — never blocks job OLTP  

---

## 1. Responsibility

| Owns | Does not own |
|------|----------------|
| Consume Redis Streams → fact/agg tables | Job transitions |
| Dashboard query APIs | Rate card writes |
| CSV/JSON export | FM live control |

---

## 2. Ingest

Subscribe (consumer group `analytics`):

- `bf.events.printer`  
- `bf.events.job`  
- `bf.events.slice`  
- `bf.events.alert`  

(`bf.events.filament` is reserved; analytics does not subscribe yet.)

Idempotent by event `id`.

---

## 3. Schema (target, analytics DB or shared Postgres schema `analytics`)

- `fact_job_completed`  
- `agg_utilization_daily` (by printer/group/site)  
- `agg_hms_daily`  
- `agg_cost_daily` (material, machine; expose $/hr, $/kg, kg/hr)  

---

## 4. HTTP

| Path | Notes |
|------|-------|
| `GET /api/v1/analytics/utilization` | optional `from`/`to` (`YYYY-MM-DD`, both required together); `group_by=printer\|site\|day` |
| `GET /api/v1/analytics/costs` | same optional `from`/`to` |
| `GET /api/v1/analytics/failures` | same optional `from`/`to` |
| `GET /api/v1/analytics/export` | JSON (default) or CSV via `format=` / `Accept: text/csv`; optional `kind=utilization\|costs\|failures\|all`; honors `from`/`to`/`group_by` |

Without `from`/`to`, series use a short recent limit. Utilization and failures read from Postgres rollups when events have been consumed. Empty series when the database is unset.

---

## 5. Package layout

```text
services/analytics/
  main.go
  app/          # process-private (when split out of main)
    consume/
    store/
    query/
```

---

## 6. MVP milestones

1. Health + empty utilization (done)  
2. Consumer writing utilization intervals (done)  
3. Job success/fail counts + `print_error` → `agg_hms_daily` / failures API (done)  
4. Cost rollups from `job.costed` → `agg_cost_daily` / costs API (done)  
5. Basic charts via web (done)  
6. CSV/JSON export (done)
