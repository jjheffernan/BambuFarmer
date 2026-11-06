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
| **web** | [web.md](./web.md) | templ + HTMX operator UI |
| **slicer** | [slicer.md](./slicer.md) | Post-MVP OrcaSlicer wrapper |

Cross-cutting: [`DEPLOYMENT_MODES.md`](../DEPLOYMENT_MODES.md) (LAN vs OIDC) · [`INSTALL.md`](../INSTALL.md) (wizard + releases).

## Cross-cutting contracts

- **Events:** [`packages/events`](../../packages/events) — Redis Streams keys + envelope
- **API DTOs:** [`packages/api`](../../packages/api)
- **DB:** [`packages/db`](../../packages/db) — migrations owned with core-facing schema
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
