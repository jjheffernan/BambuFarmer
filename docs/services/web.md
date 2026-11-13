# Service shape: `bambufarmer-web`

**Language:** Go `html/template`, Tailwind CSS (compiled), vanilla JS  
**Container:** `localhost/bambufarmer-web`  
**Role:** Operator GUI mirroring Farm Manager interaction patterns  

---

## 1. Responsibility

| Owns | Does not own |
|------|----------------|
| HTML routes + design tokens | Business rules |
| Client-side SSE handlers + forms | Direct FM calls |
| Compiled CSS (`static/css/app.css`, `go:embed`) | Auth IdP (gateway cookies/tokens) |
| Shared UI helpers (`bfJSON`, `bfFlash`, `bfBindFile`) | |

Prefer browsing via **gateway** (`:18080`) so API + UI share origin. Direct `:13000` remains for debug.

---

## 2. Screens

| Route | Purpose |
|-------|---------|
| `/` | Fleet — filters, cards, drawer, SSE |
| `/jobs` | Create (progressive disclosure) + queue table with expandable rows |
| `/artifacts` | Upload drop zone + list / expandable detail |
| `/presets` | Filament + print presets (versioned) |
| `/filament` | Spoolman status / spools |
| `/library` | Manyfold import/export |
| `/slice` | Optional slicer plugin UI |
| `/commerce` | Optional commerce plugin UI |
| `/analytics` | Charts / utilization from analytics API |
| `/admin` | Site, rates, users, lab controls |
| `/setup` | First-run install wizard (blocked by gateway until complete) |
| `/login` | LAN / OIDC entry |

---

## 3. UX rules (from FM Client)

- Status nav: Ready / Attention / Completed / Not in Task  
- Detail drawer, not full-page navigation for printer detail (fleet card opens drawer)
- Bulk action bar with multi-select  
- No browser polling of printers — gateway SSE only  
- Dense forms use progressive disclosure; dense tables use expandable rows  

Tokens live in Tailwind `@theme` (`static/css/input.css`) as `bf-*` colors/fonts, matching DESIGN_SPEC §11.4.

**CSS build:** edit `static/css/input.css`, then `make web-css` (or `cd services/web && npm run build:css`). Commit the minified `app.css` — the binary embeds it; no Tailwind CDN at runtime.

---

## 4. Package layout

```text
services/web/
  main.go
  setup/           # install wizard (web-private)
  static/
    embed.go       # GET /static/…
    css/input.css  # Tailwind source
    css/app.css    # compiled (embedded)
  package.json     # Tailwind CLI (dev)
  *_ui_test.go
```

Templ migration remains optional; current pages are inline `html/template`.

---

## 5. Config

`HTTP_ADDR`, `PUBLIC_API_BASE` (shown in UI hints; browser calls gateway).

---

## 6. Milestones

1. Nav shell (done; collapsible)  
2. Presets page: list + import/export via gateway (done)
3. Fleet grid fed by SSE (done)
4. Job composer + expandable job rows (done)
5. Analytics charts (done)
6. Admin rate card editor ($/hr, $/kg, kg/hr) (done)
