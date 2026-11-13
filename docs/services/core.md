# Service shape: `bambufarmer-core`

**Language:** Go  
**Container:** `localhost/bambufarmer-core`  
**Role:** System of record for jobs, presets, rate cards, printer registry projections  

---

## 1. Responsibility

| Owns | Does not own |
|------|----------------|
| Postgres OLTP schema (with `packages/db`) | Talking to Farm Manager |
| Versioned filament/print presets | Spoolman/Manyfold HTTP |
| Rate cards (`$/hr`, `$/kg`, `kg/hr`) | Heavy analytics rollups |
| Job state machine + scheduler policy | UI rendering |
| Printer registry (last snapshot from events) | Live MQTT |

---

## 2. Domain modules

`main.go` is wiring only. HTTP + consumers live under `internal/` (unique to core; other services use `app/`):

```text
services/core/
  main.go                 # boot: db/redis, register routes, listen
  internal/               # process-private (Go internal import rule)
    config/               # SITE_ID, DATABASE_URL, REDIS_URL, …
    httpx/                # WriteJSON, RequireDB
    presets/              # filament/print CRUD + import/export
    rates/                # RateCard admin
    printers/             # list projection + Redis consumer
    jobs/                 # create/list/transition (+ command emit)
    setup/                # install wizard API
    # later: scheduler/, users/
```

Persistence stays in [`packages/db`](../../packages/db); core packages are thin HTTP/event adapters.

---

## 3. HTTP API (core-local; gateway proxies)

### Done / partial

- `GET/POST /api/v1/presets/filament`  
- `GET/PATCH/DELETE /api/v1/presets/filament/{id}`  
- `GET/POST /api/v1/presets/print`  
- `GET/PATCH/DELETE /api/v1/presets/print/{id}`  
- `GET /api/v1/presets/export`  
- `POST /api/v1/presets/import`  

### Next

- (none blocking)  

### Recent

- Scheduler stagger (`SCHEDULER_STAGGER` / `SCHEDULER_MAX_STARTS_PER_TICK`)
- Optional MinIO/S3 artifact mirror via `OBJECT_STORE_*` (compose `--profile object`)
- Preset revision list/restore (`GET …/revisions`, `POST …/restore`)
- `GET/PUT /api/v1/admin/rate-cards`  
- `GET /api/v1/printers` (projection)  
- Jobs create/list/dispatch + scheduler  
- `POST/GET /api/v1/artifacts` (+ `/content`) — local `ARTIFACTS_DIR` staging

---

## 4. Job state machine (reminder)

`drafted → queued → assigned → dispatched → printing ⇄ paused → succeeded|failed → closed`

On assign: write `bf.commands.edge.{site}` with upload/start payload.  
On edge `job.*` events: advance state; finalize `JobCost` using rate card.

Costing:

```text
material = kg * ($/kg)
machine  = hours * ($/hr)
estimate_kg = (kg/hr) * hours   # when grams unknown
```

---

## 5. Config

`DATABASE_URL` (required), `REDIS_URL`, `SITE_ID`, `OBJECT_STORE_*`, `HTTP_ADDR`.

On boot: `db.Migrate`.

---

## 6. MVP milestones

1. Presets + baseline schema (done; see [`DATABASE.md`](../DATABASE.md))  
2. Rate cards CRUD  
3. Printer projection consumer  
4. Jobs create → command edge → track finish  
5. Artifact upload to MinIO/S3 — done when `OBJECT_STORE_ENDPOINT` is set (`--profile object`)
6. Lan users for gateway
