# Service design shapes

Per-module design docs that expand [`DESIGN_SPEC.md`](../DESIGN_SPEC.md) into implementable shapes. Each service is **one Go binary → one Containerfile → one container** (Podman-first).

| Service | Doc | Role |
|---------|-----|------|
| **gateway** | [gateway.md](./gateway.md) | Auth, reverse-proxy, public HTTP surface |
| **core** | [core.md](./core.md) | Jobs, presets, rate cards, printer registry |
| **edge** | [edge.md](./edge.md) | Farm Manager LAN adapter + event ingest |
| **edge / FM research** | [edge-farm-manager.md](./edge-farm-manager.md) | Protocol research, endpoints, lab checklist |
| **integrations** | [integrations.md](./integrations.md) | Spoolman + Manyfold adapters |
| **analytics** | [analytics.md](./analytics.md) | Event rollups, reports |
| **web** | [web.md](./web.md) | HTMX + Tailwind operator UI |
| **slicer** | [slicer.md](./slicer.md) | Post-MVP OrcaSlicer wrapper |

Cross-cutting: [`DEPLOYMENT_MODES.md`](../DEPLOYMENT_MODES.md) (LAN vs OIDC) · [`INSTALL.md`](../INSTALL.md) (wizard + releases) · [`TESTING.md`](../TESTING.md) (module `tests/` layout).

## Layout convention

| Path | Meaning |
|------|---------|
| `packages/*` | Shared libraries (API DTOs, db, events, farm abstraction, …) |
| `services/core/internal/*` | Core process-private modules (unique name; Go `internal` rule) |
| `services/gateway/auth/*` | Gateway auth helpers |
| `services/web/setup/*` | Web install-wizard pages |
| `services/<name>/app/*` | Process-private modules for edge, integrations |
| `services/<name>/tests/*` | Black-box unit tests for that module |

Do not put service-private code under `packages/`. Edge Farm Manager HTTP lives in `services/edge/app/farmmanager` (distinct from shared `packages/farm`).

## Cross-cutting contracts

- **Events:** [`packages/events`](../../packages/events) — Redis Streams keys + envelope
- **API DTOs:** [`packages/api`](../../packages/api)
- **DB:** [`packages/db`](../../packages/db) — baseline + forward migrations; see [`DATABASE.md`](../DATABASE.md)
- **Config / secrets:** [`.env.example`](../../.env.example) → `.env` (gitignored)
- **Bambu integration:** [`packages/farm`](../../packages/farm) — `ControlPlane` abstraction (edge-only implementations)

## Dependency direction

```text
web ──► gateway ──► core
                 ├─► analytics
                 ├─► integrations
                 └─► edge ──► Farm Manager Server (LAN)
core ◄── Redis Streams ◄── edge / integrations
analytics ◄── Redis Streams
```

Edge never serves browser traffic. Gateway is the only intended public/LAN HTTP entry for operators.
