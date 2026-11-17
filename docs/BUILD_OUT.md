# Build-out roadmap

Multi-session plan to grow BambuFarmer while staying **lean**, **Go-only**, and **plugin-shaped**.  
Canonical architecture: [`DESIGN_SPEC.md`](./DESIGN_SPEC.md) · plugins: [`plugins.md`](./plugins.md).

## Locked direction

| Topic | Choice |
|-------|--------|
| Language / packaging | Go monorepo; one service → one Containerfile → one container |
| Control plane now | Farm Manager (`fm_native`); fixture / CDP stub / `direct_printer` |
| Control plane later | Deeper FM parity; optional `fleet_hub` if needed |
| Long horizon | OrcaSlicer module + deeper FM feature parity — incremental |
| Commerce | Parallel **plugin** (`services/commerce`, compose profile); webhook → jobs + pack/ship echo |
| Growth model | Thin Go interfaces + optional compose services — no plugin mega-framework |

## Capability → module map

| Capability | Home |
|------------|------|
| Live fleet / HMS / camera stills | `edge` → Redis → `gateway` SSE → `web` |
| Job archive, artifacts, reprint | `core` + object store profile |
| Queue / scheduling / batch | `core` scheduler |
| Costing + utilization | rate cards + `analytics` |
| Filament / AMS / Spoolman | `integrations` + presets |
| Notifications | `core` + `NOTIFY_URL` (job finish + alerts) |
| Power / automation | `core` + `POWER_URL` (dry-run when unset) |
| Auth (LAN done; OIDC stubs; LDAP later) | `gateway` |
| Shop orders → print → pack/ship | `services/commerce` (profile; webhook + echo) |
| Direct printer (in-process) | `direct_printer` + `DIRECT_PRINTERS` |
| In-app / Orca slice | `services/slicer` (profile; Orca not bundled) |

## Phases

### P0 — Close farm MVP loop

- Live `fm_native` upload + `StartPrint`; job state machine end-to-end
- Fleet pause/resume/stop/bed_clean from UI
- Cost estimate → actual on finish
- Acceptance: [`DESIGN_SPEC.md`](./DESIGN_SPEC.md) §19

### P1 — Archive + queue

- Retain artifacts; search/filter jobs; reprint
- Richer queue (reorder, model targeting, wait-on-previous)

### P2 — Filament depth

- Spoolman consumption on finish; AMS slot mapping; low-stock signals

### P3 — Notify + power plugins

- Webhook / ntfy first; smart-plug interface behind a tiny adapter

### P4 — Commerce plugin

- Enable compose profile `commerce`
- Store webhook → jobs → pack/ship status via our APIs

### P5 — Direct printer adapter

- New `farm.Kind` implementation; same core job APIs

### P6 — Slicer + FM parity depth

- `services/slicer` + Orca; expand FM coverage without rewriting core

## Lean rule

Prefer: extend a service **or** add one profiled service + one interface.  
Reject: shared plugin SDK frameworks, frontend stack rewrites, fat all-in-one images, bundling third-party apps into our images until a dedicated module owns them.

## Session handoff

| Field | Value |
|-------|--------|
| **Done** | FM 3.x version gate (`FARM_MANAGER_MIN_VERSION`, edge status `compatible`); Studio-bypass docs; job reassign + maintenance; Orca→artifact; print-history + typed notify |
| **Next** | Live FM 3.2 LAN contract smoke (checklist in edge-farm-manager); UX ladder; OIDC code exchange |
| **Gaps** | OIDC code exchange stub; CDP stub; unread `EDGE_TOKEN`/`FARM_MODE`; DESIGN_SPEC §11 historical HTMX narrative |
| **Blocked** | Live FM 3.2 smoke until LAN FM available |
| **Deferred** | Capacitor / store shells for iOS & Android; FM network-status API if/when captain exposes it |
| **Do not** | Commit third-party product source trees; migrate off Go; bundle the Orca binary into the image; replace FM Client for adopt/unbind |
