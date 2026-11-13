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
| Reverse-proxy to core/analytics/integrations/edge/web (+ commerce/slicer when URLs set) | Business SQL |
| SSE fan-out of Redis events (`GET /api/v1/events/stream`) | Rollup computation |

---

## 2. HTTP surface

| Path | Backend |
|------|---------|
| `GET/POST /api/v1/auth/*` | local (LAN login / password; OIDC stubs) |
| `/api/v1/presets/*`, `/api/v1/jobs/*`, `/api/v1/artifacts/*`, … | core |
| `/api/v1/printers/*` | core projection + edge liveview |
| `/api/v1/analytics/*` | analytics |
| `/api/v1/library/*`, `/api/v1/filament/*` | integrations |
| `/api/v1/edge/*` | edge |
| `/api/v1/commerce/*`, `/api/v1/slice/*` | optional plugins when `COMMERCE_URL` / `SLICER_URL` set |
| `/api/v1/events/stream` | gateway SSE ← Redis |
| `/`, `/jobs`, … | web |

---

## 3. Auth modes

| Mode | Behavior |
|------|----------|
| `lan` | Local username/password via core user store; insecure; LAN bind only; boot warning |
| `oidc` | Intended OIDC code flow — **authorize/callback still stubs** (login button + stub JSON only) |
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

From repo-root `.env` (see [`.env.example`](../../.env.example)): `AUTH_MODE`, `SESSION_SECRET`, `CORE_URL`, `WEB_URL`, `ANALYTICS_URL`, `INTEGRATIONS_URL`, `EDGE_URL`, `REDIS_URL`, optional `COMMERCE_URL` / `SLICER_URL`, OIDC_* when mode=oidc.

`EDGE_TOKEN` is reserved in compose/example but **not enforced** yet (no shared-secret check on edge proxy routes).

---

## 6. Milestones

1. Proxy + LAN session (done)  
2. Real LAN user store in Postgres via core (done)  
3. SSE stream for printer/job events (done)  
4. OIDC code exchange (**stub** — wire IdP client next)  
5. Rate limiting / CSRF for cookie sessions (later)
