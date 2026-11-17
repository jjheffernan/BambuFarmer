# Studio bridge (optional plugin) — design

**Status:** slices 1–6 shipped (HTTP/drop + LAN VP + SSDP/CA + plate hints + LAN proxy)  
**Compose profile:** `studio-bridge`  
**Module:** `services/studio-bridge`  
**Role:** Make Bambu Studio / OrcaSlicer able to “send” plates into BambuFarmer without Studio↔FM printer sync, then print via existing `fm_native` — or transparently proxy a **LAN/`direct_printer`** machine.

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
| **True proxy** only for LAN/`direct_printer` targets | FM printers typically are not Studio LAN endpoints |
| No BambuNetwork cloud restore in-tree | FULU/OrcaSlicer-bambulab is a different product (cloud plugin); legal/ToS noise; does not fix FM binding |

```text
  Studio / Orca
       │  SSDP + bind + MQTT/FTP (printer-shaped)
       ▼
  bambufarmer-studio-bridge   (optional profile)
       │  receive: POST artifact + optional CreateJob
       │  proxy:   transparent TCP → LAN printer
       ▼
  core → edge fm_native → Farm Manager → printers
  (or real LAN printer in proxy mode)
```

---

## 3. Modes

### A. Receive (default)

- Advertise a virtual printer (SN, model, access code, bind IP) via SSDP + bind + MQTT/FTPS.
- Accept plate upload from slicer (or HTTP ingest / drop folder).
- Extract plate hints (`printer_model`, filament types) when present in `.3mf`.
- Register artifact in core; policy `queue` | `archive`.
- **Does not** forward control to a real FM printer. Slicer “print” = enqueue for BF.

### B. Proxy (LAN printers only)

- `STUDIO_BRIDGE_MODE=proxy` + `STUDIO_BRIDGE_PROXY_TARGET=<printer-ip>`.
- Transparent TCP relay on printer ports (`3000`, `3002`, `8883`, `990`, PASV range, optional `6000`/`322`).
- Optional SSDP identity via `STUDIO_BRIDGE_PROXY_SN` / `NAME` / `MODEL`.
- Slicer uses the **printer’s** access code (TLS is end-to-end to the printer).
- **Do not** point at FM-bound printers (no usable `:8883`; conflicts with FM).

### Explicit non-goals

- Replacing FM Client adopt/unbind  
- Hosting Farm Manager MQTT for real printers  
- Bundling Studio/Orca binaries  
- Shipping Capacitor for this path  

---

## 4. Protocol surface

| Piece | Typical ports | Bridge job |
|-------|---------------|------------|
| SSDP discovery | UDP 2021 | Advertise VP / proxy identity (`STUDIO_BRIDGE_ADVERTISE_IP`) |
| Bind / detect | TCP 3000, 3002 | Handshake so slicer accepts the VP |
| MQTT/TLS | TCP 8883 | Idle “printer” report + accept print commands (receive) |
| FTPS / file | TCP 990 + PASV | Receive `.gcode.3mf` → ingest (receive) |
| Optional | 6000 / 322 | Relayed in proxy mode |

**TLS trust (receive):** download `GET /api/v1/studio-bridge/status` → `ca_endpoint` / Admin **Download CA (.pem)**. Append the PEM to Studio/Orca’s trusted CA store (same ops idea as Bambuddy’s `printer.cer` flow — we document; we do not patch slicer installs). Fingerprint is `ca_sha256` on status.

**Discovery note:** Docker bridge often blocks SSDP multicast; set `STUDIO_BRIDGE_ADVERTISE_IP` to the host LAN IP, map `2021/udp`, or run with host networking when Studio cannot see the VP.

**Privileged ports:** FTPS 990 / optional 322 need `CAP_NET_BIND_SERVICE` or root; compose documents this.

---

## 5. Service shape

Follow [`plugins.md`](../plugins.md) + commerce/slicer pattern:

