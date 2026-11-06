# MVP freeze checklist & integration decisions

Decisions below are **locked** from operator answers (2026-11-06). Remaining boxes are engineering smoke tests before FM adapter coding.

---

## A. Locked product decisions

### A1. Farm Manager transport — **mirror Farm Manager Client**

- [x] Speak FM the way the official Client does: local ticket → JWT, FM security headers, REST `:8888`, MQTT reports when available
- [x] Primary transport goal: **FM-native** (direct session/API mirroring Client behavior / mTLS where required)
- [x] CDP remains an optional escape hatch that still drives the same Client-side API calls

Still needed before coding against a live box (not design blockers):

- FM Server version: `________________`
- FM Server LAN IP / hostname: `________________`
- Edge host OS (Windows with Client vs Linux): `________________`

### A2. Site topology — **single site**

- [x] Single site / single org (one edge, one FM server)
- [x] Multi-site deferred

### A3. Spoolman — **live URL, optional**

- [x] Configure a live Spoolman URL when you have one
- [x] **Not required** to run BambuFarmer MVP (stub / disabled if unset)
- [x] Do not bundle Spoolman as a hard dependency

### A4. Manyfold — **live URL, OAuth optional**

- [x] Configure a live Manyfold URL when you have one
- [x] OAuth client_credentials when provided; allow simpler token/no-OAuth paths when Manyfold setup allows
- [x] **Not required** to run MVP if unset

### A5. Presets — **simple + version control**

- [x] Keep presets simple (CRUD + JSON import/export)
- [x] **Version control** on filament + print presets (revision history / bump on change)
- [x] No requirement to import Orca/Studio packs for MVP

### A6. Auth — **LAN mode (insecure) + OIDC-ready**

- [x] **`AUTH_MODE=lan`** (LAN mode): local users, no public URL / IdP required — insecure by design for private networks
- [x] OIDC remains available for URL’d / secured deploys (`AUTH_MODE=oidc`)
- [x] Alias: `local_insecure` accepted as synonym of `lan`
- [x] Differences documented in [`DEPLOYMENT_MODES.md`](./DEPLOYMENT_MODES.md); secrets via [`.env.example`](../.env.example) → `.env`

### A7. Power / circuit limits

- [x] Phase 2 OK (not MVP-blocking)

### Deploy shape

- [x] Single site
- [x] **web + gateway stay separate containers**
- [x] Podman-first; one module → one image → one container

### Costing

- [x] **$/hr** (machine time)
- [x] **$/kg** (filament)
- [x] **kg/hr** (consumption / throughput rate — stored or derived)

---

## B. Still open (non-blocking)

### B2. Printer naming / groups
Prefer FM tags, FM v3 groups, or our own groups as primary? *(default: mirror FM tags/groups)*

### B4. SSE vs WebSocket
Default remains **SSE + HTMX**.

### B5. Object storage
MinIO optional (`--profile object`) until a live S3 endpoint is set.

---

## C. MVP acceptance

### Product
- [x] Decisions A1–A7 answered
- [x] Presets: simple + versioned
- [x] Costing: $/hr, $/kg, kg/hr
- [x] Auth: LAN mode + OIDC-ready
- [x] Spoolman/Manyfold optional live URLs
- [ ] Operator confirms DESIGN_SPEC §19 MVP acceptance criteria still OK

### Scaffold (engineering ready)
- [x] Go workspace + per-service Containerfiles
- [x] Podman Compose + Quadlets
- [x] Shared packages
- [x] `.env.example` → `.env` + README + [`DEPLOYMENT_MODES.md`](./DEPLOYMENT_MODES.md)
- [x] Example preset bundle
- [x] Dependabot + CI
- [x] Redis Streams bus (`packages/events`) + gateway SSE
- [x] Edge fixture mode (no live FM) + delta publish
- [x] Jobs + rate cards APIs; analytics consumer/rollups
- [x] HTMX fleet shell + sample `.gcode.3mf` fixture
- [x] First-run install wizard (`/setup`) + tagged releases workflow

### Integration readiness (before live FM adapter)
- [ ] FM reachable from edge host
- [ ] FM-native (or CDP) smoke once
- [ ] Spoolman/Manyfold URLs filled or left empty intentionally
- [x] Sample `.gcode.3mf` fixture (`fixtures/artifacts/sample.gcode.3mf`)

---

## D. Decision log

| # | Topic | Answer | Date |
|---|-------|--------|------|
| 1 | FM transport | Mirror FM Client (FM-native primary; CDP escape hatch) | 2026-11-06 |
| 2 | Multi-tenancy | Single site MVP | 2026-11-06 |
| 3 | Spoolman | Live URL optional; not required to run | 2026-11-06 |
| 4 | Manyfold | Live URL; OAuth optional | 2026-11-06 |
| 5 | Presets | Simple + version control | 2026-11-06 |
| 6 | web/gateway | Separate containers | 2026-11-06 |
| 7 | Costing | $/hr, $/kg, kg/hr | 2026-11-06 |
| 8 | Auth | `AUTH_MODE=lan` insecure LAN mode + OIDC | 2026-11-06 |
