# Changelog

All notable changes to BambuFarmer are documented here.  
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).  
Versioning follows [SemVer](https://semver.org/).

When you land a user-visible or ops-visible change, add a bullet under **[Unreleased]** in the same PR/commit.  
On release: move that section to a new `## [X.Y.Z] — YYYY-MM-DD`, bump [`VERSION`](VERSION), tag `vX.Y.Z`.

## [Unreleased]

### Added
- Shared `bfFlash` toast for copy/control feedback; Jobs row primary actions + More menu; `bf-btn-danger` + Stop confirm
- Fleet `/?sn=` deep-link; Artifacts `?artifact=`; Jobs honor `?q=`/`?state=`; job detail Fleet/Artifacts links; drawer Copy SN; commerce Open job; Jobs Copy link

### Changed
- Removed unused htmx CDN from the web shell; focus-visible rings on buttons/chips
- Jobs deep-link `?job=` opens Detail (optional `history=1`); human-readable job timestamps; Fleet alert Open → drawer
- Filament Copy spool id / tray lookup copy; Admin setup readout includes `rate_card`
- Shell Change password for LAN users (including viewers); Jobs live % from `job.progress`; job detail started/finished + artifact download
- Fleet offline / layer / print_error on cards; Commerce webhook Copy job id; Artifacts Copy object_key
- Admin install readout (`GET /api/v1/setup`) beside setup reset
- Fleet card control errors via bfJSON, HMS chips, and dismissible alert banner; shell shows site name
- Jobs Copy id; Slice `slice.*` SSE toasts; OIDC SSO button on `/login` when `AUTH_MODE=oidc`; Filament low-stock row highlight
- Jobs Cancel prompts for a reason (defaults to operator cancel)
- Jobs table shows queue `#` and target/after; create form model picker; Admin transition job picker with current state
- Analytics utilization table/charts honor `group_by` (printer/site/day); Fleet `gcode_state` filter
- Admin Lab control to re-open the setup wizard (`POST /api/v1/setup/reset`)
- Commerce webhook Spoolman spool picker; Slice advanced Orca `printer_json` / `process_json` / `filament_json` paths
- Fleet edge status strip, online/model filters, and bulk action accepted/skipped feedback
- Jobs create form optional `queue_pos` for initial queue ordering
- Filament and Library pages show Spoolman/Manyfold URL status from `GET /api/v1/integrations/status`
- Analytics export kind picker (`utilization` / `costs` / `failures` / `all`) on JSON and CSV download links
- Artifacts page Detail panel loads `GET /api/v1/artifacts/{id}` including `object_key`
- Jobs Dispatch prompts for a printer SN when the job is unassigned
- Presets page Duplicate forks a tip under a new id/name via GET + POST
- Admin Scheduler readout (`GET /api/v1/admin/scheduler`) shows `CIRCUIT_MAX` / stagger / max starts per tick
- Admin Job transition form (`POST /api/v1/jobs/{id}/transition`) for fixture/lab recovery of stuck jobs
- Jobs page Detail panel loads `GET /api/v1/jobs/{id}` (queue, spool, after-job, costs, full JSON)
- Jobs create form pickers for printer SN, after-job, Spoolman spool, and staged artifact
- Slice page print/filament preset dropdowns from `GET /api/v1/presets/*`
- Library export and Commerce webhook forms pick staged artifact ids from the artifact store
- Admin Farm tags panel lists circuit/group tags from `GET /api/v1/edge/tags`
- Slice page staged-artifact picker fills `input_path` (and name) from `GET /api/v1/artifacts`
- Admin Users: change role and Enable/Disable via existing `PATCH /api/v1/admin/users/{id}`
- Jobs page Dispatch button for queued/assigned jobs (`POST /api/v1/jobs/{id}/dispatch`) so operators can start without waiting on the scheduler tick
- Site name/timezone edit (`PATCH /api/v1/sites/{id}`) and Admin Site form for the farm display name
- Admin notify status + Send test (`GET /api/v1/notify`, `POST /api/v1/notify/test`) for `NOTIFY_URL` wiring; stack chip includes notify
- Presets page create/edit forms (POST/PATCH filament and print) with Edit on each list card
- Delete staged artifacts (`DELETE /api/v1/artifacts/{id}`) from API and Artifacts page; Admin stack shows edge connection status
- Commerce page: plugin status, test store webhook, and pack/ship status echo when the commerce profile is up
- Slice page: plugin status and submit form for `POST /api/v1/slice/jobs` when the slicer profile is up
- Artifacts page: list, upload, download, and copy id for staged plates
- Sites API (`GET /api/v1/sites`, `GET /api/v1/sites/{id}`); Admin stack panel shows the site name
- Admin stack/plugin status chips (version, Spoolman, Manyfold, power, slicer, commerce) and build version in the shell footer
- Admin power On/Off controls and `GET /api/v1/power` status when `POWER_URL` is set (dry-run when unset)
- Scheduler start stagger (`SCHEDULER_STAGGER`, optional `SCHEDULER_MAX_STARTS_PER_TICK`) so auto-assign does not slam every idle printer at once
- Filament and Library operator pages (Spoolman spools/low-stock/tray lookup; Manyfold list/import/export)
- Preset GET/PATCH/DELETE by id for filament and print; Presets page Delete removes tip and revision history
- Fleet printer filters: `GET /api/v1/printers` accepts `tag`, `state`, `model`, `q`, `online`; Fleet page tag chips + search
- Job transition history (`GET /api/v1/jobs/{id}/events`); Jobs page History panel
- Cancel a job (`POST /api/v1/jobs/{id}/cancel`); Jobs page Cancel button; stops the printer when the job was already on a machine
- Optional MinIO/S3 artifact backend (`OBJECT_STORE_*`): core mirrors uploads, edge fetches on local miss; compose `--profile object` adds MinIO + bucket init
- Analytics date window (`from`/`to`) and utilization `group_by` (`printer`/`site`/`day`) on series + export; Analytics page Apply controls
- Fleet drawer liveview JPEG refreshes every 3s while open (cache-busted fetch); stops on close
- Preset revision history and restore (`GET …/presets/{filament|print}/{id}/revisions`, `POST …/restore`); Presets page History opens the list
- `GET /api/v1/printers` returns full snapshots hydrated from `last_snapshot` (AMS, HMS, task, temps) so fleet bootstrap matches edge status
- Core forwards `alert.raised` to `NOTIFY_URL` when set (same webhook as job-finished; empty URL does nothing)
- Edge raises `alert.raised` on `bf.events.alert` for offline, PAUSE/FAILED, or new HMS codes; fleet shows a live attention banner; analytics counts the stream
- Presets page import/export (download JSON bundle, paste or upload file into `POST /api/v1/presets/import`) with list refresh
- Admin rate card editor (currency, $/hr, $/kg, kg/hr, labor flat, failure surcharge) replaces the session/rates JSON dump
- Edge publishes `printer.ams_changed` when AMS trays change (fleet SSE listens); first empty AMS sighting is skipped
- Fleet printer detail drawer (temps, AMS table, task, HMS, tags, liveview still) opened from the card title; updates live over SSE
- Fleet cards show AMS tray chips and active task name from edge snapshots; `GET /api/v1/edge/status` returns full snapshots (not summary-only)
- Printer snapshots carry AMS trays and active task metadata from MQTT/FM `push_status` (`ams`, `task_id`, `subtask_name`) on deltas and the printers API
- Slice jobs accept `print_preset_id` / `filament_preset_id` and materialize core preset JSON into Orca `--load-settings` / `--load-filaments` when `CORE_URL` and `ORCA_BIN` are set
- Slice lifecycle events (`slice.queued` / `slice.succeeded` / `slice.failed`) on `bf.events.slice` when the slicer has `REDIS_URL`; gateway SSE and analytics counters include the stream
- Slice plugin invokes operator-supplied `ORCA_BIN` when set (`input_path` required; optional settings/filament JSON paths); empty `ORCA_BIN` still records the job without running a slicer
- Printer `hms[]` from MQTT/FM push_status is formatted as `XXXX_XXXX_XXXX_XXXX` codes on snapshots and printer events; analytics failures roll up each HMS code alongside `print_error`

### Fixed
- Preset edit PATCH merges the loaded tip so JSON/models/notes are not wiped by the form fields
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
