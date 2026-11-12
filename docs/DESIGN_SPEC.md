# BambuFarmer — Design Specification

**Status:** Draft v0.8 (MVP decisions locked)  
**Date:** 2026-11-06  
**Scope:** Architecture and module design for a containerized print-farm operations platform that integrates with the Bambu Lab ecosystem (Farm Manager Server primary; Fleet Hub optional later), Spoolman (and peers), Manyfold (and peers), with multi-user access, costing, analytics, and **per-install filament/print presets**. Post-MVP: OrcaSlicer-backed slice service.

**Pre-build gate:** integration answers are recorded in [`MVP_CHECKLIST.md`](./MVP_CHECKLIST.md). Remaining items are live-box smoke tests (FM IP/version), not open design.

**Per-service design shapes:** [`docs/services/`](./services/README.md) — including Farm Manager research for edge ([`edge-farm-manager.md`](./services/edge-farm-manager.md)).

### Locked decisions

| Topic | Decision |
|-------|----------|
| Bambu control plane | **Farm Manager Server** — mirror Client behavior (ticket/JWT, headers, REST, MQTT) |
| FM transport | **`fm_native` primary**; CDP optional escape hatch |
| Site topology | **Single site** MVP |
| Language (all services) | **Go** |
| Event bus | **Redis Streams** |
| Auth | **`AUTH_MODE=lan`** (insecure LAN mode) + **OIDC-ready** for URL’d deploys |
| web-frontend | **Go + templ + HTMX** (separate container from gateway) |
| farm-edge | **Go** |
| Spoolman | **Optional live URL** — not required to run |
| Manyfold | **Optional live URL**; OAuth optional |
| Presets | Simple CRUD + **version history**; JSON import/export; per-install |
| Costing | **$/hr**, **$/kg**, and **kg/hr** |
| Slicing | **Out of MVP**; OrcaSlicer later |
| Analytics | Dedicated **`bambufarmer-analytics`** |
| Runtime | **Podman-first**; Docker fallback |
| Packaging | **One module → one image → one container** |

---

## 1. Vision

BambuFarmer is an operations layer for running a Bambu Lab print farm. It sits *beside* Bambu’s official farm stack—not as a replacement for printer binding/control—and adds:

- Central multi-user farm operations (jobs, queues, permissions)
- Event-driven fleet telemetry and alerting
- Filament inventory sync (Spoolman-first, pluggable)
- 3D library import/export (Manyfold-first, pluggable)
- Costing (material + machine time + overhead)
- Analytics for utilization, failure rates, cost per part

Bambu Farm Manager Server remains the trusted LAN control plane for printers in our deployment. BambuFarmer orchestrates workflows, inventory, files, money, and people around that plane. A Fleet Hub adapter may be added later for sites that adopt Hub hardware; it is not required for v1.

---

## 2. Goals & Non-Goals

### 2.1 Goals

| ID | Goal |
|----|------|
| G1 | Easily web-hosted; works behind reverse proxies / Traefik / Caddy / nginx |
| G2 | Fully containerized; **Podman-first**, also Docker / Proxmox / k3s; one module → one image → one container |
| G3 | 12-factor compliant (config via env, stateless app processes, backing services) |
| G4 | Pluggable filament managers (Spoolman primary) |
| G5 | Pluggable file libraries (Manyfold primary) |
| G6 | UI mirrors Bambu Lab visual language (Farm Manager / Studio cues) |
| G7 | Multi-user with roles |
| G8 | Cost pricing and analytics |
| G9 | Event-driven (not constant polling of printers) |
| G10 | Proper job management (queue, scheduling, circuit/power awareness, retry) |
| G11 | Prefer Rust / Go / Python / C++; avoid Java and TypeScript-heavy stacks |

### 2.2 Non-Goals (MVP)

- Replacing Bambu Farm Manager Server or Fleet Hub as the device bind/control authority
- Cloud-only Bambu Handy control paths for farm-bound printers
- **In-app slicing** (MVP ingests pre-sliced `.gcode.3mf` from Bambu Studio / OrcaSlicer desktop; a dedicated slicer module is planned post-MVP)
- Full MES/ERP replacement (expose APIs so others can integrate)

---

## 3. Constraints & Design Principles

1. **Farm Manager first.** Our control plane is Bambu Farm Manager Server (LAN REST :8888 + farm MQTT). Isolate the FM adapter behind `FarmControlPlane` so a Fleet Hub adapter can land later without rewriting core.
2. **Edge owns LAN secrets.** mTLS certs and Farm Manager JWTs stay on the farm-edge host—never in the public web tier.
3. **Events over polling.** Prefer farm MQTT report stream when reachable; otherwise short-interval REST delta poll on edge only. Never poll from the browser.
4. **Adapters, not hardcodes.** Spoolman/Manyfold are first adapters behind stable internal ports.
5. **Jobs are first-class.** Every print is a Job with state machine, audit trail, and cost accrual.
6. **12-factor.** Config = env; processes are disposable; logs to stdout; backing services (Postgres, Redis, object store) are attached, not embedded as durable state in the app image.
7. **Modular containers (hard rule).** Each major module is its own image and container (distroless, single binary). No fat “all-in-one” app image. Prefer **Podman** (compose or Quadlets).
8. **Auth modes.** Default farm bring-up is **`AUTH_MODE=lan`** (insecure LAN mode: local users, no public URL/IdP). Use **`AUTH_MODE=oidc`** when the stack is exposed via a real URL. Never put LAN mode on the public internet.

---

## 4. High-Level Architecture

```
┌──────────────────────────────────────────────────────────────────────────┐
│                         Operators (browser)                              │
│                              │                                           │
│                     ┌────────▼────────┐                                  │
│                     │  web-frontend   │  (Bambu-like GUI)                │
│                     └────────┬────────┘                                  │
│                              │ HTTPS                                     │
│                     ┌────────▼────────┐                                  │
│                     │   api-gateway   │  auth, rate limit, OpenAPI       │
│                     └────────┬────────┘                                  │
│                              │                                           │
│              ┌───────────────┼────────────────┬──────────────────┐       │
│              │               │                │                  │       │
│     ┌────────▼──────┐ ┌──────▼──────┐ ┌───────▼────────┐ ┌───────▼─────┐ │
│     │ core-backend  │ │  analytics  │ │  integrations  │ │ slicer*     │ │
│     │ (jobs, users, │ │  (dedicated │ │  (Spoolman,    │ │ (OrcaSlicer │ │
│     │  costing,     │ │   rollups)  │ │   Manyfold)    │ │  post-MVP)  │ │
│     │  inventory)   │ │             │ │                │ │             │ │
│     └───────┬───────┘ └──────▲──────┘ └───────▲────────┘ └───────▲─────┘ │
│             │                │ events         │                  │       │
│             │         ┌──────┴────────────────┴──────────────────┘       │
│             │         │ event-bus (Redis Streams)                        │
│             │         └──────▲───────────────────────────────────────────│
│             │                │                                           │
│             │                │                                           │
│     ┌───────▼────────────────┴───────────────┐                           │
│     │         farm-edge (Go)                 │  (one per site / LAN)     │
│     │  Farm Manager adapter (v1)             │                           │
│     │  MQTT/event normalizer · file cache    │                           │
│     │  [Fleet Hub adapter — future]         │                           │
│     └───────────────────┬────────────────────┘                           │
│                         │                                                │
│              ┌──────────▼───────────┐                                    │
│              │ Farm Manager Server  │                                    │
│              │ (LAN, mTLS :8888,    │                                    │
│              │  MQTT broker :1883)  │                                    │
│              └──────────┬───────────┘                                    │
│                         ▼                                                │
│                   Bambu printers                                         │
└──────────────────────────────────────────────────────────────────────────┘

External (same Podman/Compose network or linked hosts):
  Spoolman · Manyfold · Postgres · Redis · Object storage (S3/MinIO) · optional OIDC
```

