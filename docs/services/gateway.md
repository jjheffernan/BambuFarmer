# Service shape: `bambufarmer-gateway`

**Language:** Go  
**Container:** `localhost/bambufarmer-gateway`  
**Role:** Operator-facing HTTP entry — auth, reverse-proxy, SSE fan-out  

---

## 1. Responsibility

| Owns | Does not own |
|------|----------------|
| `AUTH_MODE=lan` / `oidc` | Job state machine |
| Session cookies for BambuFarmer users (Argon2id LAN store) | FM credentials |
| Reverse-proxy to core/analytics/integrations/edge/web (+ commerce/slicer/studio-bridge when URLs set) | Business SQL |
| SSE fan-out of Redis events (`GET /api/v1/events/stream`) | Rollup computation |

---

## 2. HTTP surface

| Path | Backend |
|------|---------|
| `GET/POST /api/v1/auth/*` | local (LAN login / password; OIDC code flow) |
| `/api/v1/presets/*`, `/api/v1/jobs/*`, `/api/v1/artifacts/*`, … | core |
| `/api/v1/printers/*` | core projection + edge liveview |
| `/api/v1/analytics/*` | analytics |
| `/api/v1/library/*`, `/api/v1/filament/*` | integrations |
| `/api/v1/edge/*` | edge |
| `/api/v1/commerce/*`, `/api/v1/slice/*`, `/api/v1/studio-bridge/*` | optional plugins when `COMMERCE_URL` / `SLICER_URL` / `STUDIO_BRIDGE_URL` set |
| `/api/v1/events/stream` | gateway SSE ← Redis |
| `/`, `/jobs`, … | web |

---

## 3. Auth modes

| Mode | Behavior |
|------|----------|
| `lan` | Local username/password via core user store; insecure; LAN bind only; boot warning |
| `oidc` | Authorization-code redirect; discovery via issuer; upserts local user + session cookie |
| `local_insecure` | Alias of `lan` |

**LAN vs OIDC / URL’d placement, secrets inventory:** [`../DEPLOYMENT_MODES.md`](../DEPLOYMENT_MODES.md).

---

## 4. Package layout

```text
services/gateway/
  main.go
  auth/         # LAN session helpers (gateway-private)
  tests/
```

---

## 5. Config

From repo-root `.env` (see [`.env.example`](../../.env.example)): `AUTH_MODE`, `SESSION_SECRET`, `CORE_URL`, `WEB_URL`, `ANALYTICS_URL`, `INTEGRATIONS_URL`, `EDGE_URL`, `REDIS_URL`, optional `COMMERCE_URL` / `SLICER_URL` / `STUDIO_BRIDGE_URL`, OIDC_* when mode=oidc.

`EDGE_TOKEN` (optional): when set on gateway **and** edge, gateway injects `X-Edge-Token` on edge proxy requests and edge rejects unauthenticated API calls. Empty = open (default lab compose).

---

## 6. Milestones

1. Proxy + LAN session (done)  
2. Real LAN user store in Postgres via core (done)  
3. SSE stream for printer/job events (done)  
4. OIDC code exchange (done — discovery + token + ID-token verify + session)  
5. Admin per-component update requests (`ops` + host `--watch`)  
6. CSRF origin check on session-mutating requests; login rate limit (`LOGIN_RATE_LIMIT`)
