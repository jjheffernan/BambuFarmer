# Service shape: `bambufarmer-integrations`

**Language:** Go  
**Container:** `localhost/bambufarmer-integrations`  
**Role:** Optional Spoolman + Manyfold adapters behind stable ports  

Upstream references:

- Spoolman: [Donkie/Spoolman](https://github.com/Donkie/Spoolman) — REST `/api/v1/*` (no built-in auth)
- Manyfold: [manyfold3d/manyfold](https://github.com/manyfold3d/manyfold) — REST v0, `Accept: application/vnd.manyfold.v0+json`

---

## 1. Responsibility

| Owns | Does not own |
|------|----------------|
| Spoolman HTTP client | Job scheduling |
| Manyfold HTTP client (PAT / OAuth) | FM control |
| Filament consume on `job.finished` (next) | Rate card CRUD |
| Library import/export orchestration (next) | UI |

**MVP rule:** live URL optional. Empty `SPOOLMAN_URL` / `MANYFOLD_URL` → adapter disabled; service still healthy.

---

## 2. Layout

```text
services/integrations/
  main.go
  app/
    config/
    status/
    spoolman/     # GET /api/v1/spool
    manyfold/     # GET /models (JSON-LD)
    httpapi/      # BambuFarmer routes
  tests/
```

---

## 3. Ports (Go interfaces)

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

Today: `ListSpools` / `ListModels` implemented on concrete clients; full interfaces land with consume + import/export.

---

## 4. HTTP (proxied via gateway)

| Path | Behavior |
|------|----------|
| `GET /api/v1/integrations/status` | configured? + upstream repo refs |
| `GET /api/v1/filament/spools` | Spoolman `GET /api/v1/spool?limit=1000` or empty if unset |
| `GET /api/v1/library/models` | Manyfold `GET /models` or empty if unset |
| `GET /api/v1/filament/links?tray=` | AMS tray index → spool id from `AMS_SLOT_MAP` (empty map → none) |
| `POST /api/v1/library/import` | Create a Manyfold model by name (`POST /models`); empty URL does not call Manyfold |
| `POST /api/v1/library/export` | artifact → Manyfold (next) |

**Manyfold auth:** prefer `MANYFOLD_TOKEN` (personal access token). Falls back to OAuth2 `client_credentials` when `MANYFOLD_CLIENT_ID` / `SECRET` are set. Always send `Accept: application/vnd.manyfold.v0+json`.

**Spoolman auth:** Spoolman has none; `SPOOLMAN_API_TOKEN` is forwarded as `X-API-Key` + `Authorization: Bearer` for reverse proxies.

---

## 5. Config

`SPOOLMAN_URL`, `SPOOLMAN_API_TOKEN`, `LOW_STOCK_GRAMS`, `AMS_SLOT_MAP`, `MANYFOLD_URL`, `MANYFOLD_CLIENT_ID`, `MANYFOLD_CLIENT_SECRET`, `MANYFOLD_TOKEN`, `REDIS_URL`, `CORE_URL` / object store as needed.

---

## 6. MVP milestones

1. Status stub — **done**  
2. Spoolman list when URL set — **done**  
3. Manyfold list when URL set — **done**  
4. Consume filament on job finish  
5. Import/export happy paths
