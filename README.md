# BambuFarmer

Operations layer for a Bambu Lab print farm: Farm Manager integration, jobs, presets, Spoolman/Manyfold adapters, costing, and analytics.

**Status:** design freeze in progress + scaffold stubs.  
See [`docs/DESIGN_SPEC.md`](docs/DESIGN_SPEC.md) and answer [`docs/MVP_CHECKLIST.md`](docs/MVP_CHECKLIST.md) before wiring live Farm Manager.

## Architecture (MVP services)

| Service | Host port (compose) | Role |
|---------|---------------------|------|
| `gateway` | 18080 | Auth + API edge |
| `core` | 18081 | Jobs, presets, domain |
| `edge` | 18082 | Farm Manager adapter |
| `integrations` | 18083 | Spoolman / Manyfold |
| `analytics` | 18084 | Rollups / reports |
| `web` | 13000 | templ + HTMX UI |
| Postgres / Redis / MinIO | 5432 / 6379 / 19000 | Backing stores |

Host ports use the `1xxxx` range to avoid clashing with common local apps on 8080/3000.

All first-party code is **Go**.

## Quick start (Docker)

No local Go required:

```bash
cp deploy/env.example deploy/.env   # edit FM / Spoolman / Manyfold when ready
make up
make health
open http://127.0.0.1:13000
```

Import example presets into the core stub:

```bash
curl -sS -X POST http://127.0.0.1:18081/api/v1/presets/import \
  -H 'Content-Type: application/json' \
  --data-binary @fixtures/presets/example_bundle.json
curl -sS http://127.0.0.1:18081/api/v1/presets/export | jq .
```

## Local Go (optional)

Install Go 1.23+, then:

```bash
make tidy
cd services/web && HTTP_ADDR=:3000 go run .
```

## Design exit criteria

Do not treat Farm Manager integration as “in progress” until checklist section **A** is answered (or defaults accepted) and section **C** product boxes are checked.
