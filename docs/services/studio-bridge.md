# Studio bridge (optional plugin) — design

**Status:** slices 1–2 shipped (skeleton + HTTP/drop ingest); protocol VP later  
**Compose profile:** `studio-bridge`  
**Module:** `services/studio-bridge`  
**Role:** Make Bambu Studio / OrcaSlicer able to “send” plates into BambuFarmer without Studio↔FM printer sync, then print via existing `fm_native`.

Capability inspiration only (no vendored source): [Bambuddy Virtual Printer](https://wiki.bambuddy.cool/features/virtual-printer/) (Print Queue + Proxy), public Studio/LAN protocol literacy. Clean-room: implement against our contracts; do not paste Bambuddy or FULU/Orca trees.

---

## 1. Problem

FM-claimed printers leave Bambu Cloud. Studio LAN bind to those printers is flaky (port conflicts with FM Server, discovery, closed printer `:8883`). Operators still want “slice in Studio → Print” muscle memory.

We already ship: offline slice → upload / Orca CLI → artifact → job → FM. Studio bridge closes the gap where Studio/Orca **discovers something that looks like a printer** and dumps a `.gcode.3mf` into our queue.

---

## 2. Locked stance

| Rule | Why |
|------|-----|
| **FM remains the only process that talks to FM-bound printers** (`fm_native`) | Do not compete with FM MQTT/bind |
| Bridge is a **compose-profiled plugin**, default off | Privileged ports, host network, CA install — opt-in |
| Prefer **receive / queue** mode first | Matches FM fleets; no need for printer `:8883` |
| **True proxy** only for LAN/`direct_printer` targets later | FM printers typically are not Studio LAN endpoints |
| No BambuNetwork cloud restore in-tree | FULU/OrcaSlicer-bambulab is a different product (cloud plugin); legal/ToS noise; does not fix FM binding |

```text
  Studio / Orca
       │  SSDP + bind + MQTT/FTP (printer-shaped)
       ▼
  bambufarmer-studio-bridge   (optional profile)
       │  POST artifact + optional CreateJob
       ▼
  core → edge fm_native → Farm Manager → printers
```

---

## 3. Modes

### A. Receive (MVP — ship first)

Bambuddy analogue: **Immediate / Review / Print Queue** server modes.

- Advertise N virtual printers (SN, model, access code, bind IP).
- Accept plate upload from slicer.
- Register artifact in core (`CORE_URL` / shared `ARTIFACTS_DIR`).
- Policy: archive only | pending review | auto-queue job (unassigned or `target_model`).

**Does not** forward MQTT/control to a real FM printer. Slicer “print” = enqueue for BF.

### B. Proxy (later — LAN printers only)

Bambuddy analogue: **Proxy Mode**.

- Transparent TCP relay to a real printer still in LAN mode (`:8883`, FTPS, bind, etc.).
- Useful when some machines are **not** FM-bound (lab / hybrid), or alongside `direct_printer`.
- Explicit non-goal for FM-claimed fleet until/unless we prove a safe coexistence story.

### Explicit non-goals (v1)

- Replacing FM Client adopt/unbind  
- Hosting Farm Manager MQTT for real printers  
- Bundling Studio/Orca binaries  
- Shipping Capacitor for this path  

---

## 4. Protocol surface (receive MVP)

Slicer expects a printer-shaped stack. Exact fields are version-sensitive — pin against Studio/Orca builds you use; re-verify on upgrade.

| Piece | Typical ports | Bridge job |
|-------|---------------|------------|
| SSDP discovery | UDP 2021 (and related) | Advertise VP identity on LAN (often needs host network) |
| Bind / detect | TCP 3000, 3002 | Handshake so slicer accepts the VP |
| MQTT/TLS | TCP 8883 | Minimal “idle printer” report + accept print commands |
| FTPS / file | TCP 990 + passive range | Receive `.gcode.3mf` |
| Optional later | 6000 FT tunnel, 322 RTSP | Skip for MVP receive |

**TLS trust:** slicers often trust a bundled BBL CA store, not the system store. Receive mode will need an operator-installed CA (or documented workaround), similar to Bambuddy’s `printer.cer` append flow — document, do not silently patch slicer installs from BF.

**Privileged ports:** FTPS 990 / optional 322 need `CAP_NET_BIND_SERVICE` or root; compose must document this.

---

## 5. Service shape (shipped)

Follow [`plugins.md`](../plugins.md) + commerce/slicer pattern:

1. `services/studio-bridge/` — Go module, Containerfile  
2. Compose `profiles: ["studio-bridge"]`; gateway proxies when `STUDIO_BRIDGE_URL` set  
3. Env: `CORE_URL`, `SITE_ID`, `STUDIO_BRIDGE_MODE` (`receive`|`proxy`), `STUDIO_BRIDGE_POLICY` (`queue`|`archive`), `STUDIO_BRIDGE_DROP_DIR`  
4. Events: `studio_bridge.received` / `studio_bridge.queued` on `bf.events.studio_bridge` when `REDIS_URL` set  
5. Web: Admin Studio bridge panel (status + multipart ingest)

### API

- `GET /api/v1/studio-bridge/status` — mode, policy, core/events/drop_dir, `proxy_ready: false`  
- `POST /api/v1/studio-bridge/ingest` — multipart `file` (+ optional `name`, `target_model`, `queue`)  

Drop folder: any `*.3mf` under `STUDIO_BRIDGE_DROP_DIR` is ingested on a 5s poll, then moved to `processed/`.

Enable:

```bash
# .env
STUDIO_BRIDGE_URL=http://studio-bridge:8080
podman compose --env-file .env -f deploy/compose.yml --profile studio-bridge up -d --build
```

Core stays unaware of SSDP/MQTT — bridge calls existing `POST /api/v1/artifacts` and `POST /api/v1/jobs`.

---

## 6. Ship order (vertical slices)

| Slice | Deliverable | Acceptance |
|-------|-------------|------------|
| **0** | This doc + BUILD_OUT / plugins rows | Stance locked |
| **1** | Skeleton service + compose profile + status + Admin chip | **done** |
| **2** | Artifact ingest API (HTTP + drop volume watch) | **done** |
| **3** | Bind + MQTT stub + FTPS receive (one VP) | Studio Send lands as artifact |
| **4** | Auto-queue / review policies + AMS hints from plate metadata | Job appears with `target_model` when known |
| **5** | Multi-VP + host-network / Tailscale notes | Ops runbook |
| **6** | Optional `proxy` mode for LAN/`direct_printer` only | Documented FM exclusion |

Do not start slice 3 until LAN protocol fixtures are ready — protocol work is the expensive part.

---

## 7. Risks

| Risk | Mitigation |
|------|------------|
| Protocol drift (Studio versions) | Pin tested Studio/Orca versions; feature-detect; fixture captures |
| Host network / port clashes with FM Server | Prefer dedicated host or IPs; document conflict with FM `:1883`/discovery |
| CA install friction | Clear Admin docs; fail with actionable status |
| Scope creep into “replace FM” | Review against locked stance; proxy gated |
| AGPL / ToS contamination | Clean-room; no FULU/Bambuddy source in tree |

---

## 8. References (capability only)

- Bambuddy Virtual Printer / Proxy: https://wiki.bambuddy.cool/features/virtual-printer/  
- Bambu Farm FAQ (Studio + FM): https://wiki.bambulab.com/en/software/bambu-farm-faq-troubleshoot  
- FULU OrcaSlicer-bambulab = BambuNetwork restore (cloud), **not** FM sync: https://github.com/FULU-Foundation/OrcaSlicer-bambulab  
- Our FM research: [`edge-farm-manager.md`](./edge-farm-manager.md) · slicer path: [`slicer.md`](./slicer.md)
