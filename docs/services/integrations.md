# Service shape: `bambufarmer-integrations`

**Language:** Go  
**Container:** `localhost/bambufarmer-integrations`  
**Role:** Optional Spoolman + Manyfold adapters behind stable ports  

---

## 1. Responsibility

| Owns | Does not own |
|------|----------------|
| Spoolman HTTP/WS client | Job scheduling |
| Manyfold OAuth/token + TUS upload | FM control |
| Filament consume on `job.finished` | Rate card CRUD |
| Library import/export orchestration | UI |

**MVP rule:** live URL optional. Empty `SPOOLMAN_URL` / `MANYFOLD_URL` → adapter disabled; service still healthy.

---

## 2. Ports (Go interfaces)

```go
type FilamentProvider interface {
    ListSpools(ctx context.Context, filter SpoolFilter) ([]Spool, error)
    UseFilament(ctx context.Context, spoolID string, grams float64) error
    FindMatch(ctx context.Context, material, color, brand string) ([]Spool, error)
}

type LibraryProvider interface {
    Search(ctx context.Context, q string) ([]ModelSummary, error)
    DownloadFile(ctx context.Context, modelID, fileID string) (io.ReadCloser, error)
    UploadModel(ctx context.Context, meta ModelMeta, files []Upload) (ModelRef, error)
}
```

---

## 3. HTTP (proxied via gateway)

| Path | Behavior |
|------|----------|
| `GET /api/v1/integrations/status` | configured? (done stub) |
| `GET /api/v1/filament/spools` | Spoolman list or 503/empty if unset |
| `PUT /api/v1/filament/links` | AMS tray ↔ spool id |
| `GET /api/v1/library/models` | Manyfold search |
| `POST /api/v1/library/import` | → object store + core artifact |
| `POST /api/v1/library/export` | artifact → Manyfold |

Manyfold: OAuth client_credentials when secrets set; optional bearer token path when instance allows.

---

## 4. Workers

- Subscribe `bf.events.job` → on `job.finished` compute grams → Spoolman use  
- Subscribe Spoolman websocket → optional cache invalidate  
- Emit `bf.events.filament`  

---

## 5. Config

`SPOOLMAN_URL`, `SPOOLMAN_API_TOKEN`, `MANYFOLD_URL`, `MANYFOLD_CLIENT_ID`, `MANYFOLD_CLIENT_SECRET`, `MANYFOLD_TOKEN`, `REDIS_URL`, `CORE_URL` / object store as needed.

---

## 6. MVP milestones

1. Status stub (done)  
2. Spoolman list when URL set  
3. Manyfold list when URL set  
4. Consume filament on job finish  
5. Import/export happy paths
