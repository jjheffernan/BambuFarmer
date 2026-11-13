# Service shape: `bambufarmer-commerce` (optional)

**Language:** Go  
**Container:** `localhost/bambufarmer-commerce`  
**Compose:** `--profile commerce`  
**Role:** Shop webhook → core job; pack/ship status echo for operators  

---

## 1. Responsibility

Optional plugin beside the MVP farm loop. Does **not** own printer control or job state machines — it calls core when `CORE_URL` is set.

| Owns | Does not own |
|------|----------------|
| Store webhook intake | Job scheduling / FM upload |
| Pack / ship status echo | Payment processing |
| Status for the Commerce UI page | Postgres tables (v1) |

---

## 2. API

- `GET /api/v1/commerce/status` — service readiness + short capability note  
- `POST /api/v1/commerce/webhook` — queues a core job when `CORE_URL` is set  
- `POST /api/v1/commerce/orders/status` — echoes packed/shipped for UI/ops  

Gateway proxies `/api/v1/commerce/` when the commerce URL is configured.

---

## 3. Enable

```bash
podman compose --env-file .env -f deploy/compose.yml --profile commerce up -d --build
```

See [`plugins.md`](../plugins.md) and [`BUILD_OUT.md`](../BUILD_OUT.md) P4.
