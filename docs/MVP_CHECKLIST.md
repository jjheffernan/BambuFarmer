# MVP freeze checklist & integration questions

Answer the questions below before we leave design and implement Farm Manager / Spoolman / Manyfold wiring for real. Scaffolding can proceed with the **assumed defaults** marked in italics.

---

## A. Must-answer (blocks real FM integration)

### A1. Farm Manager transport
How will `bambufarmer-edge` talk to Farm Manager Server?

- [ ] **CDP** — FM Client on the edge host with `--remote-debugging-port=9222` (community-proven; fastest bring-up)
- [ ] **Direct mTLS** — we have (or can export/issue) client certs for a service account
- [ ] **Unknown / need help deciding**

*Assumed default until answered: CDP for Phase 1, design direct as follow-up.*

**Also tell us:**
- FM Server version (e.g. `03.01.02.00`)?
- FM Server LAN IP / hostname?
- Is the edge host **Windows** (same box as FM Client) or **Linux** reaching a Windows FM Client?

### A2. Site topology for MVP
- [ ] **Single site / single org** appliance (one edge, one FM server)
- [ ] Multi-site from day one

*Assumed default: single site.*

### A3. Spoolman
- [ ] Spoolman already running — base URL: `________________`
- [ ] Bundle Spoolman in compose for MVP
- [ ] Stub only in MVP (real sync in Phase 3)

*Assumed default: compose optional profile + stub adapter if URL unset.*

### A4. Manyfold
- [ ] Manyfold already running — base URL: `________________`
- [ ] Bundle / document external Manyfold
- [ ] Stub only in MVP

*Assumed default: stub if URL unset; OAuth client_credentials when configured.*

### A5. Presets source of truth (filament + print)
Where do install-specific presets live today, and how should we import them?

- [ ] OrcaSlicer / Bambu Studio exported JSON profiles
- [ ] Hand-maintained spreadsheet / YAML
- [ ] Only Spoolman filament defs (print presets separate)
- [ ] Start empty in BambuFarmer UI; import later

**Formats you can provide:** ________________

*Assumed default: Postgres-backed preset store + JSON import/export; empty on first boot; UI CRUD in MVP.*

### A6. Auth for first real deploy
- [ ] `local_insecure` on LAN only (lab)
- [ ] OIDC ready with issuer: `________________`

*Assumed default: `local_insecure` for MVP bring-up.*

### A7. Power / circuit limits
- [ ] Required in MVP (tags → max concurrent per circuit)
- [ ] Phase 2 OK

*Assumed default: Phase 2; simple global concurrency env for MVP optional.*

---

## B. Should-answer (shapes UX / costing)

### B1. Currency & costing
- Currency: `____` (e.g. USD)
- Do you have $/kg filament and $/hour machine rates already? Y/N
- Labor overhead in MVP? Y/N

### B2. Printer naming / groups
- Prefer FM tags, FM v3 groups, or our own groups as primary?

### B3. Deploy shape
- [x] Separate containers per module (gateway, core, edge, web, …) — **required**
- [ ] ~~Combine web+gateway~~ — rejected; keep modules split

*Locked: one module → one image → one container (Podman-first).*

### B4. SSE vs WebSocket
*Assumed default: SSE + HTMX (already in design).* Confirm if you need browser→server push beyond forms.

### B5. Object storage
- [ ] MinIO in compose (default)
- [ ] Existing S3-compatible endpoint: `________________`

---

## C. MVP acceptance (design-complete when all boxes checked)

### Product
- [ ] Decisions A1–A7 answered (or explicit “use assumed default”)
- [ ] Presets model accepted (filament + print; per-install; import/export)
- [ ] MVP scope in `DESIGN_SPEC.md` §19 accepted
- [ ] Out-of-MVP list accepted (slicer, Fleet Hub, multi-site, firmware UI)

### Scaffold (engineering ready)
- [x] Go workspace + per-service `Containerfile` (small images, not one fat image)
- [x] Podman Compose (`deploy/compose.yml`) brings up postgres, redis, stub services with `/healthz`
- [x] Quadlet stubs under `deploy/podman/`
- [x] Shared `packages/events` + `packages/api` (+ config, httpserver)
- [x] `deploy/env.example` documents all config knobs
- [x] README explains Podman-first run path (Docker fallback) and local Go
- [x] Example preset bundle + core import/export stub

### Integration readiness (before coding FM adapter for real)
- [ ] FM reachable from edge host (network test)
- [ ] Chosen transport proven once (manual CDP or mTLS smoke)
- [ ] Spoolman/Manyfold URLs or stub mode confirmed
- [ ] At least one sample `.gcode.3mf` and one preset JSON for fixtures

---

## D. Open design items still in `DESIGN_SPEC.md` §23

Track answers here when resolved:

| # | Topic | Answer | Date |
|---|-------|--------|------|
| 1 | FM transport default | | |
| 2 | Multi-tenancy | | |
| 3 | SSE vs WebSocket | SSE (default) | |
| 4 | web/gateway split | | |
| 5 | Presets import format | | |
| 6 | Spoolman/Manyfold mode | | |
