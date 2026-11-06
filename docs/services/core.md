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

`main.go` is wiring only. HTTP + consumers live under `internal/`:

```text
services/core/
  main.go                 # boot: db/redis, register routes, listen
  internal/
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
- `GET/POST /api/v1/presets/print`  
- `GET /api/v1/presets/export`  
- `POST /api/v1/presets/import`  

### Next

- `GET/PUT /api/v1/admin/rate-cards`  
- `GET /api/v1/printers` (projection)  
- `CRUD /api/v1/jobs` + queue/cancel/reprint  
- `POST /api/v1/artifacts` (object store + metadata)  
- Preset revision list/restore endpoints  

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

1. Presets + migrations (done)  
2. Rate cards CRUD  
3. Printer projection consumer  
4. Jobs create → command edge → track finish  
5. Artifact upload to MinIO/S3  
6. Lan users for gateway
