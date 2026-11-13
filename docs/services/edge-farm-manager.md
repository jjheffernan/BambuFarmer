# Farm Manager research notes (for `bambufarmer-edge`)

**Status:** living research doc — interop / best-effort  
**Audience:** edge implementers validating against a real Farm Manager Server  
**Disclaimer:** Not affiliated with Bambu Lab. Based on public wiki + independent community research ([OpenBambuFarmAPI](https://github.com/ATILLLLITA/OpenBambuFarmAPI), [GeraldCox API_Reference](https://github.com/GeraldCox/bambu-farm-manager-script/blob/main/API_Reference.md)). Confirm every path against **your** server build before relying on it in production.

Our locked approach: **mirror Farm Manager Client** (`FARM_MANAGER_TRANSPORT=fm_native`), with **CDP** as an escape hatch.

---

## 1. What Farm Manager is

Official suite ([wiki](https://wiki.bambulab.com/en/software/bambu-farm-manager)):

| Piece | Role |
|-------|------|
| **Farm Manager Server** | LAN control plane: binds printers, hosts REST + MQTT broker, stores tasks/files |
| **Farm Manager Client** | Electron UI; holds non-exportable mTLS client cert (“cuckoo” proxy upgrades plain `fetch` → HTTPS) |

Printers in farm mode leave Bambu Handy/cloud binding. Server and printers must be IP-reachable on the LAN.

### Official ports ([FAQ](https://wiki.bambulab.com/en/software/bambu-farm-faq-troubleshoot))

| Port | Proto | Role |
|------|-------|------|
| **8888/TCP** | HTTP+HTTPS (dual) | Farm REST (control, files, liveview, login) |
| **1883/TCP** | MQTT/TLS | Farm broker — **printers dial out** as MQTT clients |
| **3002/TCP** | TLS | Bind handshake — **server dials printer** |
| UDP 1900/1991/2021/2022 | Discovery | SSDP / farm discovery |
| 443/TCP | HTTPS | Client/server cloud (activation, firmware catalog) |

---

## 2. Architecture (farm mode)

```text
  FM Client (Electron)                    Printers
         │                                    │
         │ REST (via mTLS / cuckoo)           │ MQTT client → :1883
         ▼                                    ▼
  ┌──────────────────────────────────────────────────┐
  │           Farm Manager Server                    │
  │  REST :8888 (HTTP for client; HTTPS for some     │
  │        printer file fetches)                     │
  │  MQTT broker :1883/TLS (mochi-mqtt)              │
  │  Bind dialer → printer :3002                     │
  └──────────────────────────────────────────────────┘
```

Key inversion vs cloud: **printer is MQTT client**; printer’s own `:8883` broker is closed in farm mode.

OpenBambuFarmAPI notes `:8888` is **dual-protocol** — plain HTTP for many client/liveview paths, TLS (SNI `farm_server.bambulab.com`) for printer job download. Edge usually talks like the Client (HTTP to LAN IP through mTLS tunnel or CDP).

---

## 3. Authentication (Client-shaped)

### Required headers (community-captured Client traffic)

| Header | Value |
|--------|-------|
| `Authorization` | `Bearer <JWT>` |
| `x-bbl-sec-ver` | `1` |
| `x-bbl-sec-dev` | `1` |
| `X-BBL-Client-Name` | `BambuFarmManager` |
| `X-BBL-Client-Type` | `farm-web` |
| `x-bbl-sec-sid` | Server ID (UUID from `/captain` / login) |

### Local login ticket flow

```text
POST /login/local/tickets
  { "user_name": "admin", "password": "<password>" }
→ ticket_id, ticket (challenge + salt)

POST /login/local/tokens
  { "ticket": "<ticket>" }
→ JWT (~7 days, iss BambuFarmServer)
```

Challenge-response password hashing is required for a correct tokens exchange (see community client implementations). Edge must implement the same hash steps as the Client — do not send raw password on the tokens call.

### CDP shortcut (escape hatch)

If Client is running with `--remote-debugging-port=9222`:

1. CDP into the renderer  
2. Read `localStorage.app.servers[0].token` and `.id`  
3. Issue `fetch('http://<fm-ip>:8888/...')` **from the renderer** so cuckoo applies mTLS  

Useful when client certs are locked in Windows CNG and cannot be exported.

---

## 4. REST surfaces edge cares about (MVP)

Pin observed routes; re-verify on your FM Server version (`GET /captain` → `release`).

| Area | Methods (observed) | Edge use |
|------|-------------------|----------|
| Health / identity | `GET /captain` | `server_id`, `release`, power_limit, session keepalive |
| Devices | `GET /devices`, `GET /devices2`, `GET /device/{dev_id}` | Fleet snapshot + normalization |
| Control | `POST /device/{dev_id}/opt` | `pause` / `resume` / `stop` / (bed_clean via MQTT bridge) |
| Tags | `GET /tags` | Circuit / group membership (scheduler) |
| Tasks | `GET /task`, `POST /task`, launch routes | External task create / status |
| Files | `POST /file/upload3mf`, `GET /file/...` | Stage `.gcode.3mf` |
| Liveview | `GET /file?asset_path=liveview\{sn}\liveview.jpeg` | JPEG proxy (MVP) |
| Scan | `GET /scan_printers?prefix=` | Optional discovery UI |

### Device control via `/opt`

Client `POST /device/{serial}/opt` with body like `{ "opt": "pause" }` is bridged to MQTT `device/<sn>/request`. Community-confirmed opts include pause; resume/stop inferred / observed in MQTT command set (`pause` / `resume` / `stop` / `bed_clean`).

### `gcode_state` (normalize to our events)

`IDLE` · `PREPARE` · `RUNNING` · `PAUSE` · `FINISH` · `FAILED`

Map to FM UI buckets:

| FM nav | Derivation |
|--------|------------|
| Ready to Start | IDLE + pending queue work |
| Attention Needed | PAUSE / FAILED / offline |
| Completed Task | FINISH (awaiting bed_clean / confirm removal) |
| Not in Task | IDLE, no pending work |

---

## 5. MQTT (preferred event source)

Topics:

| Topic | Direction |
|-------|-----------|
| `device/<sn>/request` | server → printer (commands) |
| `device/<sn>/report` | printer → server (telemetry, mostly `push_status`) |

Telemetry field set aligns with OpenBambuAPI `push_status` (`gcode_state`, `mc_percent`, temps, AMS, …). `msg: 0` = full snapshot, `msg: 1` = delta.

**Edge strategy:**

1. Prefer MQTT report subscription **if** we can authenticate to the farm broker as a privileged observer (research: may require farm TLS material — validate on lab server).  
2. Fallback: poll `GET /devices` every 2–5s, emit Redis events only on **delta** (hash snapshot).  
3. Never let the browser poll FM.

### Starting a print (managed task)

Farm start is not “one REST call equals motion.” Observed Client/server path:

1. Upload 3MF (`POST /file/upload3mf`)  
2. Create/launch task via REST (`POST /task` and/or launch routes)  
3. Server publishes MQTT `print.project_file` with `hpart://file/f3mf/{id}/f3mffile`  
4. Printer downloads plate from `:8888` and runs  

Managed `project_file` carries `task_id` / `subtask_id` and farm URL scheme — required for printer “Collected / Reprint / Cancel” behavior. See OpenBambuFarmAPI `farm-mqtt.md` / `farm-tasks.md`.

GeraldCox docs also capture a rich `POST /task` body used by the Client for multi-printer launch — treat as **version-sensitive**; capture yours from Client DevTools when implementing.

---

## 6. Edge package shape (implementation plan)

```text
services/edge/
  main.go                 # process, health, status, control
  app/
    config/               # env: SITE_ID, FARM_MANAGER_*, REDIS_URL, CDP_URL
    adapter/              # Connect + REST poll + Status (M1)
    fm/
      client.go           # HTTP client + BBL headers + JWT
      auth.go             # tickets → tokens (+ HashPasswordChallenge helper)
      devices.go          # list / get
      control.go          # POST /device/{sn}/opt
      # tasks.go / files.go / transport/cdp — planned
    tracker/              # per-SN packages/printerstate.Machine
    normalize/            # FM JSON → packages/api.PrinterSnapshot
    delta/                # snapshot hash filter
    mqtt/                 # report JSON → DeviceReport (broker dial still later)
    bus/                  # Redis Streams publish + command consumer (planned)
packages/printerstate/    # IDLE/PREPARE/RUNNING/PAUSE/FINISH/FAILED + attention buckets
```

### Go interfaces

Canonical types live in [`packages/farm`](../../packages/farm):

```go
// farm.ControlPlane — implemented by edge/app/bambu/{native,fixture,cdp,none}
type ControlPlane interface {
    Kind() Kind
    Connect(ctx context.Context) error
    Ready(ctx context.Context) error
    Captain(ctx context.Context) (api.CaptainInfo, error)
    ListDevices(ctx context.Context) ([]DeviceReport, error)
    Control(ctx context.Context, sn string, opt api.ControlOpt) error
    Upload3MF(ctx context.Context, name string, r io.Reader) (FileRef, error)
    StartPrint(ctx context.Context, req PrintRequest) (ExternalJobRef, error)
    LiveviewJPEG(ctx context.Context, sn string) ([]byte, error)
}

type EventSource interface {
    Subscribe(ctx context.Context) (<-chan DeviceReport, error) // MQTT later
}
```

Edge agent (`app/agent`) polls `ListDevices`, runs `printerstate` tracking, publishes Redis deltas, and consumes commands via the same plane.

### Redis

| Stream | Direction |
|--------|-----------|
| `bf.events.printer` | edge → core/analytics/gateway |
| `bf.events.job` | edge emits progress/finished; core owns queue semantics |
| `bf.commands.edge.{site_id}` | core → edge (start_print, control, upload) |

---

## 7. Lab checklist (your working FM Server)

Run these before coding the full adapter; paste results into an internal runbook (do not commit secrets).

### Connectivity

Point edge at your lab FM via **`.env` only** (never hardcode host/port in source):

```bash
FARM_MANAGER_URL=http://<fm-lan-ip>:8888
FARM_MANAGER_USER=...
FARM_MANAGER_PASSWORD=...
FARM_MANAGER_TRANSPORT=fm_native
```

You may need a Farm Manager local account (or API access) created on that server first. Leave `FARM_MANAGER_URL` empty for fixture mode / CI.

- [ ] From edge host: `curl -vk https://<fm-ip>:8888/captain` (expect TLS/cert behavior)  
- [ ] From edge host: TCP connect to `:1883`  
- [ ] Note FM Server `release` from `/captain` or Server Monitor UI  
- [ ] Confirm printers show in official Client (IDLE / RUNNING)

### Auth path decision

- [ ] **Path A — fm_native:** Can we present a client cert + complete ticket login from Go?  
- [ ] **Path B — CDP:** Can we launch Client with `--remote-debugging-port=9222` on the edge host (usually Windows)?  
- [ ] Capture one successful `GET /devices` response (redact SNs if needed) as a fixture  

### Event path decision

- [ ] MQTT subscribe possible? (credentials/certs?)  
- [ ] If not: measure poll interval that keeps UI fresh without hammering FM  

### Print path smoke (manual Client first)

- [ ] Upload one `.gcode.3mf` via Client  
- [ ] Start on one idle printer  
- [ ] Note task id / file id in Client network tab (CDP) for fixture design  

### Capture for fixtures

Store under `fixtures/fm/` (git-safe redacted JSON):

- `captain.json`  
- `devices_idle.json`  
- `devices_running.json`  
- `tags.json`  
- (later) `task_create_request.json`

---

## 8. Risks & rules of engagement

| Risk | Rule |
|------|------|
| Undocumented API drift | Pin `release`; contract tests vs fixtures; feature-detect |
| mTLS / CNG non-exportable certs | Prefer CDP on Windows edge; document native cert path when available |
| Dual HTTP/HTTPS on 8888 | Client-shaped traffic; don’t assume printer download semantics for edge |
| Accidental cloud coupling | Edge stays LAN-only; no Handy/cloud control for farm-bound printers |
| Over-polling | Delta filter mandatory; MQTT preferred |

---

## 9. References

- [Bambu Farm Manager Quick Start](https://wiki.bambulab.com/en/software/bambu-farm-manager)  
- [Farm Manager Features](https://wiki.bambulab.com/en/software/bambu-farm-features)  
- [Farm Manager FAQ / ports](https://wiki.bambulab.com/en/software/bambu-farm-faq-troubleshoot)  
- [OpenBambuFarmAPI](https://github.com/ATILLLLITA/OpenBambuFarmAPI) — farm MQTT/REST/bind research  
- [GeraldCox bambu-farm-manager-script API_Reference](https://github.com/GeraldCox/bambu-farm-manager-script/blob/main/API_Reference.md) — Client REST capture (~02.04.00.00)  
- [Doridian/OpenBambuAPI](https://github.com/Doridian/OpenBambuAPI) — single-printer cloud/LAN (not farm mode; useful for `push_status` field literacy)
