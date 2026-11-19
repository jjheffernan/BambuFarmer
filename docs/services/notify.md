# Service shape: `bambufarmer-notify` (optional)

**Language:** Go  
**Container:** `localhost/bambufarmer-notify`  
**Compose:** `--profile notify`  
**Role:** Farm event fan-out to CRUD webhook channels; ingest into Redis Streams; printer routing  

---

## 1. Responsibility

Optional notify plane beside the MVP farm loop. Consumes job/alert streams, POSTs typed JSON to operator-configured webhooks, and accepts external ingest for the same path.

| Owns | Does not own |
|------|----------------|
| Channel + printer-route tables | Job state machine / FM control |
| Stream consumer → webhook export | ntfy / Web Push (later) |
| `POST /events` ingest → Redis Streams | Legacy single `NOTIFY_URL` (still in core) |

---

## 2. API

- `GET /api/v1/notify` — status (`service=notify`, channel/printer counts)
- `GET/POST /api/v1/notify/channels` · `GET/PATCH/DELETE …/channels/{id}`
- `GET/POST /api/v1/notify/printers` · `GET/PATCH/DELETE …/printers/{sn}`
- `POST /api/v1/notify/events` — ingest allowlisted types onto Redis Streams + immediate fan-out
- `POST /api/v1/notify/test` — sample `job.finished` to one channel or all enabled

Gateway proxies `/api/v1/notify*` when `NOTIFY_SERVICE_URL` is set; otherwise core’s legacy `NOTIFY_URL` handlers remain.

---

## 3. Enable

```bash
# .env
NOTIFY_SERVICE_URL=http://notify:8080

podman compose --env-file .env -f deploy/compose.yml --profile notify up -d --build
```

See [`plugins.md`](../plugins.md) and [`BUILD_OUT.md`](../BUILD_OUT.md) P3.
