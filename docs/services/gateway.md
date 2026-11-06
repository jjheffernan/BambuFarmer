# Service shape: `bambufarmer-gateway`

**Language:** Go  
**Container:** `localhost/bambufarmer-gateway`  
**Role:** Operator-facing HTTP entry — auth, reverse-proxy, SSE fan-out  

---

## 1. Responsibility

| Owns | Does not own |
|------|----------------|
| `AUTH_MODE=lan` / `oidc` | Job state machine |
| Session / JWT for BambuFarmer users | FM credentials |
| Reverse-proxy to core/analytics/integrations/edge/web | Business SQL |
| SSE/WebSocket fan-out of Redis events (planned) | Rollup computation |

---

## 2. HTTP surface (target)

| Path | Backend |
|------|---------|
| `GET/POST /api/v1/auth/*` | local (lan login / OIDC) |
| `/api/v1/presets/*` | core |
| `/api/v1/jobs/*` | core (planned) |
| `/api/v1/printers/*` | core projection + edge liveview |
| `/api/v1/analytics/*` | analytics |
| `/api/v1/library/*`, `/api/v1/filament/*` | integrations |
| `/api/v1/edge/*` | edge |
| `/api/v1/events/stream` | gateway SSE ← Redis |
| `/`, `/jobs`, … | web |

---

## 3. Auth modes

| Mode | Behavior |
|------|----------|
| `lan` | Local username/password; insecure; LAN bind only; boot warning |
| `oidc` | Standard OIDC code flow |
| `local_insecure` | Alias of `lan` |

Stub today: `/api/v1/auth/me`, `/api/v1/auth/login` (lan accepts any non-empty user).

---

## 4. Package layout (target)

```text
services/gateway/
  main.go
  internal/
    auth/       # lan + oidc
    proxy/      # upstream URLs
    sse/        # Redis → browser
    middleware/ # request id, authn
```

---

## 5. Config

`AUTH_MODE`, `SESSION_SECRET`, `CORE_URL`, `WEB_URL`, `ANALYTICS_URL`, `INTEGRATIONS_URL`, `EDGE_URL`, `REDIS_URL`, OIDC_* when mode=oidc.

---

## 6. MVP milestones

1. Proxy + lan stub (done)  
2. Real lan user store in Postgres (via core)  
3. SSE stream for printer/job events  
4. OIDC path  
5. Rate limiting / CSRF for cookie sessions
