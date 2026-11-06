# Service shape: `bambufarmer-edge`

**Language:** Go  
**Container:** `localhost/bambufarmer-edge`  
**Trust boundary:** Printer / Farm Manager LAN  
**Depends on:** Redis; Farm Manager Server; optional FM Client (CDP)  
**Consumers:** gateway (status/liveview proxy), core (commands), analytics (events)

Deep protocol research: [`edge-farm-manager.md`](./edge-farm-manager.md).

---

## 1. Responsibility

Edge is the **only** process that speaks Farm Manager. It:

1. Authenticates like the FM Client (`fm_native` or `cdp`)  
2. Ingests printer telemetry (MQTT preferred, REST delta fallback)  
3. Publishes normalized events to Redis Streams  
4. Executes print/control/upload commands from `bf.commands.edge.{site_id}`  
5. Never stores operator passwords for BambuFarmer users (FM service account only)

---

## 2. Process layout

```text
┌─ HTTP (:8080) ─────────────────────────────────────┐
│  GET /healthz /readyz                              │
│  GET /api/v1/edge/status                           │
│  GET /api/v1/edge/liveview/{sn}  (JPEG proxy)      │
└────────────────────────────────────────────────────┘
┌─ Workers ──────────────────────────────────────────┐
│  auth refresher (JWT)                              │
│  event source (MQTT | poll+delta) → Redis publish  │
│  command consumer ← Redis bf.commands.edge.{site}  │
└────────────────────────────────────────────────────┘
```

---

## 3. Config

| Env | Default | Notes |
|-----|---------|-------|
| `SITE_ID` | `default` | Single-site MVP |
| `FARM_MANAGER_URL` | required for live | e.g. `https://192.168.x.x:8888` |
| `FARM_MANAGER_TRANSPORT` | `fm_native` | or `cdp` |
| `FARM_MANAGER_USER` / `PASSWORD` | | Local FM admin/service user |
| `EDGE_CERT_DIR` | | Client cert material when native |
| `CDP_URL` | | e.g. `http://127.0.0.1:9222` |
| `REDIS_URL` | required | Streams |
| `HTTP_ADDR` | `:8080` | Status API only |

---

## 4. Packages (target)

```text
services/edge/internal/
  config/
  fm/            # FarmControlPlane implementation
  mqtt/
  normalize/
  delta/
  bus/
```

Shared types: `packages/events`, `packages/api` (`PrinterSummary` → expand to full snapshot DTO).

---

## 5. MVP milestones

| Milestone | Done when |
|-----------|-----------|
| M0 | Status reports transport + `connected=false` (today) |
| M1 | `fm_native` or CDP login + `GET /devices` → status JSON |
| M2 | Delta poll → `bf.events.printer` |
| M3 | Command: pause/resume/stop/bed_clean |
| M4 | Upload 3MF + start print command path |
| M5 | MQTT reports (if broker access proven) |

---

## 6. Readiness

`/readyz` should fail when:

- Redis unreachable, or  
- FM session invalid and reconnect exhausted  

`/healthz` = process up (even if FM down) so orchestrators can restart cleanly.

---

## 7. Non-goals

- Binding/unbinding printers (use official Client for farm topology changes in MVP)  
- Firmware OTA UI  
- Replacing FM Server  
- Serving the operator SPA
