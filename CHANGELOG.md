# Changelog

All notable changes to BambuFarmer are documented here.  
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).  
Versioning follows [SemVer](https://semver.org/).

When you land a user-visible or ops-visible change, add a bullet under **[Unreleased]** in the same PR/commit.  
On release: move that section to a new `## [X.Y.Z] — YYYY-MM-DD`, bump [`VERSION`](VERSION), tag `vX.Y.Z`.

## [Unreleased]

### Added
- Slice plugin invokes operator-supplied `ORCA_BIN` when set (`input_path` required; optional settings/filament JSON paths); empty `ORCA_BIN` still records the job without running a slicer
- Printer `hms[]` from MQTT/FM push_status is formatted as `XXXX_XXXX_XXXX_XXXX` codes on snapshots and printer events; analytics failures roll up each HMS code alongside `print_error`
- Analytics export (`GET /api/v1/analytics/export`) returns utilization, costs, and failures as JSON or CSV (`format`, `kind`); Analytics page links download both
- Analytics page charts for daily utilization, costs, and failure codes (CSS bars + tables from existing APIs)
- Analytics costs (`GET /api/v1/analytics/costs`) roll up `job.costed` material, machine, and labor into `agg_cost_daily`
- Analytics failures (`GET /api/v1/analytics/failures`) roll up non-zero `print_error` codes into `agg_hms_daily`
- Farm tags (`GET /api/v1/edge/tags`) attach circuit/group labels onto printer snapshots; scheduler respects `CIRCUIT_MAX` when set
- Printer liveview JPEG (`GET /api/v1/printers/{id}/liveview`) via edge; Farm Manager uses `GET /file?asset_path=liveview\{sn}\liveview.jpeg`, fixture and direct return a tiny still offline
- Edge subscribes to `device/+/report` when `FARM_MQTT_URL` and cert files in `EDGE_CERT_DIR` are both set. Polling continues either way
- Farm report payloads (`device/<sn>/report`) fold into the existing printer snapshot. `msg` 0 replaces telemetry and `msg` 1 updates only fields that are present
- Direct printer control plane (`FARM_MANAGER_TRANSPORT=direct_printer`) lists printers from `DIRECT_PRINTERS` and runs upload, start, and pause in process; it does not open a network connection
- Slice plugin (`compose --profile slicer`, `POST /api/v1/slice/jobs`) records a slice job; empty `ORCA_BIN` does not invoke a slicer, and the Orca binary is not bundled
- Library export (`POST /api/v1/library/export`) creates a Manyfold model from an artifact id and name when `MANYFOLD_URL` is set; an empty URL does not call Manyfold
- Library import (`POST /api/v1/library/import`) creates a Manyfold model by name when `MANYFOLD_URL` is set; an empty URL does not call Manyfold
- Commerce pack/ship status (`POST /api/v1/commerce/orders/status`) echoes `packed` or `shipped` for an order id; nothing is stored
- Power switch (`POST /api/v1/power`) forwards `on` or `off` when `POWER_URL` is set; an empty URL does nothing
- AMS tray lookup (`GET /api/v1/filament/links?tray=`) resolves a spool id from `AMS_SLOT_MAP`; an empty map resolves nothing
- Spool list includes `low_stock` when remaining grams are under `LOW_STOCK_GRAMS` (default 100); Spoolman unset reports none
- Commerce store webhook (`POST /api/v1/commerce/webhook`) queues a core job when `CORE_URL` is set; the service stays on the commerce compose profile
- Job finish can POST a JSON webhook when `NOTIFY_URL` is set; an empty URL does nothing
- Finished jobs post filament grams to Spoolman when a spool id is set and `SPOOLMAN_URL` is configured; an empty URL does nothing
- Job queue: earlier/later reorder, optional printer model target, and wait-until-previous-job-succeeds
- Jobs list search and state filter (`GET /api/v1/jobs?q=&state=`) over retained name, id, printer SN, and artifact
- Reprint a retained plate as a new queued job (`POST /api/v1/jobs/{id}/reprint`) from the jobs page, without re-uploading
- Edge publishes printer deltas immediately after StartPrint/Control (no poll wait; fixture/SSE friendly)
- Fleet bulk control: multi-select + sticky action bar; `POST /api/v1/printers/actions` fans out pause/resume/stop/bed_clean via edge command stream
- Local artifact staging: `POST/GET /api/v1/artifacts` (+ `/content`), shared compose `artifacts` volume (`ARTIFACTS_DIR`)
- Edge uploads staged plate to FM before `StartPrint` when `file_id` is absent (`object_key` / `artifact_id`)
- Jobs page: `.gcode.3mf` file upload + optional FM `file_id`; create form (name, artifact, printer SN, copies, estimates) + list with state/cost/error and live SSE refresh
- Core job dispatch: scheduler assigns idle printers, publishes `command.start_print`, consumes `job.*` to advance state + finalize cost
- Edge `fm_native` `Upload3MF` / `StartPrint` (FM `POST /file/upload3mf` + `POST /task`); command consumer emits `job.dispatched` / `job.progress` / `job.finished` / `job.failed`
- Build-out roadmap + lean plugin docs; optional `commerce` compose profile (health stub)
- Reserved `farm.KindDirect` for a future LAN printer adapter behind `ControlPlane`
- LAN local accounts: Postgres `users`, Argon2id passwords, HMAC `bf_session` cookies
- Setup wizard creates first admin (password on Finish); `/login` + admin user CRUD
- Password change (self) and admin password reset APIs + Admin UI
- Stack updates: `./scripts/bf update` / `make update`, `--pull`, Admin “Request update” + `--watch`
- Operator CLI `./scripts/bf` (`setup`, `update`, `doctor`, `uninstall`, `help`)
- Tailwind CSS toolchain for web (`make web-css`, embedded `/static/css/app.css`)
- Shared API errors (`packages/api.Error`) with JSON `{error,code}` via `httpserver.WriteError`
- Postgres baseline migration + lab seed / `make db-reset`
- Spoolman and Manyfold list clients (optional live URLs)
- Module-local `tests/` packages and FORCE_TEST_FAIL canary
- GitHub issue/PR templates; CI changelog shape check

### Changed
- Fixture control plane: `StartPrint`/`Control` advance per-SN gcode_state overlays (RUNNING/PAUSE/IDLE) without a live FM — LAN-free lab path
- Analytics page: daily utilization table + event counter chips instead of raw JSON dump
- Jobs page: live list refresh via gateway SSE (`job.*` events) instead of a 10s poll
- Go toolchain raised to 1.26 (CI, Containerfiles, modules) for current dependency minimums
- Dependencies: `pgx/v5` 5.12.0, `go-redis/v9` 9.23.0, `golang.org/x/crypto` 0.58.0
- Schema/seed defined in Go (`packages/db`); `*.sql` gitignored (no committed SQL dumps)
- Restyled fleet UI and install wizard with shared `bf-*` Tailwind components
- Relocated layout: `core/internal`, `gateway/auth`, `web/setup`, `edge`/`integrations` `app/`
- Safer web JSON fetches (`bfJSON`) so non-JSON API errors are readable
- After install, gateway requires login for UI/API (except setup/login/health)

### Fixed
- Gateway image build: COPY `packages/db` and `packages/version` into the Containerfile context (auth/`version` imports)
- `command.start_print` no longer aliases our `artifact_id` as FM `file_id`/`f3mf_id` (fails closed until upload or explicit `file_id`)
- CI compose config check seeds `.env` from `.env.example`
- Analytics `/api/v1/analytics/counters` served from current image (was 404 → browser JSON parse error)

## [0.1.0] — 2026-11-06

### Added
- Initial modular Go services: gateway, core, edge, integrations, analytics, web
- Podman-first compose, `.env.example`, `AUTH_MODE=lan` + OIDC stubs
- Farm Manager `fm_native` adapter scaffold + printer state machine
- Versioned filament/print presets in Postgres
- First-run install wizard (`/setup`) with site, auth, costing, FM, and integrations steps
- Redis Streams bus, gateway SSE, edge fixture mode, jobs/rate cards, analytics consumer
- Release workflow (tagged `v*` → GitHub Release + module images)
- Deployment modes doc (LAN vs OIDC)

[Unreleased]: https://github.com/jjheffernan/BambuFarmer/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/jjheffernan/BambuFarmer/releases/tag/v0.1.0