### 4.1 Module Map

| Module | Container | Responsibility |
|--------|-----------|----------------|
| **web-frontend** | `bambufarmer-web` | **Go + templ + HTMX.** Operator GUI; Bambu-like design language |
| **api-gateway** | `bambufarmer-gateway` | TLS termination (optional), JWT/OIDC, routing, OpenAPI surface |
| **core-backend** | `bambufarmer-core` | Users, orgs, printers registry, jobs, costing rules, policies |
| **farm-edge** | `bambufarmer-edge` | **Go.** LAN integration with Farm Manager (v1); event ingest; file push |
| **integrations** | `bambufarmer-integrations` | Spoolman, Manyfold, other adapters; sync workers |
| **analytics** | `bambufarmer-analytics` | **Dedicated.** Event → rollups, dashboards API, exports (separate from core) |
| **slicer** | `bambufarmer-slicer` | **Post-MVP.** OrcaSlicer CLI wrapper; profile/job queue → `.gcode.3mf` |
| **event-bus** | Redis Streams | Domain events; same Redis may also hold sessions/cache |
| **db** | Postgres | System of record |
| **object-store** | MinIO / S3 | `.gcode.3mf`, snapshots, exports |

Minimum viable deploy: `web` + `gateway` + `core` + `edge` + `analytics` + Postgres + Redis (+ Spoolman/Manyfold as external deps).  
`slicer` is optional and **not** in MVP compose.

---

## 5. Bambu Integration Strategy

### 5.1 Primary (our setup): Farm Manager Server

