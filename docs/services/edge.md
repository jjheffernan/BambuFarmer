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

1. Selects a `farm.ControlPlane` (`fm_native` | `fixture` | `cdp` stub | later `fleet_hub`)  
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
| `FARM_MANAGER_TRANSPORT` | `fm_native` | or `cdp` |
| `FARM_MANAGER_USER` / `PASSWORD` | | FM service account |
| `EDGE_FIXTURE_DIR` | | Used when URL empty |
| `REDIS_URL` | | Streams |
| `HTTP_ADDR` | `:8080` | |

---

## 4. Packages

```text
packages/farm/                 # ControlPlane abstraction (shared)
services/edge/app/             # process-private (not packages/)
  agent/                       # edge runtime
  bambu/                       # factory + native|fixture|cdp|none
  farmmanager/                 # low-level Farm Manager HTTP (native only)
  tracker/ normalize/ delta/
  bus/ fixture/ config/
```

---

## 5. MVP milestones

| Milestone | Status |
|-----------|--------|
| M1 `farm.ControlPlane` + agent + native/fixture | **done** |
| M2 Redis delta publish + commands | **done** (StartPrint stub on native) |
| M3 HTTP control | **done** |
| M4 Upload3MF + StartPrint on native | next |
| M5 MQTT EventSource | later |
| CDP real implementation | later |

---

## 6. Readiness

`/healthz` = process up. `/readyz` should eventually require Redis + plane `Ready()` (tighten next).
