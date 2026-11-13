# Service shape: `bambufarmer-edge`

**Language:** Go  
**Container:** `localhost/bambufarmer-edge`  
**Trust boundary:** Printer / Farm Manager LAN  
**Depends on:** Redis; Farm Manager Server; optional FM Client (CDP)  
**Consumers:** gateway (status/liveview proxy), core (commands), analytics (events)

Deep protocol research: [`edge-farm-manager.md`](./edge-farm-manager.md).  
Shared abstraction: [`packages/farm`](../../packages/farm) (`farm.ControlPlane`).

---

## 1. Responsibility

Edge is the **only** process that speaks Bambu farm backends. It:

1. Selects a `farm.ControlPlane` (`fm_native` | `fixture` | `cdp` stub | `direct_printer`; `fleet_hub` reserved)  
2. Runs the **edge agent** (poll / track / delta / Redis publish / command consumer)  
3. Never stores BambuFarmer operator passwords (FM service account only)

---

## 2. Architecture

```text
┌─ HTTP (:8080) ─────────────────────────────────────┐
│  GET  /api/v1/edge/status                          │
│  POST /api/v1/edge/control/{sn}                    │
└────────────────────────────────────────────────────┘
┌─ agent ────────────────────────────────────────────┐
│  farm.ControlPlane  ←── bambu.Open(config)         │
│       │                                            │
│       ├─ native (fm_native REST)                   │
│       ├─ fixture (fixtures/fm JSON)                │
│       ├─ direct_printer (in-process; no LAN)       │
│       ├─ cdp (stub)                                │
│       └─ none                                      │
│  tracker (printerstate) + delta filter             │
│  bus → bf.events.printer / ← bf.commands.edge.*    │
└────────────────────────────────────────────────────┘
```

`packages/farm.ControlPlane` methods: `Connect`, `Captain`, `ListDevices`, `Control`, `Upload3MF`, `StartPrint`, `LiveviewJPEG`.

---

## 3. Config

| Env | Default | Notes |
|-----|---------|-------|
| `SITE_ID` | `default` | Single-site MVP |
| `FARM_MANAGER_URL` | | Live FM → `fm_native` (or `cdp` if transport set) |
| `FARM_MANAGER_TRANSPORT` | `fm_native` | or `cdp` / `direct_printer` |
| `FARM_MANAGER_USER` / `PASSWORD` | | FM service account |
| `DIRECT_PRINTERS` | | `sn:model,…` when transport is `direct_printer` |
| `FARM_MQTT_URL` | | Optional MQTT reports (needs certs in `EDGE_CERT_DIR`) |
| `EDGE_FIXTURE_DIR` | | Used when URL empty |
| `REDIS_URL` | | Streams |
| `HTTP_ADDR` | `:8080` | |

---

## 4. Packages

```text
packages/farm/                 # ControlPlane abstraction (shared)
services/edge/app/             # process-private (not packages/)
  agent/                       # edge runtime
  bambu/                       # factory + native|fixture|cdp|direct|none
  farmmanager/                 # low-level Farm Manager HTTP (native only)
  mqtt/                        # optional report subscribe
  tracker/ normalize/ delta/
  bus/ fixture/ config/
```

---

## 5. MVP milestones

| Milestone | Status |
|-----------|--------|
| M1 `farm.ControlPlane` + agent + native/fixture | **done** |
| M2 Redis delta publish + commands | **done** |
| M3 HTTP control | **done** |
| M4 Upload3MF + StartPrint on native | **done** (fixture/unit; live FM smoke deferred) |
| M5 MQTT EventSource | **done when** `FARM_MQTT_URL` + certs set |
| `direct_printer` in-process plane | **done** (no LAN) |
| CDP real implementation | stub |

---

## 6. Readiness

`/healthz` = process up. `/readyz` should eventually require Redis + plane `Ready()` (tighten next).
