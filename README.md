# BambuFarmer

Operations layer for a Bambu Lab print farm: Farm Manager integration, jobs, presets, Spoolman/Manyfold adapters, costing, and analytics.

**Status:** MVP farm loop on `dev` (fixture/unit covered; live Farm Manager smoke still deferred).  
See [`docs/DESIGN_SPEC.md`](docs/DESIGN_SPEC.md), [`docs/MVP_CHECKLIST.md`](docs/MVP_CHECKLIST.md), [`docs/BUILD_OUT.md`](docs/BUILD_OUT.md), and per-service shapes in [`docs/services/`](docs/services/).  
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
| `web` | `localhost/bambufarmer-web` | 13000 | Operator UI (`html/template` + SSE + Tailwind) |
| Postgres / Redis | upstream images | 5432 / 6379 | Backing stores |

Optional compose profiles: `commerce`, `slicer`, `object` (MinIO).

Each app service is a **separate distroless image** with a single Go binary (`services/<name>/Containerfile`). We do not ship an all-in-one app container.

Host ports use the `1xxxx` range to avoid clashing with common local apps on 8080/3000.

## Quick start (Podman)

```bash
cp .env.example .env   # never commit .env (gitignored)
# edit .env: SESSION_SECRET, POSTGRES_PASSWORD, and optional FARM_MANAGER_* / OIDC_*
make up                # creates .env if missing; podman compose --env-file .env -f deploy/compose.yml
make health
open http://127.0.0.1:18080   # first visit → /setup install wizard

# later: one-command upgrade (volumes kept) — also Admin → Request update + bf update --watch
./scripts/bf update
```

Do **not** put real passwords, tokens, or private LAN addresses in the README, docs examples, or committed config. Use `.env` locally; commit only `.env.example` placeholders (`change-me-…`). Schema lives in Go under `packages/db` — `*.sql` files are gitignored.

Install wizard + releases: [`docs/INSTALL.md`](docs/INSTALL.md) · [`CHANGELOG.md`](CHANGELOG.md) (Keep a Changelog — update **Unreleased** with each notable PR).  
Build-out / plugins: [`docs/BUILD_OUT.md`](docs/BUILD_OUT.md) · [`docs/plugins.md`](docs/plugins.md).

**Auth / deploy modes:** default is **LAN mode** (`AUTH_MODE=lan`) — local users, no IdP, **private network only**. Use `AUTH_MODE=oidc` when the stack has a real URL or reverse proxy. Full comparison: [`docs/DEPLOYMENT_MODES.md`](docs/DEPLOYMENT_MODES.md).

With `FARM_MANAGER_URL` empty, edge runs **fixture mode** (`fixtures/fm` → Redis → gateway SSE → fleet UI) so you can exercise the spine without a live Farm Manager.

Docker works as a fallback: `make up ENGINE=docker`.

Build one module only:

```bash
podman build -t localhost/bambufarmer-edge:dev -f services/edge/Containerfile .
# or: make images
```

Systemd / Quadlets: see [`deploy/podman/README.md`](deploy/podman/README.md).

Import example presets into core:

```bash
curl -sS -X POST http://127.0.0.1:18081/api/v1/presets/import \
  -H 'Content-Type: application/json' \
  --data-binary @fixtures/presets/example_bundle.json
curl -sS http://127.0.0.1:18081/api/v1/presets/export | jq .
```

## Local Go (optional)

Install Go 1.26+, then:

```bash
make tidy
cd services/web && HTTP_ADDR=:3000 go run .
```

## Live Farm Manager

Product decisions in checklist **A** are locked. Remaining gate before calling FM integration “production-proven” is a live LAN smoke (checklist **C** integration readiness) — fixture mode covers CI and local demos without FM.
