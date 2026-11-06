# BambuFarmer

Operations layer for a Bambu Lab print farm: Farm Manager integration, jobs, presets, Spoolman/Manyfold adapters, costing, and analytics.

**Status:** design freeze in progress + scaffold stubs.  
See [`docs/DESIGN_SPEC.md`](docs/DESIGN_SPEC.md), [`docs/MVP_CHECKLIST.md`](docs/MVP_CHECKLIST.md), and per-service shapes in [`docs/services/`](docs/services/).  
Edge / Farm Manager research: [`docs/services/edge-farm-manager.md`](docs/services/edge-farm-manager.md).

[![CI](https://github.com/jjheffernan/BambuFarmer/actions/workflows/ci.yml/badge.svg)](https://github.com/jjheffernan/BambuFarmer/actions/workflows/ci.yml)

## Modules (one container each)

| Service | Image | Host port | Role |
|---------|-------|-----------|------|
| `gateway` | `localhost/bambufarmer-gateway` | 18080 | Auth + API ingress |
| `core` | `localhost/bambufarmer-core` | 18081 | Jobs, presets, domain |
| `edge` | `localhost/bambufarmer-edge` | 18082 | Farm Manager LAN adapter |
| `integrations` | `localhost/bambufarmer-integrations` | 18083 | Spoolman / Manyfold |
| `analytics` | `localhost/bambufarmer-analytics` | 18084 | Rollups / reports |
| `web` | `localhost/bambufarmer-web` | 13000 | templ + HTMX UI |
| Postgres / Redis | upstream images | 5432 / 6379 | Backing stores |

Each app service is a **separate distroless image** with a single Go binary (`services/<name>/Containerfile`). We do not ship an all-in-one app container.

Host ports use the `1xxxx` range to avoid clashing with common local apps on 8080/3000.

## Quick start (Podman)

```bash
cp deploy/env.example deploy/.env   # AUTH_MODE=lan by default; optional Spoolman/Manyfold URLs
make up                             # uses: podman compose -f deploy/compose.yml
make health
open http://127.0.0.1:13000
```

Default auth is **LAN mode** (`AUTH_MODE=lan`) — local users, no IdP, private network only. Use `AUTH_MODE=oidc` when you have a real URL.

Docker works as a fallback: `make up ENGINE=docker`.

Build one module only:

```bash
podman build -t localhost/bambufarmer-edge:dev -f services/edge/Containerfile .
# or: make images
```

Systemd / Quadlets: see [`deploy/podman/README.md`](deploy/podman/README.md).

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