[Bambu Farm Manager](https://wiki.bambulab.com/en/software/bambu-farm-manager) = Server (LAN control) + Client (Electron UI). **This is our production control plane.**

Community protocol notes ([OpenBambuFarmAPI](https://github.com/ATILLLLITA/OpenBambuFarmAPI), [GeraldCox API Reference](https://github.com/GeraldCox/bambu-farm-manager-script/blob/main/API_Reference.md)):

- Server REST on **:8888** with mTLS (TLS 1.2/1.3)
- Printers are MQTT *clients* to the farm broker (**:1883**/TLS) after TCP-**3002** bind
- Client often holds non-exportable mTLS material; community tools use CDP proxy through the Electron client, or negotiate certs where available
- Surfaces we care about: devices/telemetry, tasks, files (`.gcode.3mf`), tags (incl. circuit tags), users, liveview, settings
- Circuit-aware concurrency (tags → max printers per circuit) is a first-class scheduler input

**farm-edge (Go)** implements Farm Manager as the v1 adapter:

1. Store FM server URL, site credentials, and TLS materials in edge secrets volume / env
2. Auth: login ticket → challenge-hash → JWT (`Authorization: Bearer …` + Bambu security headers)
3. Prefer **MQTT `device/<sn>/report`** (or equivalent farm report topics) for state changes
4. If MQTT access is blocked, fall back to REST `GET /devices` on a short interval and emit events only on **delta**
5. Normalize all payloads into internal `PrinterSnapshot` / `PrinterEvent` types before publishing to Redis Streams
6. REST for task create, file upload, pause/resume/stop, tags as needed
7. Pin tested Farm Manager Server versions in docs; treat the API as **interop / best-effort** (undocumented by Bambu)

**Transport (locked: mirror Farm Manager Client):**

| Mode | Env | Behavior |
|------|-----|----------|
| **FM-native (default)** | `FARM_MANAGER_TRANSPORT=fm_native` | Same flows as FM Client: local login tickets → JWT, required Bambu security headers, REST to `:8888`, prefer MQTT device reports for events. Use mTLS/client identity as the Client does when the server requires it. |
| CDP escape hatch | `FARM_MANAGER_TRANSPORT=cdp` | Drive a running FM Client via DevTools (`CDP_URL`); still issues the same Client API calls — useful when cert material is locked in Electron. |

Do not invent a parallel protocol; stay aligned with observed FM Client ↔ Server behavior ([OpenBambuFarmAPI](https://github.com/ATILLLLITA/OpenBambuFarmAPI), community API notes).

### 5.2 Future (optional): Fleet Hub

[Fleet Hub](https://wiki.bambulab.com/en/acc/manual/fleet-hub) remains a **future adapter** for sites that buy Hub hardware (documented HTTP API, ≤50 printers/hub, mTLS + activation). Same `FarmControlPlane` trait; no core changes required. Not in Phase 1 scope.

### 5.3 Adapter interface (edge)

Implemented in Go as [`packages/farm.ControlPlane`](../packages/farm/plane.go). Backends: `fm_native`, `fixture`, `cdp` (stub), later `fleet_hub`. The edge **agent** (`services/edge/app/agent`) is the only runtime that calls a plane.

```text
farm.ControlPlane {
  Connect / Ready / Captain
  ListDevices() -> []DeviceReport
  Control(sn, Pause|Resume|Stop|BedClean)
  Upload3MF / StartPrint / LiveviewJPEG
}
// optional: farm.EventSource.Subscribe() for MQTT push
```

Multiple edges can register with core (multi-site farms). Each edge has a `site_id`.

---

## 6. Event-Driven Design

### 6.1 Principle

Do **not** have the web UI or core poll printers every N seconds. Instead:

1. **farm-edge** observes Farm Manager MQTT and/or REST deltas
2. Emits normalized events onto **Redis Streams**
3. **core** updates projections and job state
4. **web** receives updates via SSE/WebSocket from gateway (fan-out of stream topics)

Polling is allowed only:

- As edge fallback when MQTT is unavailable
- For external systems without webhooks (Spoolman has websockets—use them)
- For health checks (liveness ≠ full telemetry)

### 6.2 Core event catalog (initial)

| Event | Producer | Consumers |
|-------|----------|-----------|
| `printer.online` / `printer.offline` | edge | core, analytics, web |
| `printer.state_changed` | edge | core, web |
| `printer.ams_changed` | edge | integrations (filament), web |
| `job.queued` | core | edge, web |
| `job.dispatched` | edge | core, web |
| `job.progress` | edge | core, analytics, web |
| `job.finished` / `job.failed` | edge | core, costing, filament, analytics |
| `job.cancelled` | core/edge | all |
| `filament.consumed` | integrations | costing, analytics |
| `library.file_imported` | integrations | core, web |
| `alert.raised` | core | web, optional notify |

### 6.3 Event envelope

```json
{
  "id": "uuid",
  "type": "job.finished",
  "ts": "2026-11-06T17:00:00Z",
  "site_id": "shop-a",
  "source": "farm-edge",
  "correlation_id": "job-uuid",
  "payload": { }
}
```

Idempotency keys on `id` / `(source, external_ref)` so reconnects do not double-count filament or cost.

---

## 7. Job Management System

Jobs are the heart of the product. A **Job** is an operator intent; an **ExternalTask** is what Farm Manager actually runs.

### 7.1 Job state machine

```
                ┌────────────┐
                │  drafted   │
                └─────┬──────┘
                      │ submit
                ┌─────▼──────┐
         ┌──────│   queued   │◄────────── requeue (policy)
         │      └─────┬──────┘
         │            │ scheduler assigns printer(s)
         │      ┌─────▼──────┐
         │      │  assigned  │
         │      └─────┬──────┘
         │            │ edge accepted / file staged
         │      ┌─────▼──────┐
 cancel  │      │ dispatched │
─────────┼──────┴─────┬──────┘
         │            │ printer RUNNING
         │      ┌─────▼──────┐
         │      │  printing  │──pause──► paused ──resume──┐
         │      └─────┬──────┘◄───────────────────────────┘
         │            │
         │     ┌──────┴──────┐
         │     ▼             ▼
         │  succeeded     failed
         │     │             │
         │     └──────┬──────┘
         │            ▼
         │      ┌───────────┐
         └─────►│  closed   │  (cost finalized, filament posted)
                └───────────┘
```

### 7.2 Job fields (logical)

- Identity: `job_id`, `org_id`, `created_by`, timestamps
- Artifact: library file ref and/or uploaded `.gcode.3mf` object key, plate path, MD5
- Targeting: printer tags / model / nozzle diameter / AMS requirements / site
- Scheduling: priority, not-before, deadline, circuit/power group, max concurrency
- Mapping: filament slot map (ams_id/slot_id), material expectations
- Economics: estimated cost, actual cost (material + machine + labor overhead)
- External: `site_id`, `external_task_id`, Farm Manager file/task refs
- Audit: state transitions with actor + reason

### 7.3 Scheduler responsibilities (core)

1. Select idle-compatible printers (model, nozzle, filament availability)
2. Respect **power/circuit limits** (inspired by circuit-aware Farm Manager tooling: tags → max concurrent per circuit)
3. Stagger starts if configured (Farm Manager “staggered print initiation” analogue)
4. Emit `job.assigned` → edge performs upload/print
5. On failure: retry policy (same printer / failover / human hold)

### 7.4 Batch & multi-copy

- One Job can expand into N copies (subtasks) across M printers
- Track per-copy status; parent job succeeds when policy says (all / quorum)
- Deduplicate file upload via content hash when FM/server already has the slice

---

## 8. Filament & File Integrations

### 8.1 Filament port (Spoolman-first)

Internal interface:

```text
FilamentProvider {
  list_spools(filter) -> Vec<Spool>
  get_spool(id) -> Spool
  use_filament(spool_id, grams) -> Result   // post-print
  reserve(spool_id, grams) -> Reservation   // optional pre-print
  find_match(material, color, brand) -> Vec<Spool>
}
```

**Spoolman adapter:** optional live `SPOOLMAN_URL`. REST `/api/v1/*` + websockets when configured.  
Map AMS tray RFID/`tag_uid` + color/type to Spoolman spools when possible. If URL unset → disabled/stub (MVP still runs).

Other providers (future): Spoolman-compatible forks, spreadsheet CSV, Homelab-specific DBs—same port.

On `job.finished`, integrations computes grams used (from slicer metadata if present, else estimate from length × density) and posts consumption.

### 8.2 Library port (Manyfold-first)

```text
LibraryProvider {
  search(query) -> Vec<ModelSummary>
  get_model(id) -> ModelDetail
  download_file(model_id, file_id) -> Stream
  upload_model(meta, files) -> ModelRef
  ensure_collection(path) -> CollectionRef
}
```

**Manyfold adapter:** optional live `MANYFOLD_URL`. Prefer OAuth2 client_credentials when client id/secret are set; allow token or unauthenticated read paths when the Manyfold instance permits. TUS upload then `POST /models` when write/upload is configured. JSON-LD `application/vnd.manyfold.v0+json`. If URL unset → disabled/stub (MVP still runs).

Flows:

- **Import to farm:** Manyfold model → download printable / attach existing `.gcode.3mf` from object store → create Job
- **Export from farm:** Successful print artifacts / archived slices → Manyfold model in a “Farm Productions” collection
- **Sync:** Optional bidirectional metadata (tags, creator) without duplicating binary blobs unnecessarily

Other libraries (Printables local mirrors, custom S3 libraries) implement the same port.

---

## 8.3 Slicer module (post-MVP — out of MVP scope)

MVP accepts **pre-sliced** `.gcode.3mf` (Bambu Studio / OrcaSlicer desktop / library). A dedicated **`bambufarmer-slicer`** service is planned afterward.

**Dependency:** [OrcaSlicer](https://github.com/OrcaSlicer/OrcaSlicer) as a containerized CLI binary (not embedded as a library). Headless pattern:

```text
orca-slicer model.3mf \
  --load-settings "printer.json;process.json" \
  --load-filaments filament.json \
  --arrange 1 --slice 0 \
  --export-3mf out.gcode.3mf
```

**Module shape**

| Piece | Role |
|-------|------|
| `slicer` service | Job queue, profile store refs, invoke OrcaSlicer, write artifacts to object store |
| Profiles | Printer/process/filament JSON managed in UI or synced from a profile pack |
| Output | `.gcode.3mf` + metadata (time estimate, filament grams) → core Job / costing |
| Events | `slice.queued` / `slice.succeeded` / `slice.failed` on Redis Streams |

**Why separate container:** OrcaSlicer images are large and CPU-heavy; isolate from edge/core so slice spikes do not starve farm control. Optional scale-to-zero on small farms.

**Non-goals for slicer v1:** full interactive plate editor in-browser (defer; start with upload model → pick profile → slice).

### 8.4 Presets (filament + print) — per install

Presets **must not** be hardcoded in the container image. Each farm install maintains its own catalog because materials, nozzles, and process preferences differ site-to-site.

Owned by **core** (CRUD + import/export). UI under Admin / Presets. Optional sync hints toward Spoolman filament types and (post-MVP) OrcaSlicer profile JSON.

#### FilamentPreset

| Field | Purpose |
|-------|---------|
| `id`, `name`, `version` | Identity; bump version on edit or keep history rows |
| `material` | PLA, PETG, ABS, … |
| `vendor` / `product_line` | Optional |
| `color_hex` | Display / AMS matching hint |
| `density_g_cm3`, `diameter_mm` | Mass estimates |
| `nozzle_temp_min` / `max`, `bed_temp` | Process bounds |
| `tray_info_idx` | Bambu filament id hint (e.g. `GFA00`) for FM AMS mapping |
| `price_per_kg` | Costing override (else Spoolman / rate card) |
| `spoolman_filament_id` | Optional link |
| `notes`, `active` | Ops |

#### PrintPreset

| Field | Purpose |
|-------|---------|
| `id`, `name`, `version` | Identity |
| `compatible_models[]` | e.g. P1S, X1C |
| `nozzle_diameter_mm` | 0.4 default |
| `layer_height_mm`, `infill_pct`, `sparse_infill_pattern` | Metadata for MVP tagging |
| `process_json` | Opaque JSON blob (Orca/Bambu process export) — stored, not interpreted in MVP |
| `printer_json` | Opaque printer profile blob (optional) |
| `default_for_model` | Optional unique default per model |
| `est_speed_factor` | Optional costing/ETA fudge |
| `active` | Soft delete |

#### Behaviors (MVP)

1. **CRUD** in web UI + REST `/api/v1/presets/filament`, `/api/v1/presets/print`
2. **Version control (simple):** each save creates a new immutable revision (`version`++), keeps history; UI can view/restore prior revisions; export includes versions
3. **Import/export** JSON bundle: `{ "filament_presets": [...], "print_presets": [...] }`
4. Job composer can attach `filament_preset_id` (+ optional `version`) / `print_preset_id`; does not require slicer
5. Artifacts may record which preset revision was intended when uploaded
6. When slicer lands (Phase 6), `process_json` / `printer_json` feed Orca `--load-settings` / `--load-filaments`

Presets are **site/org-scoped** (single-site MVP = one catalog). No image rebuild to change them. Keep the model simple — no external VCS dependency; versioning lives in Postgres.

---

## 9. Costing & Analytics

### 9.1 Costing model

Primary rates (locked):

| Rate | Unit | Use |
|------|------|-----|
| Machine | **$/hr** | `hours × machine_rate_per_hour` |
| Filament | **$/kg** | `kg × price_per_kg` (from preset, Spoolman, or rate card) |
| Throughput | **kg/hr** | Expected or measured filament burn rate; also `kg = kg/hr × hours` for estimates |

Per job:

```
material_cost     = kg_used × ($/kg)
machine_time_cost = hours × ($/hr)
kg_hr_implied     = kg_used / hours          # analytics / validation
estimate_kg       = (kg/hr) × hours          # when grams unknown pre-print
cost              = material_cost + machine_time_cost + labor_overhead + failure_amortization
```

| Component | Inputs |
|-----------|--------|
| Material | kg × **$/kg** (preset `price_per_kg`, Spoolman, or org default) |
| Machine time | hours × **$/hr** (by model, tag, or org default) |
| Throughput | **kg/hr** on rate card and/or derived from job actuals |
| Labor | optional flat per-job or per-hour |
| Failure | optional surcharge pool / insurance rate |

Pricing tables and quote/actual calculation live in **core**. Analytics *reads* finalized costs and rolls up $/hr, $/kg, and kg/hr.

### 9.2 Dedicated analytics module

**`bambufarmer-analytics` is a first-class standalone service**, not a package inside core. It exists so heavy rollups, long-range history, and dashboard queries never contend with job dispatch OLTP.

**Responsibilities**

| Owns | Does not own |
|------|----------------|
| Event consumption from Redis Streams | Job state machine |
| Materialized rollups / warehouse tables | Rate card CRUD |
| Dashboard & report APIs (`/api/v1/analytics/...`) | Live printer control |
| CSV/JSON/PDF exports | Auth IdP (uses gateway tokens) |

**Ingest:** subscribe to `printer.*`, `job.*`, `filament.*`, `alert.*` (and optional CDC from Postgres for backfill). Idempotent upserts by event `id`.

**Rollups (v1)**

- Utilization % by printer / group / tag / site / day
- Success vs fail rates; top HMS / error codes
- Avg job duration vs estimate; queue wait time; throughput
- Cost per part / per gram / per hour (from core cost finals)
- Filament burn-down vs Spoolman remaining
- Group-level stats (aligns with Farm Manager v3 printer grouping)

**Serving:** REST aggregates for HTMX analytics screens (and JSON for charts); optional Prometheus metrics on all services for ops (separate from business analytics).

**Scale note:** analytics can run its own read replica or DuckDB/export sidecar later without changing core.

---

## 10. Multi-User & Security

### 10.1 Identity

Auth is **LAN mode by default** for farms that do not want a public URL or IdP, and **OIDC-ready** for secured/URL’d installs.

**Full comparison (auth + network placement + secrets):** [`DEPLOYMENT_MODES.md`](./DEPLOYMENT_MODES.md).

| Mode | Env | Behavior |
|------|-----|----------|
| **`lan`** | `AUTH_MODE=lan` | **LAN mode (insecure).** Local username/password (Argon2id). No OIDC, no public URL required. Bind to LAN/VPN only. Loud boot warning. Default for MVP compose. |
| `oidc` | `AUTH_MODE=oidc` | OIDC (Authentik, Keycloak, Google Workspace, …). Use when the stack has a real URL / reverse proxy. |

Alias: `AUTH_MODE=local_insecure` is accepted as a synonym of `lan`.

Implementation notes:

- Same role model in both modes; only the identity provider changes
- Session/JWT issuance stays in gateway so swapping modes does not rewrite core
- OIDC code paths are stubbed/validated whenever `AUTH_MODE=oidc`
- Single-site MVP; multi-site later
- Edge / Farm Manager secrets stay LAN-local in **both** modes (see §10.3)

### 10.2 Roles (initial)

| Role | Capabilities |
|------|----------------|
| `viewer` | Read printers, jobs, analytics |
| `operator` | Create/cancel jobs, pause/resume, basic filament ops |
| `librarian` | Manyfold import/export, file management |
| `inventory` | Spoolman sync, pricing tables |
| `admin` | Users, sites, adapters, power limits, billing rules |

Farm Manager local admin credentials are **service accounts** held by edge—not end-user logins.

### 10.3 Secrets & network

- Edge must sit on printer LAN (or routable to Farm Manager Server)
- Core/web can be LAN-only or reverse-proxied with auth
- mTLS / CDP materials: edge-only (`EDGE_CERT_DIR` or sidecar FM Client); never mount into web
- All inter-service traffic on internal Podman/Compose network; gateway is the only public ingress by default
- When `AUTH_MODE=lan`, bind gateway to LAN / VPN only (never public ingress)

---

## 11. Frontend & Design Language

### 11.1 Goals (mirror Farm Manager Client)

Farm Manager Client is a dense, real-time Electron ops console ([features](https://wiki.bambulab.com/en/software/bambu-farm-features)). We mirror that *interaction model*, not pixel-perfect assets:

| Farm Manager trait | Our UI implication |
|--------------------|--------------------|
| Printers overview with status nav (Ready to Start / Attention Needed / Completed / Not in Task) | Live fleet grid + filter chips driven by event stream |
| Detail side-panel (temps, SN, nozzle, bed) | Drawer / split pane, not full-page navigation |
| Bulk operations (pause, cancel, resume, LED, speed, temps, unload, calibrate) | Multi-select + sticky action bar |
| Task queue (copies, printer set, green Start when ready, edit, terminate) | Job board with optimistic local state |
| Two-level `.gcode.3mf` folders + compatibility metadata | Artifact browser with model/nozzle/bed/ETA columns |
| Customizable printer list columns (v3) | Client-side column prefs persisted per user |
| Liveview / RTSPS (server-relayed) | Detail pane media; JPEG snapshots first, stream later |
| Groups (v3) + tags | Group/tag filters for dispatch and analytics |
| Multi-user local accounts | Our roles + OIDC / LAN mode |
| High update rate across many printers | Fine-grained reactive UI; avoid full-page HTML swaps |

### 11.2 Frontend decision (locked: Go + HTMX)

We previously compared **Go + templ + HTMX** vs **Rust + Leptos**. Leptos scored higher on dense live fleet UX; HTMX scored higher on ship speed, CI simplicity, and one-language ops.

**v0.5 decision: unify on Go** — edge and web both move to Go so the whole system shares one toolchain. HTMX is Plan A. Leptos is retired as a contingency (not maintained as Plan B).

| Factor | Outcome |
|--------|---------|
| Toolchain / CI | Single Go module graph; no WASM |
| Shared DTOs | One `packages/` Go module for events + API types across edge, core, web |
| Fleet live updates | SSE → HTMX `sse` / OOB swaps of card rows + status chips; throttle morphs |
| Multi-select bulk ops | Alpine (or small vanilla JS) for selection set; actions POST to gateway |
| Liveview | `<img>` poll/SSE refresh of JPEG; video deferred |
| Risk accepted | At very large fleets, full-table morphs may need virtualized table or coarser update intervals |

#### Historical comparison (kept for context)

| Workload | HTMX | Leptos | Notes |
|----------|------|--------|-------|
| Dense per-cell updates | SSE partial swaps | Signal store | Leptos smoother; HTMX OK with row-level targets |
| Bulk select + optimistic UI | Alpine assist | Native | Accept small Alpine surface |
| Admin CRUD | Excellent | Heavier | Favors HTMX |
| Shared types with Go edge | Native Go | Would need dual stack | Favors HTMX after Go edge lock |
| Compile on small VMs | Light | WASM toolchain | Favors HTMX |

### 11.3 Go + HTMX stack (Plan A)

- **Server:** Go (`net/http` or chi) + [templ](https://templ.guide/) components
- **Interactivity:** HTMX for navigation, forms, SSE-driven fragment updates; Alpine.js only for selection/checkbox state and dropdowns
- **Auth:** session cookie from gateway (same-origin or trusted proxy); web may be combined with gateway in small deploys
- **Data:** initial HTML from server; live fleet/jobs via `GET /api/v1/events/stream` (SSE) triggering HTMX swaps of `#printer-{id}`, `#status-nav`, job rows
- **Shared types:** Go packages under `packages/events`, `packages/api` imported by edge, core, gateway, web
- **Styling:** CSS variables; dark industrial Farm-inspired theme; original icons/assets only
- **No browser polling of Farm Manager** — only SSE from our gateway

### 11.4 Visual design tokens (initial)

```css
:root {
  --bf-bg-0: #0e1114;
  --bf-bg-1: #161b20;
  --bf-bg-2: #1e252c;
  --bf-text: #e8eef2;
  --bf-muted: #8b9aab;
  --bf-accent: #00ae42;      /* ready / success — Bambu-green cue */
  --bf-warn: #f0a202;        /* attention */
  --bf-danger: #e23d28;      /* failed / offline hard */
  --bf-info: #3aa0ff;        /* running / info */
  --bf-border: #2a343d;
  --bf-radius: 6px;
  --bf-font-ui: "IBM Plex Sans", "Segoe UI", sans-serif;
  --bf-font-mono: "IBM Plex Mono", ui-monospace, monospace;
}
```

Status color mapping: `IDLE` muted · `RUNNING` info · `PAUSE`/`FAILED` warn/danger · `FINISH` accent (awaiting bed_clean).

### 11.5 Primary screens

1. Fleet overview (status nav + grid/table + bulk bar + detail drawer)  
2. Job queue / board  
3. Job composer (artifact + filament map + targets + cost estimate)  
4. Library browser (Manyfold)  
5. Filament inventory (Spoolman)  
6. Analytics (backed by `bambufarmer-analytics`)  
7. Admin (sites, adapters, users, rates)  
8. Presets (filament + print; import/export)  
9. *(Post-MVP)* Slice jobs (OrcaSlicer using stored preset JSON)

---

## 12. Technology Recommendations

| Layer | Choice | Rationale |
|-------|--------|-----------|
| api-gateway | **Go** (chi/echo) | Small static binary; OIDC + local auth; OpenAPI |
| core-backend | **Go** | CRUD, jobs, SQL—fast to ship; excellent Postgres story |
| farm-edge | **Go** | mTLS/MQTT/Redis; same language as the rest of the stack |
| integrations | **Go** | Spoolman/Manyfold HTTP glue |
| analytics | **Go** | Dedicated service; SQL rollups; optional DuckDB later |
| web-frontend | **Go + templ + HTMX** | Unified Go; SSE fragment updates; tiny Alpine only where needed |
| slicer | **Go** + **OrcaSlicer** binary | Post-MVP sidecar; CLI wrapper only |
| DB | **Postgres 16+** | System of record |
| Bus / cache | **Redis** (Streams + optional session/cache) | One small container; easy on Podman/Proxmox |
| Objects | **MinIO** / S3 | Slice files & snapshots |

**Why Redis over NATS for us:** fewer moving parts at small scale, ops familiarity, and one service covers Streams (events) plus optional gateway sessions/rate-limit keys. Revisit NATS only if we outgrow Redis Streams semantics.

**Language policy:** Go for all first-party services. Python only if an adapter ecosystem forces it. No TypeScript. No Rust in MVP (may revisit later for a hot path only if measured need).

C++ reserved for any future native protocol helpers if required; not needed for MVP HTTP/MQTT adapters.

---

## 13. 12-Factor & Deployment

### 13.1 Config (env examples)

Secrets and config live in a **repo-root `.env`** (gitignored). Template: [`.env.example`](../.env.example). Modes: [`DEPLOYMENT_MODES.md`](./DEPLOYMENT_MODES.md).

```bash
cp .env.example .env   # then edit secrets

# gateway / auth
AUTH_MODE=lan                     # insecure LAN mode (default); use oidc for URL’d deploys
SESSION_SECRET=...

# edge — mirror FM Client (printer LAN in every mode)
FARM_MANAGER_TRANSPORT=fm_native  # or: cdp
FARM_MANAGER_URL=http://192.168.1.20:8888
FARM_MANAGER_PASSWORD=...

# costing
CURRENCY=USD
MACHINE_RATE_PER_HOUR=...         # $/hr
FILAMENT_PRICE_PER_KG=...         # $/kg
FILAMENT_KG_PER_HOUR=...          # kg/hr

# integrations (optional; empty = disabled)
SPOOLMAN_URL=
MANYFOLD_URL=
```

Compose: `make up` → `--env-file .env -f deploy/compose.yml` (creates `.env` from the example if missing).

### 13.2 Compose topology (illustrative)

```yaml
# deploy/compose.yml — each build uses services/<name>/Containerfile
services:
  gateway:       # localhost/bambufarmer-gateway
  core:          # localhost/bambufarmer-core
  edge:          # localhost/bambufarmer-edge (LAN / FM)
  integrations:  # localhost/bambufarmer-integrations
  analytics:     # localhost/bambufarmer-analytics
  web:           # localhost/bambufarmer-web
  postgres:
  redis:
  # minio:       # profile "object"
  # slicer:      # post-MVP, separate heavy image
```

### 13.3 Host targets

| Host | Notes |
|------|-------|
| **Podman Compose** | **Preferred** developer + small farm (`make up`) |
| **Podman Quadlets** | Preferred systemd/rootless production (`deploy/podman/`) |
| Docker Compose | Compatible fallback (`ENGINE=docker`) |
| Proxmox | CT/VM with Podman; edge on printer VLAN |
| K8s / k3s | Later; same per-module images, Ingress instead of compose ports |

Edge may need host networking or a VLAN interface for SSDP discovery—document as a deployment profile `edge-lan`.

### 13.4 Observability

- Structured JSON logs to stdout  
- `/healthz` + `/readyz` on each service  
- OpenTelemetry traces optional from day one stubs  
- Metrics: job throughput, edge lag, adapter errors  

---

## 14. Data Model

Owned primarily by **core** (OLTP). Analytics maintains its own rollup tables. Projections of printer live state may be cached in Redis; durable history is Postgres + events.

### 14.1 Entity sketch

```text
Organization
  └── Site (id, name, timezone, network_notes)
        └── Printer (id, site_id, external_sn, model, name, group_id?, tags[],
                     nozzle_diameter, last_seen_at, last_snapshot jsonb)
        └── PrinterGroup (id, site_id, name)          # FM v3 grouping analogue
        └── PowerCircuit (id, site_id, name, max_concurrent)
              └── circuit ↔ printer membership
  └── User (id, subject, display_name, auth_provider)
  └── RoleBinding (user_id, role, scope_org|site)
  └── RateCard (currency, machine_rate_per_hour $/hr, filament_price_per_kg $/kg,
                filament_kg_per_hour kg/hr, labor_flat, failure_surcharge_pct)
  └── Artifact (id, object_key, md5, filename, plate_path, compatible_models[],
                nozzle_mm, bed_type, est_seconds, est_grams, source_library_ref?)
  └── Job (see §7)
        └── JobEvent (at, from_state, to_state, actor, reason)
        └── JobCost (estimate_*, actual_*, currency, finalized_at)
  └── FilamentLink (printer_id, ams_id, slot_id, provider, spool_id)
  └── FilamentPreset (per-install; see §8.4)
  └── PrintPreset (per-install; see §8.4)
  └── AdapterConfig (kind, site_id, secret_ref, config jsonb)
```

### 14.2 Printer snapshot (normalized, from edge)

```json
{
  "site_id": "shop-a",
  "external_sn": "01P00C…",
  "model": "P1S",
  "name": "Bay-3",
  "online": true,
  "gcode_state": "RUNNING",
  "mc_percent": 42,
  "mc_remaining_time": 3600,
  "temps": { "nozzle": 220.1, "bed": 60.0, "chamber": null },
  "ams": [{ "id": 0, "trays": [{ "slot": 0, "type": "PLA", "color": "00AE42FF", "tag_uid": "…" }] }],
  "task": { "external_task_id": "…", "subtask_name": "…" },
  "hms": [],
  "observed_at": "2026-11-06T17:00:00Z"
}
```

Farm Manager remains device authority; core stores last-known + history for UX/analytics when FM is briefly unreachable.

### 14.3 Analytics tables (owned by analytics service)

- `fact_job_completed` (job_id, site_id, printer_id, duration_s, success, cost_actual, grams, …)
- `fact_printer_state_sample` (optional downsampled) or derive utilization from state-change intervals
- `agg_utilization_daily` (printer_id|group_id|site_id, day, idle_s, run_s, fail_s)
- `agg_hms_daily` (code, count)
- `agg_cost_daily` (dimensions…, material, machine, labor)

---

## 15. Redis Streams Topology

| Stream key | Producers | Consumers | Purpose |
|------------|-----------|-----------|---------|
| `bf.events.printer` | edge | core, analytics, gateway (SSE) | online/state/ams/hms |
| `bf.events.job` | core, edge | core, analytics, integrations, gateway | lifecycle |
| `bf.events.filament` | integrations | analytics, core | consumption / link changes |
| `bf.events.alert` | core, edge | gateway, analytics | operator attention |
| `bf.events.slice` | slicer (post-MVP) | core, analytics | slice job outcomes |
| `bf.commands.edge.{site_id}` | core | edge | start_print, control, upload |

Consumer groups: `core`, `analytics`, `integrations`, `gateway-fanout`.  
Retention: time- or maxlen-bounded (e.g. 7 days raw events); analytics rollups are durable in Postgres.

Command stream ack: edge writes `job.dispatched` / `job.failed` on the job stream; core uses idempotent state transitions.

---

## 16. farm-edge Internals (Go)

```text
┌─────────────────────────────────────────────────────┐
│ bambufarmer-edge (Go)                               │
│  ┌─────────────┐   ┌──────────────┐   ┌──────────┐  │
│  │ FM transport│──►│ normalizer   │──►│ redis    │  │
│  │ direct|cdp  │   │ PrinterEvent │   │ publish  │  │
│  └─────────────┘   └──────────────┘   └──────────┘  │
│  ┌─────────────┐   ┌──────────────┐                 │
│  │ MQTT sub or │──►│ delta filter │                 │
│  │ REST poller │   │ (hash state) │                 │
│  └─────────────┘   └──────────────┘                 │
│  ┌──────────────────────────────────────────────┐   │
│  │ command worker ← bf.commands.edge.{site}     │   │
│  │  upload_slice / start_print / pause|…        │   │
│  └──────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────┘
```

Suggested libs: `eclipse/paho.mqtt.golang` or `mqtt` client; stdlib or `crypto/tls` for mTLS; `go-redis` for Streams; chromedp only if `FARM_MANAGER_TRANSPORT=cdp`.

**Delta filter:** keep last snapshot hash per `external_sn`; publish only on change (or heartbeat every N minutes for liveness).  
**Backpressure:** if Redis is down, buffer bounded in-memory then drop-oldest with metric alarm—never block printer MQTT forever.  
**Health:** `/healthz` = process up; `/readyz` = FM session valid + Redis writable.

---

## 17. API Surface (gateway)

Versioned public API: `/api/v1/...`. OpenAPI 3 is the contract; web may call JSON endpoints or prefer server-side gateway/core clients and only expose HTML.

### 17.1 Auth

| Method | Path | Notes |
|--------|------|-------|
| POST | `/api/v1/auth/login` | `AUTH_MODE=lan` (and alias `local_insecure`) |
| GET | `/api/v1/auth/oidc/login` | redirect |
| GET | `/api/v1/auth/oidc/callback` | code exchange |
| POST | `/api/v1/auth/logout` | |
| GET | `/api/v1/auth/me` | roles + org/site |

### 17.2 Fleet & jobs (MVP core)

| Method | Path | Notes |
|--------|------|-------|
| GET | `/api/v1/sites` | |
| GET | `/api/v1/printers` | filters: site, group, tag, state |
| GET | `/api/v1/printers/{id}` | |
| GET | `/api/v1/printers/{id}/liveview` | JPEG proxy via edge |
| POST | `/api/v1/printers/actions` | bulk: pause/resume/stop/bed_clean/led/… |
| GET | `/api/v1/artifacts` | |
| GET | `/api/v1/jobs/{id}` | |
| POST | `/api/v1/jobs` | create + optional auto-queue |
| POST | `/api/v1/jobs/{id}/queue` | |
| POST | `/api/v1/jobs/{id}/cancel` | |
| POST | `/api/v1/jobs/{id}/reprint` | |
| POST | `/api/v1/artifacts` | multipart `.gcode.3mf` |
| GET | `/api/v1/artifacts` | folder tree analogue |
| GET | `/api/v1/events/stream` | SSE (printer/job/alert) |

### 17.3 Integrations & analytics

| Method | Path | Notes |
|--------|------|-------|
| GET | `/api/v1/library/models` | Manyfold proxy |
| POST | `/api/v1/library/import` | → artifact |
| POST | `/api/v1/library/export` | artifact → Manyfold |
| GET | `/api/v1/filament/spools` | Spoolman proxy |
| PUT | `/api/v1/filament/links` | AMS ↔ spool |
| GET | `/api/v1/analytics/utilization` | query params: from, to, group_by |
| GET | `/api/v1/analytics/costs` | |
| GET | `/api/v1/analytics/failures` | |
| GET | `/api/v1/analytics/export` | CSV/JSON |

### 17.4 Presets & admin

| Method | Path | Notes |
|--------|------|-------|
| GET/POST | `/api/v1/presets/filament` | list / create |
| GET/PATCH/DELETE | `/api/v1/presets/filament/{id}` | |
| GET/POST | `/api/v1/presets/print` | list / create |
| GET/PATCH/DELETE | `/api/v1/presets/print/{id}` | |
| POST | `/api/v1/presets/import` | JSON bundle |
| GET | `/api/v1/presets/export` | JSON bundle |
| CRUD | `/api/v1/admin/users` | |
| CRUD | `/api/v1/admin/rate-cards` | |
| CRUD | `/api/v1/admin/sites` | |
| GET/PUT | `/api/v1/admin/adapters/{site}` | non-secret config; secrets via env/refs |
| GET/POST | `/api/v1/setup` | first-run install wizard status / save / complete |
| POST | `/api/v1/setup/reset` | re-open wizard (dev) |
| GET | `/api/v1/version` | build version / commit |
| POST | `/internal/edge/events` | optional HTTP ingest alt to Redis; token auth |
| GET | `/healthz`, `/readyz` | |

Post-MVP slicer: `/api/v1/slice/jobs` (submit/status/download).

---

## 18. Key Flows

### 18.1 Live fleet update

```text
FM MQTT/REST → edge normalizer → bf.events.printer
  → gateway fanout → SSE → HTMX swaps `#printer-{id}` / status nav
  → analytics consumer → utilization intervals
  → core (optional) → refresh printers.last_snapshot
```

### 18.2 Print job (happy path)

```text
Operator (HTMX form) → POST /jobs (artifact_id, targets, filament_map)
  → core: Job queued → bf.events.job
  → scheduler: assign printer (idle + model + nozzle + circuit cap)
  → bf.commands.edge.{site}: upload if needed + start_print
  → edge → FM REST task API
  → edge: job.dispatched / job.progress / job.finished
  → core: state → succeeded → JobCost actual
  → integrations: filament.consumed (Spoolman)
  → analytics: fact_job_completed
```

### 18.3 Attention Needed

Derived in UI (and optionally core alerts) when:

- `gcode_state` in `PAUSE` | `FAILED`
- `online == false` while expected
- `FINISH` awaiting bed_clean / confirm removal
- HMS codes above configured severity

Maps to FM’s “Attention Needed” / “Completed Task” nav buckets.

---

## 19. MVP Acceptance Criteria

MVP is done when a single-site farm can:

1. Run compose stack (`gateway`, `core`, `edge`, `analytics`, `web`, Postgres, Redis) with **`AUTH_MODE=lan`**
2. Edge authenticates to Farm Manager via **`fm_native`** (or CDP escape hatch) and publishes printer deltas to Redis
3. HTMX fleet page shows live states via SSE without browser polling FM
4. Operator uploads `.gcode.3mf`, creates a job, prints on an idle compatible printer
5. Pause / resume / stop / bed_clean work from UI for a selected printer
6. Job completes → cost estimate/actual with **$/hr**, **$/kg**, and **kg/hr**
7. Analytics stores daily utilization + job success/fail + cost/rate rollups (basic UI OK)
8. Spoolman/Manyfold optional — empty URL = disabled; live URL when configured
9. Entire first-party codebase builds with Go only; one image per module
10. Filament + print presets: CRUD, **version history**, JSON import/export
11. Single site; web and gateway remain separate containers

Out of MVP: OrcaSlicer, forcing OIDC, Fleet Hub, multi-site, firmware upgrade UI, live RTSPS video (JPEG OK).

**Design-phase exit:** checklist §A locked (done) + scaffold green; live FM IP/version smoke before adapter coding.

---

## 20. Phased Roadmap

### Phase 0 — Foundations
- Go monorepo / multi-module layout + compose ✅
- Postgres schema, Redis Streams topics, gateway with `AUTH_MODE=lan` + OIDC stubs ✅
- **analytics** service skeleton (stream consumer + rollup tables + health) ✅
- HTMX shell with Farm-like design tokens + SSE fleet ✅ (templ migration optional)

### Phase 1 — Farm Manager edge MVP
- Go farm-edge FM adapter: **fm_native** (mirror Client) or CDP escape hatch; devices; REST-delta → Redis Streams ✅ (fixture mode + live path; MQTT later)
- HTMX fleet UI live updates via SSE ✅ (fixture→Redis→SSE path)
- Manual job: upload `.gcode.3mf` → print on selected idle printers ⏳ (job API + sample 3mf; FM start_print stub)
- Analytics ingest of printer/job events (basic utilization) ✅

### Phase 2 — Job system
- Full state machine, tags/groups, circuit/power concurrency
- Bulk pause/resume/stop/bed_clean from UI (FM-like action bar; Alpine selection)
- Cost estimates + actuals (manual filament prices OK)

### Phase 3 — Integrations + analytics UI
- Spoolman adapter + AMS linking + consumption on finish
- Manyfold import/export
- Analytics dashboards v1 (utilization, fail rates, cost)

### Phase 4 — Multi-site + production auth
- Multiple edge agents / sites
- Document `AUTH_MODE=oidc` for URL’d deploys; richer RBAC; exports/API for MES

### Phase 5 — Hardening (+ optional Fleet Hub)
- Failure failover policies, alerting (webhook/email)
- Optional Fleet Hub adapter behind same trait
- Performance & security review; refuse LAN mode behind public ingress

### Phase 6 — Slicer (post-MVP)
- `bambufarmer-slicer` + OrcaSlicer container
- Profile management; async slice jobs → `.gcode.3mf` into object store / job composer
- Feed slice metadata into costing + analytics

### Plugin tracks (parallel; see [`BUILD_OUT.md`](./BUILD_OUT.md), [`plugins.md`](./plugins.md))
- **Commerce** (`services/commerce`, compose `--profile commerce`): stub ✅; store webhook → jobs → pack/ship later
- **Direct printer** (`farm.KindDirect`): reserved behind `ControlPlane`; implement after FM job loop is solid
- **Notify / power**: thin adapters, not core hard deps
- Growth rule: interfaces + optional services — no plugin mega-framework; stay Go

---

## 21. Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| Farm Manager API unofficial / fragile | Isolate in Go edge adapter; version-pin tested FM server builds; contract tests against captured fixtures |
| mTLS / CDP fragility | Support `direct` and `cdp` transports; edge-only secrets; never leak to browser |
| Polling temptation | Code review rule: UI may not call device list on interval; use Redis→SSE → HTMX swaps |
| HTMX fleet morph cost at scale | Row-level OOB targets; throttle SSE; coarsen updates; virtualize table if needed |
| LAN mode (`AUTH_MODE=lan`) exposed publicly | Boot warning; bind LAN-only; prefer OIDC for any public URL |
| Filament estimate inaccurate | Prefer slicer metadata (post-MVP Orca / Studio); allow manual override; Spoolman reconcile |
| Manyfold API pre-v1 | Adapter isolation; pin tested Manyfold version |
| OrcaSlicer image size / CPU | Separate optional container; queue + concurrency limits; not in MVP |
| Trademark / UI cloning | Inspired-by theme; original assets; no Bambu logos without permission |
| Legal: reverse-engineered FM APIs | Document as interop/best-effort; no Bambu binaries redistributed; optional Fleet Hub later for documented API path |

---

## 22. Repository Layout (proposed)

```text
BambuFarmer/
  docs/
    DESIGN_SPEC.md          ← this document
  deploy/
    compose.yml             # Podman-first (Docker-compatible)
    podman/                 # Quadlets — one .container per module
    env.example
  services/
    gateway/                # Go + Containerfile
    core/                   # Go + Containerfile
    edge/                   # Go + Containerfile
    integrations/           # Go + Containerfile
    analytics/              # Go + Containerfile
    web/                    # Go + templ + HTMX + Containerfile
    slicer/                 # post-MVP — own heavy Containerfile
  packages/
    api/                    # shared request/response types
    events/                 # shared stream envelopes
    openapi/                # OpenAPI specs / generated clients if any
  scripts/
```

Each `services/*` has its own `Containerfile` and image `localhost/bambufarmer-<name>:<version>` (single binary, distroless). Prefer a Go workspace (`go.work`) at repo root. **Never** collapse app modules into one large container.

---

## 23. Remaining Open Decisions

**Locked (see MVP_CHECKLIST):** FM-native mirror Client · single site · optional Spoolman/Manyfold live URLs · simple versioned presets · $/hr+$/kg+kg/hr · `AUTH_MODE=lan` + OIDC · separate web/gateway · Podman-first modular images · SSE default · Go everywhere.

Still needed for **live integration** (not design):

1. FM Server version + LAN IP/hostname + edge host OS  
2. Optional: Spoolman/Manyfold URLs when ready  
3. Prefer FM tags vs v3 groups as primary UI grouping (default: mirror FM)

---

## 24. References

- [Bambu Farm Manager Wiki](https://wiki.bambulab.com/en/software/bambu-farm-manager) — **our control plane**
- [OpenBambuFarmAPI](https://github.com/ATILLLLITA/OpenBambuFarmAPI) — farm protocol research (MQTT, bind, REST)
- [GeraldCox bambu-farm-manager-script API_Reference](https://github.com/GeraldCox/bambu-farm-manager-script/blob/main/API_Reference.md) — FM REST + circuit-aware batching
- [Fleet Hub Product Intro](https://wiki.bambulab.com/en/acc/manual/fleet-hub) — optional future adapter
- [Fleet Hub Developer Center](https://bambulab.com/en/fleet-hub/developer)
- [Bambu Fleet Hub HTTP API v1.0.0](https://portal.bblmw.com/fleet-hub/wb76oog4i8k/files/Bambu%20Fleet%20Hub%20HTTP%20API%20EN%20v1.0.0.pdf)
- [Spoolman](https://github.com/Donkie/Spoolman)
- [Manyfold REST API](https://github.com/manyfold3d/manyfold)
- [Bambu Farm Manager Features](https://wiki.bambulab.com/en/software/bambu-farm-features) — UI behaviors we mirror
- [OrcaSlicer CLI](https://www.orcaslicer.com/wiki/cli/cli_mode) — post-MVP slicer dependency
- [templ](https://templ.guide/) — Go HTML components
- [HTMX](https://htmx.org/) — locked web interactivity model

---

## Appendix A — Why this split?

| Concern | Module | Why separate |
|---------|--------|--------------|
| Public HTTP + auth | gateway | Swap IdP / rate limits without touching jobs |
| Domain logic | core | Stable job/cost source of truth |
| Hostile/fragile LAN protocols | edge | Crash/restart without taking down UI; keeps certs local |
| Third-party APIs | integrations | Spoolman/Manyfold churn isolated |
| Heavy reads / reports | analytics | Dedicated service; won’t lock job OLTP |
| UI | web (templ+HTMX) | Deploy frontend cadence separately |
| CPU-heavy slicing | slicer | Optional; isolate OrcaSlicer from control plane |

All modules are **Go**; the split is about process/trust boundaries **and** deploy units (separate images).

**Do not** combine gateway/web/core/edge/analytics into one container to “save resources.” Scale or omit modules instead. Keep **edge** on the printer LAN trust boundary; keep **analytics** and **slicer** optional/isolated when unused.

---

## Appendix B — Language unification rationale (v0.5)

Edge and web were briefly locked to Rust (edge reliability + Leptos UI). That created two toolchains for little shared-type gain once services were split.

**v0.5:** Go for edge and web as well.

| Kept | Dropped for MVP |
|------|-----------------|
| Module boundaries (edge/core/gateway/…) | Rust edge |
| Event-driven Redis design | Leptos / WASM |
| Farm Manager–mirroring UX goals | Dual Plan A/B frontend tracks |

Revisit Rust later only if a measured hotspot (e.g. MQTT fan-in) demands it—not as a default.
