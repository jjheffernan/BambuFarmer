# Service shape: `bambufarmer-web`

**Language:** Go (`html/template` → templ), HTMX, minimal Alpine  
**Container:** `localhost/bambufarmer-web`  
**Role:** Operator GUI mirroring Farm Manager interaction patterns  

---

## 1. Responsibility

| Owns | Does not own |
|------|----------------|
| HTML routes + design tokens | Business rules |
| HTMX partials / SSE-driven swaps | Direct FM calls |
| Client column prefs (later) | Auth IdP (gateway cookies/tokens) |

Prefer browsing via **gateway** (`:18080`) so API + UI share origin. Direct `:13000` remains for debug.

---

## 2. Screens (MVP)

| Route | Purpose |
|-------|---------|
| `/` | Fleet — status chips + grid (SSE later) |
| `/jobs` | Queue / board |
| `/presets` | Filament + print presets (versioned) |
| `/analytics` | Charts from analytics API |
| `/admin` | Site, rates, adapters, lan users |
| `/setup` | First-run install wizard (blocked by gateway until complete) |

Fleet/SSE and setup wizard are live; other panels are still light.

---

## 3. UX rules (from FM Client)

- Status nav: Ready / Attention / Completed / Not in Task  
- Detail drawer, not full-page navigation for printer detail  
- Bulk action bar with multi-select (Alpine)  
- No browser polling of printers — gateway SSE only  

Tokens: see DESIGN_SPEC §11.4 (`--bf-accent`, etc.).

---

## 4. Package layout (target)

```text
services/web/
  main.go
  internal/
    pages/       # handlers
    components/  # templ partials
    static/      # css (or embedded)
```

Migrate from inline `html/template` to [templ](https://templ.guide/) once pages grow.

---

## 5. Config

`HTTP_ADDR`, `PUBLIC_API_BASE` (shown in UI hints; browser calls gateway).

---

## 6. MVP milestones

1. Nav shell (done)  
2. Presets page: list + import via HTMX → gateway  
3. Fleet table fed by SSE fragments  
4. Job composer form  
5. Analytics charts  
6. Admin rate card editor ($/hr, $/kg, kg/hr)
