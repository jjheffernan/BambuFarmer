# Plugin architecture (lean)

Plugins in BambuFarmer are **not** a dynamic SDK. They are:

1. **Go interfaces** in `packages/*` (stable contracts), and/or  
2. **Optional compose services** behind a profile (own binary, own Containerfile).

Stay lean: one interface, one default adapter, env-driven enablement.

## Control plane adapters

[`packages/farm.ControlPlane`](../packages/farm/plane.go) is the only seam for printer/farm backends.

| Kind | Status | Process |
|------|--------|---------|
| `fm_native` | Primary | `services/edge` |
| `cdp` | Escape hatch | `services/edge` |
| `fixture` | Lab / CI | `services/edge` |
| `direct_printer` | In-process list (`DIRECT_PRINTERS`); no LAN | `services/edge` |
| `fleet_hub` | Reserved | future |

Core, gateway, and web must speak jobs/printers via APIs and events—never backend HTTP types.

## Compose-profile services

Pattern (see `services/commerce`):

1. New module under `services/<name>/` with `main.go`, `go.mod`, `Containerfile`, optional `tests/`.
2. Add to [`go.work`](../go.work).
3. Compose service with `profiles: ["<name>"]` so default `make up` stays unchanged.
4. Gateway proxies `/api/v1/<name>/` only when that URL is configured (or always proxy and let it 502 if down—prefer env URL empty = no route later).
5. Migrations: if the plugin owns tables, add `packages/db/migration_00N_*.go` or a plugin-local migrate later; for stubs, **no tables**.

Enable:

```bash
podman compose --env-file .env -f deploy/compose.yml --profile commerce up -d --build
```

## Event bus

Plugins that react to farm activity should consume [`packages/events`](../packages/events) streams rather than polling core HTTP in a tight loop. Publish domain events with stable type names (`job.finished`, `order.paid`, …).

## Auth

Browser/operator traffic still enters through **gateway**. Plugin services trust the internal network; gateway enforces LAN session / OIDC / admin roles for public paths.

## Adding a plugin checklist

- [ ] Interface or DTO in `packages/api` (and `packages/farm` if control-related)
- [ ] Service skeleton + Containerfile + `go.work` entry
- [ ] Compose profile (default off)
- [ ] Doc row in [`services/README.md`](./services/README.md)
- [ ] Changelog Unreleased bullet
- [ ] CI matrix module entry when tests exist