1. `services/studio-bridge/` — Go module, Containerfile  
2. Compose `profiles: ["studio-bridge"]`; gateway proxies when `STUDIO_BRIDGE_URL` set  
3. Env: `CORE_URL`, `SITE_ID`, `STUDIO_BRIDGE_MODE` (`receive`|`proxy`), `STUDIO_BRIDGE_POLICY` (`queue`|`archive`), `STUDIO_BRIDGE_DROP_DIR`, `STUDIO_BRIDGE_TARGET_MODEL`  
4. Events: `studio_bridge.received` / `studio_bridge.queued` on `bf.events.studio_bridge` when `REDIS_URL` set  
5. Web: Admin Studio bridge panel (status, CA download, multipart ingest)

### API

- `GET /api/v1/studio-bridge/status` — mode, policy, `vp{…}`, `proxy{…}`, `ca_sha256`, `advertise_ip`  
- `GET /api/v1/studio-bridge/ca.pem` — PEM for slicer trust (after VP has generated certs)  
- `POST /api/v1/studio-bridge/ingest` — multipart `file` (+ optional `name`, `target_model`, `queue`); returns `hints` when extracted  

Drop folder: any `*.3mf` under `STUDIO_BRIDGE_DROP_DIR` is ingested on a 5s poll, then moved to `processed/`.

### LAN virtual printer (`STUDIO_BRIDGE_VP=1`, receive mode)

| Port | Role |
|------|------|
| UDP 2021 | SSDP M-SEARCH reply + periodic NOTIFY |
| TCP 3000, 3002 | Bind/detect listeners (JSON identity reply) |
| TCP 990 + PASV | Implicit TLS FTPS; user `bblp`, password = access code; `STOR` → ingest |
| TCP 8883 | TLS MQTT broker; idle `device/{sn}/report`; auth `bblp` / access code |

**Studio path:** Bind with Access Code → host IP + `STUDIO_BRIDGE_VP_ACCESS_CODE`. Install CA from Admin if TLS fails. Fallback: HTTP ingest / drop folder.

**Verify FTPS without Studio:** FileZilla → host IP, port 990, user `bblp`, password = access code, upload any `*.3mf` > 1KB.

### Proxy mode

```bash
# .env — LAN / direct_printer only
STUDIO_BRIDGE_URL=http://studio-bridge:8080
STUDIO_BRIDGE_MODE=proxy
STUDIO_BRIDGE_PROXY_TARGET=192.168.1.50
STUDIO_BRIDGE_PROXY_SN=BFPROXY000000000
STUDIO_BRIDGE_ADVERTISE_IP=192.168.1.10   # host IP Studio should see
podman compose --env-file .env -f deploy/compose.yml --profile studio-bridge up -d --build
```

Receive + VP:

```bash
STUDIO_BRIDGE_URL=http://studio-bridge:8080
STUDIO_BRIDGE_VP=1
STUDIO_BRIDGE_VP_ACCESS_CODE=12345678
STUDIO_BRIDGE_ADVERTISE_IP=192.168.1.10
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
| **3** | Bind + MQTT stub + FTPS receive (one VP) | **done** |
| **4** | Auto-queue / plate `target_model` + filament hints | **done** |
| **5** | SSDP + CA download + advertise-IP notes | **done** |
| **6** | Optional `proxy` mode for LAN/`direct_printer` only | **done** |

---

## 7. Risks

| Risk | Mitigation |
|------|------------|
| Protocol drift (Studio versions) | Pin tested Studio/Orca versions; feature-detect; fixture captures |
| Host network / port clashes with FM Server | Prefer dedicated host or IPs; document conflict with FM `:1883`/discovery |
| CA install friction | Admin CA link + `ca_sha256`; fail with actionable status |
| Scope creep into “replace FM” | Review against locked stance; proxy gated to LAN |
| AGPL / ToS contamination | Clean-room; no FULU/Bambuddy source in tree |

---

## 8. References (capability only)

- Bambuddy Virtual Printer / Proxy: https://wiki.bambuddy.cool/features/virtual-printer/  
- Bambu Farm FAQ (Studio + FM): https://wiki.bambulab.com/en/software/bambu-farm-faq-troubleshoot  
- FULU OrcaSlicer-bambulab = BambuNetwork restore (cloud), **not** FM sync: https://github.com/FULU-Foundation/OrcaSlicer-bambulab  
- Our FM research: [`edge-farm-manager.md`](./edge-farm-manager.md) · slicer path: [`slicer.md`](./slicer.md)
