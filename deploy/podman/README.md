# Podman deployment

**Preferred runtime:** Podman (rootless). Compose for bring-up; Quadlets for systemd installs.

## Module rule

One process → one image → one container:

| Image | Module |
|-------|--------|
| `localhost/bambufarmer-gateway` | auth / ingress API |
| `localhost/bambufarmer-core` | jobs, presets, domain |
| `localhost/bambufarmer-edge` | Farm Manager LAN adapter |
| `localhost/bambufarmer-integrations` | Spoolman / Manyfold |
| `localhost/bambufarmer-analytics` | rollups |
| `localhost/bambufarmer-web` | HTMX UI |

Do **not** combine these into a fat “all-in-one” app container. Backing services (Postgres, Redis, optional MinIO) stay separate too.

## Compose (dev / small farm)

From repo root:

```bash
cp .env.example .env   # secrets; see docs/DEPLOYMENT_MODES.md (LAN vs OIDC)
# or: make env
make up                # podman compose --env-file .env -f deploy/compose.yml up -d --build
```

Optional MinIO: `podman compose --env-file .env -f deploy/compose.yml --profile object up -d`

## Build a single module

```bash
podman build -t localhost/bambufarmer-edge:dev -f services/edge/Containerfile .
```

Each `services/*/Containerfile` only copies that service + the shared packages it needs.

## Quadlets (systemd user units)

Copy `*.container` / `*.network` into `~/.config/containers/systemd/`, then:

```bash
systemctl --user daemon-reload
systemctl --user start bambufarmer-gateway
```

Build images first (`make build`). Point Quadlet `EnvironmentFile=` at the repo-root `.env` (same keys as `.env.example`).
