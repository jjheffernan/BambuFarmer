package main

import (
	"html/template"
	"log/slog"
	"net/http"
	"os"

	"github.com/jjheffernan/BambuFarmer/packages/httpserver"
	"github.com/jjheffernan/BambuFarmer/services/web/setup"
	"github.com/jjheffernan/BambuFarmer/services/web/static"
)

type pageData struct {
	Title   string
	Active  string
	APIBase string
}

func main() {
	service := "web"
	apiBase := getenv("PUBLIC_API_BASE", "http://127.0.0.1:18080")
	mux := httpserver.NewMux(service)
	mux.Handle("GET /static/", static.Handler())

	tpl := template.Must(template.New("layout").Parse(layoutHTML))

	render := func(w http.ResponseWriter, title, active string) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = tpl.Execute(w, pageData{Title: title, Active: active, APIBase: apiBase})
	}

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, _ *http.Request) {
		render(w, "Fleet", "fleet")
	})
	mux.HandleFunc("GET /jobs", func(w http.ResponseWriter, _ *http.Request) {
		render(w, "Jobs", "jobs")
	})
	mux.HandleFunc("GET /artifacts", func(w http.ResponseWriter, _ *http.Request) {
		render(w, "Artifacts", "artifacts")
	})
	mux.HandleFunc("GET /presets", func(w http.ResponseWriter, _ *http.Request) {
		render(w, "Presets", "presets")
	})
	mux.HandleFunc("GET /filament", func(w http.ResponseWriter, _ *http.Request) {
		render(w, "Filament", "filament")
	})
	mux.HandleFunc("GET /library", func(w http.ResponseWriter, _ *http.Request) {
		render(w, "Library", "library")
	})
	mux.HandleFunc("GET /slice", func(w http.ResponseWriter, _ *http.Request) {
		render(w, "Slice", "slice")
	})
	mux.HandleFunc("GET /commerce", func(w http.ResponseWriter, _ *http.Request) {
		render(w, "Commerce", "commerce")
	})
	mux.HandleFunc("GET /analytics", func(w http.ResponseWriter, _ *http.Request) {
		render(w, "Analytics", "analytics")
	})
	mux.HandleFunc("GET /admin", func(w http.ResponseWriter, _ *http.Request) {
		render(w, "Admin", "admin")
	})
	mux.HandleFunc("GET /login", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = template.Must(template.New("login").Parse(loginHTML)).Execute(w, pageData{Title: "Login", APIBase: apiBase})
	})
	setup.Register(mux, apiBase)

	if err := httpserver.ListenAndServe(httpserver.Options{
		Service: service,
		Addr:    httpserver.DefaultAddr(),
		Mux:     mux,
	}); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

const layoutHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1"/>
  <title>{{.Title}} · BambuFarmer</title>
  <link rel="preconnect" href="https://fonts.googleapis.com"/>
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin/>
  <link href="https://fonts.googleapis.com/css2?family=IBM+Plex+Mono:wght@400;500&family=IBM+Plex+Sans:wght@400;500;600;700&family=Syne:wght@600;700&display=swap" rel="stylesheet"/>
  <link rel="stylesheet" href="/static/css/app.css"/>
  <script src="https://unpkg.com/htmx.org@2.0.4"></script>
  <script>
    // Parse JSON API responses; surface non-JSON / error bodies clearly.
    async function bfJSON(url, init) {
      init = Object.assign({ credentials: 'include' }, init || {});
      const r = await fetch(url, init);
      const text = await r.text();
      let data = null;
      if (text) {
        try { data = JSON.parse(text); }
        catch (e) {
          throw new Error(r.status + ' ' + url + ' — not JSON: ' + text.slice(0, 160));
        }
      }
      if (!r.ok) {
        const msg = (data && (data.error || data.message)) || text.slice(0, 160) || r.statusText;
        throw new Error(r.status + ' ' + url + ' — ' + msg);
      }
      return data;
    }
  </script>
</head>
<body>
  <div class="bf-shell">
    <aside class="bf-aside">
      <div class="bf-brand">BambuFarmer<span>LAN farm ops</span></div>
      <nav class="bf-nav">
        <a href="/" {{if eq .Active "fleet"}}class="active"{{end}}>Fleet</a>
        <a href="/jobs" {{if eq .Active "jobs"}}class="active"{{end}}>Jobs</a>
        <a href="/artifacts" {{if eq .Active "artifacts"}}class="active"{{end}}>Artifacts</a>
        <a href="/presets" {{if eq .Active "presets"}}class="active"{{end}}>Presets</a>
        <a href="/filament" {{if eq .Active "filament"}}class="active"{{end}}>Filament</a>
        <a href="/library" {{if eq .Active "library"}}class="active"{{end}}>Library</a>
        <a href="/slice" {{if eq .Active "slice"}}class="active"{{end}}>Slice</a>
        <a href="/commerce" {{if eq .Active "commerce"}}class="active"{{end}}>Commerce</a>
        <a href="/analytics" {{if eq .Active "analytics"}}class="active"{{end}}>Analytics</a>
        <a href="/admin" {{if eq .Active "admin"}}class="active"{{end}}>Admin</a>
      </nav>
      <div class="mt-auto text-xs text-bf-muted space-y-2">
        <p id="bf-user-label" class="hidden"></p>
        <button type="button" id="bf-logout" class="bf-btn hidden w-full">Log out</button>
        <p class="hidden md:block">Gateway API<br/><code class="text-[0.7rem] break-all">{{.APIBase}}</code></p>
        <p class="font-mono text-[0.7rem]" id="bf-build-version">…</p>
      </div>
    </aside>
    <script>
      (async () => {
        const API = {{printf "%q" .APIBase}};
        try {
          const v = await bfJSON(API + '/api/v1/version');
          const el = document.getElementById('bf-build-version');
          if (el) {
            const parts = [v.version || v.Version, v.commit || v.Commit].filter(Boolean);
            el.textContent = parts.length ? parts.join(' · ') : 'dev';
            if (v.built_at || v.BuiltAt) el.title = String(v.built_at || v.BuiltAt);
          }
        } catch (_) {
          const el = document.getElementById('bf-build-version');
          if (el) el.textContent = '';
        }
        try {
          const me = await bfJSON(API + '/api/v1/auth/me');
          if (me.authenticated && me.user) {
            const el = document.getElementById('bf-user-label');
            const btn = document.getElementById('bf-logout');
            if (el) { el.textContent = me.user.username + ' · ' + (me.roles||[]).join(','); el.classList.remove('hidden'); }
            if (btn) {
              btn.classList.remove('hidden');
              btn.onclick = async () => {
                await fetch(API + '/api/v1/auth/logout', { method: 'POST', credentials: 'include' });
                location.href = '/login';
              };
            }
          } else if (me.auth_mode === 'lan') {
            // Gateway will 401/redirect protected routes; soft-hint only.
          }
        } catch (_) {}
      })();
    </script>
    <main class="bf-main">
      <h1 class="bf-title">{{.Title}}</h1>
      {{if eq .Active "fleet"}}
        <p class="bf-lede">Live fleet via gateway SSE — no browser polling of Farm Manager.</p>
        <div id="sse-status" class="bf-status">connecting…</div>
        <div id="alert-banner" class="hidden mb-3 border border-bf-danger/40 bg-bf-danger/10 rounded-bf px-3 py-2 text-sm" role="status"></div>
        <div class="flex flex-wrap gap-2 mb-3" id="bucket-chips">
          <span class="bf-chip on" data-bucket="all">All</span>
          <span class="bf-chip" data-bucket="ready_to_start">Ready to Start</span>
          <span class="bf-chip" data-bucket="attention_needed">Attention Needed</span>
          <span class="bf-chip" data-bucket="completed_task">Completed</span>
          <span class="bf-chip" data-bucket="not_in_task">Not in Task</span>
          <span class="bf-chip" data-bucket="printing">Printing</span>
        </div>
        <div class="flex flex-wrap items-center gap-2 mb-4">
          <input class="bf-input max-w-xs" id="fleet-q" type="search" placeholder="Search name, SN, model"/>
          <div class="flex flex-wrap gap-2" id="tag-chips">
            <span class="bf-chip on" data-tag="">All tags</span>
          </div>
        </div>
        <div id="bulk-bar" class="hidden sticky top-0 z-10 mb-3 flex flex-wrap items-center gap-2 border border-bf-border rounded-bf bg-bf-2/95 px-3 py-2 backdrop-blur">
          <span id="bulk-count" class="text-sm text-bf-muted">0 selected</span>
          <button type="button" class="bf-btn" data-bulk="pause">Pause</button>
          <button type="button" class="bf-btn" data-bulk="resume">Resume</button>
          <button type="button" class="bf-btn" data-bulk="stop">Stop</button>
          <button type="button" class="bf-btn" data-bulk="bed_clean">Bed clean</button>
          <button type="button" class="bf-btn" id="bulk-clear">Clear</button>
        </div>
        <div id="printer-grid" class="grid gap-3 grid-cols-[repeat(auto-fill,minmax(220px,1fr))]"></div>
        <div id="drawer-backdrop" class="bf-drawer-backdrop" hidden></div>
        <aside id="printer-drawer" class="bf-drawer" aria-hidden="true">
          <div class="flex items-start justify-between gap-3 mb-4">
            <div>
              <h2 class="font-display text-lg" id="drawer-title">Printer</h2>
              <p class="bf-card-meta m-0 font-mono text-xs" id="drawer-sn"></p>
            </div>
            <button type="button" class="bf-btn" id="drawer-close">Close</button>
          </div>
          <div id="drawer-body" class="space-y-4 text-sm"></div>
        </aside>
        <script>
          const API = {{printf "%q" .APIBase}};
          const printers = new Map();
          const selected = new Set();
          let filter = 'all';
          let tagFilter = '';
          let fleetQ = '';
          let openSn = '';
          let liveviewTimer = null;
          const statusEl = document.getElementById('sse-status');
          const grid = document.getElementById('printer-grid');
          const bulkBar = document.getElementById('bulk-bar');
          const bulkCount = document.getElementById('bulk-count');
          const drawer = document.getElementById('printer-drawer');
          const drawerBackdrop = document.getElementById('drawer-backdrop');
          const drawerBody = document.getElementById('drawer-body');
          const drawerTitle = document.getElementById('drawer-title');
          const drawerSn = document.getElementById('drawer-sn');
          const alertBanner = document.getElementById('alert-banner');
          const recentAlerts = [];
          const LIVEVIEW_MS = 3000;

          document.querySelectorAll('#bucket-chips .bf-chip').forEach(chip => {
            chip.addEventListener('click', () => {
              document.querySelectorAll('#bucket-chips .bf-chip').forEach(c => c.classList.remove('on'));
              chip.classList.add('on');
              filter = chip.dataset.bucket;
              render();
            });
          });
          document.getElementById('fleet-q').oninput = (ev) => {
            fleetQ = String(ev.target.value || '').trim().toLowerCase();
            render();
          };
          const tagChips = document.getElementById('tag-chips');
          tagChips.addEventListener('click', (ev) => {
            const chip = ev.target.closest('[data-tag]');
            if (!chip) return;
            tagChips.querySelectorAll('.bf-chip').forEach(c => c.classList.remove('on'));
            chip.classList.add('on');
            tagFilter = chip.dataset.tag || '';
            render();
          });

          function rebuildTagChips() {
            const tags = new Set();
            printers.forEach(p => (p.tags || []).forEach(t => { if (t) tags.add(String(t)); }));
            const sorted = [...tags].sort((a, b) => a.localeCompare(b));
            const cur = tagFilter;
            let html = '<span class="bf-chip' + (cur === '' ? ' on' : '') + '" data-tag="">All tags</span>';
            sorted.forEach(t => {
              html += '<span class="bf-chip' + (cur === t ? ' on' : '') + '" data-tag="' + esc(t) + '">' + esc(t) + '</span>';
            });
            tagChips.innerHTML = html;
          }

          function upsert(p) {
            if (!p || !p.external_sn) return;
            const prev = printers.get(p.external_sn) || {};
            printers.set(p.external_sn, Object.assign({}, prev, p));
            rebuildTagChips();
            render();
            if (openSn === p.external_sn) fillDrawer(printers.get(p.external_sn));
          }

          function esc(s) {
            return String(s == null ? '' : s)
              .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/"/g, '&quot;');
          }

          function taskLine(p) {
            const t = p.task || {};
            const name = t.subtask_name || '';
            const id = t.external_task_id || '';
            if (!name && !id) return '';
            const label = name || id;
            return '<div class="bf-card-meta mt-1 truncate" title="' + esc(label) + '">Task · ' + esc(label) + '</div>';
          }

          function amsChips(p) {
            const units = p.ams || [];
            const chips = [];
            units.forEach(u => (u.trays || []).forEach(tr => {
              const typ = tr.type || '—';
              const rem = tr.remain ? (' ' + tr.remain + '%') : '';
              const hex = (tr.color || '').replace(/^#/, '').slice(0, 6);
              const style = hex.length === 6 ? (' style="border-color:#' + hex + '"') : '';
              const title = 'AMS ' + (u.id ?? 0) + ' slot ' + (tr.slot ?? '?');
              chips.push('<span class="bf-chip text-xs"' + style + ' title="' + esc(title) + '">' + esc(typ + rem) + '</span>');
            }));
            if (!chips.length) return '';
            return '<div class="flex flex-wrap gap-1 mt-2" data-ams-chips>' + chips.join('') + '</div>';
          }

          function liveviewURL(sn) {
            return API + '/api/v1/printers/' + encodeURIComponent(sn) + '/liveview?t=' + Date.now();
          }

          function stopLiveview() {
            if (liveviewTimer) {
              clearInterval(liveviewTimer);
              liveviewTimer = null;
            }
          }

          function startLiveview(sn) {
            stopLiveview();
            const bump = () => {
              if (openSn !== sn) return;
              const img = document.getElementById('drawer-liveview');
              if (!img) return;
              img.src = liveviewURL(sn);
            };
            bump();
            liveviewTimer = setInterval(bump, LIVEVIEW_MS);
          }

          function openDrawer(sn) {
            openSn = sn;
            const p = printers.get(sn);
            if (!p) return;
            drawer.classList.add('open');
            drawerBackdrop.classList.add('open');
            drawerBackdrop.hidden = false;
            drawer.setAttribute('aria-hidden', 'false');
            fillDrawer(p);
            startLiveview(sn);
          }

          function closeDrawer() {
            openSn = '';
            stopLiveview();
            drawer.classList.remove('open');
            drawerBackdrop.classList.remove('open');
            drawerBackdrop.hidden = true;
            drawer.setAttribute('aria-hidden', 'true');
          }

          function fillDrawer(p) {
            drawerTitle.textContent = p.name || p.external_sn || 'Printer';
            drawerSn.textContent = p.external_sn || '';
            const t = p.task || {};
            const temps = [
              ['Nozzle', p.nozzle_temper, p.nozzle_target_temper],
              ['Bed', p.bed_temper, p.bed_target_temper],
              ['Chamber', p.chamber_temper, null]
            ].map(row => {
              const cur = row[1] != null ? Number(row[1]).toFixed(1) + '°' : '—';
              const tgt = row[2] != null && row[2] !== 0 ? ' → ' + Number(row[2]).toFixed(0) + '°' : '';
              return '<div class="flex justify-between gap-2"><span class="text-bf-muted">' + esc(row[0]) +
                '</span><span class="font-mono text-xs">' + esc(cur + tgt) + '</span></div>';
            }).join('');
            let amsRows = '';
            (p.ams || []).forEach(u => (u.trays || []).forEach(tr => {
              amsRows += '<tr><td class="font-mono text-xs">' + esc(u.id) + '/' + esc(tr.slot) +
                '</td><td>' + esc(tr.type || '—') + '</td><td class="font-mono text-xs">' +
                esc(tr.remain ? tr.remain + '%' : '—') + '</td><td class="font-mono text-xs truncate">' +
                esc(tr.tag_uid || '') + '</td></tr>';
            }));
            const hms = (p.hms || []).map(c => '<span class="bf-chip text-xs">' + esc(c) + '</span>').join('') ||
              '<span class="bf-empty">None</span>';
            const tags = (p.tags || []).map(c => '<span class="bf-chip text-xs">' + esc(c) + '</span>').join('') ||
              '<span class="bf-empty">None</span>';
            const prevImg = document.getElementById('drawer-liveview');
            const keepSrc = prevImg && prevImg.dataset.sn === p.external_sn ? prevImg.src : liveviewURL(p.external_sn);
            drawerBody.innerHTML =
              '<div class="bf-panel p-3">' +
              '<div class="bf-card-meta m-0 mb-2">' + esc(p.model || '') + ' · <span class="state-' +
              esc(p.gcode_state || '') + '">' + esc(p.gcode_state || '?') + '</span>' +
              (p.mc_percent ? ' · ' + p.mc_percent + '%' : '') +
              (p.mc_remaining_time ? ' · ' + p.mc_remaining_time + 's left' : '') + '</div>' +
              '<div class="space-y-1">' + temps + '</div>' +
              (p.wifi_signal ? '<p class="bf-card-meta mt-2 mb-0">Wi‑Fi ' + esc(p.wifi_signal) + '</p>' : '') +
              '</div>' +
              '<div><h3 class="font-display text-base mb-2">Task</h3>' +
              (t.subtask_name || t.external_task_id
                ? '<p class="m-0">' + esc(t.subtask_name || '—') + '</p><p class="bf-card-meta font-mono text-xs">' +
                  esc(t.external_task_id || '') + '</p>'
                : '<p class="bf-empty m-0">No active task</p>') +
              '</div>' +
              '<div><h3 class="font-display text-base mb-2">AMS</h3>' +
              (amsRows
                ? '<table class="bf-table"><thead><tr><th>Slot</th><th>Type</th><th>Remain</th><th>Tag</th></tr></thead><tbody>' +
                  amsRows + '</tbody></table>'
                : '<p class="bf-empty m-0">No AMS data</p>') +
              '</div>' +
              '<div><h3 class="font-display text-base mb-2">HMS</h3><div class="flex flex-wrap gap-1">' + hms + '</div></div>' +
              '<div><h3 class="font-display text-base mb-2">Tags</h3><div class="flex flex-wrap gap-1">' + tags + '</div></div>' +
              '<div><h3 class="font-display text-base mb-2">Liveview</h3>' +
              '<p class="bf-card-meta text-xs mb-2">Refreshes every ' + (LIVEVIEW_MS/1000) + 's while open</p>' +
              '<img id="drawer-liveview" data-sn="' + esc(p.external_sn) + '" class="w-full rounded-bf border border-bf-border bg-bf-0" alt="liveview" ' +
              'src="' + keepSrc + '" ' +
              'onerror="this.onerror=null;this.replaceWith(Object.assign(document.createElement(\'p\'),{className:\'bf-empty m-0\',textContent:\'No still\'}))"/>' +
              '</div>';
          }

          document.getElementById('drawer-close').onclick = closeDrawer;
          drawerBackdrop.onclick = closeDrawer;
          document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape' && openSn) closeDrawer();
          });

          function syncBulkBar() {
            const n = selected.size;
            bulkCount.textContent = n + ' selected';
            bulkBar.classList.toggle('hidden', n === 0);
          }

          async function postBulk(action) {
            const serials = [...selected];
            if (!serials.length) return;
            await fetch(API + '/api/v1/printers/actions', {
              method: 'POST',
              credentials: 'include',
              headers: {'Content-Type': 'application/json'},
              body: JSON.stringify({action, serials})
            });
          }

          document.querySelectorAll('#bulk-bar [data-bulk]').forEach(btn => {
            btn.addEventListener('click', () => postBulk(btn.dataset.bulk));
          });
          document.getElementById('bulk-clear').addEventListener('click', () => {
            selected.clear();
            render();
          });

          function render() {
            const items = [...printers.values()].filter(p => {
              if (filter !== 'all' && p.bucket !== filter) return false;
              if (tagFilter) {
                const tags = (p.tags || []).map(t => String(t));
                if (tags.indexOf(tagFilter) < 0) return false;
              }
              if (fleetQ) {
                const hay = ((p.name || '') + ' ' + (p.external_sn || '') + ' ' + (p.model || '')).toLowerCase();
                if (hay.indexOf(fleetQ) < 0) return false;
              }
              return true;
            });
            if (!items.length) {
              grid.innerHTML = '<div class="bf-panel bf-empty">' +
                (printers.size ? 'No printers match this filter.' : 'Waiting for edge deltas (fixture or Farm Manager)…') +
                '</div>';
              syncBulkBar();
              return;
            }
            grid.innerHTML = items.map(p => {
              const sn = p.external_sn;
              const on = selected.has(sn) ? ' checked' : '';
              return '<article class="bf-card" id="printer-' + sn + '">' +
                '<div class="flex items-start gap-2 mb-2">' +
                '<input type="checkbox" class="mt-1 bulk-pick" data-sn="' + sn + '"' + on + '/>' +
                '<button type="button" class="text-left flex-1 min-w-0 printer-open cursor-pointer" data-sn="' + sn + '">' +
                '<span class="bf-card-name">' + esc(p.name || sn) + '</span>' +
                '<span class="bf-card-meta block">' + esc(p.model || '') + ' · <span class="state-' + esc(p.gcode_state||'') + '">' + esc(p.gcode_state||'?') + '</span>' +
                (p.mc_percent ? ' · ' + p.mc_percent + '%' : '') + '</span>' +
                taskLine(p) +
                '</button></div>' +
                amsChips(p) +
                '<div class="flex flex-wrap gap-1.5 mt-3">' +
                '<button type="button" class="bf-btn" data-sn="' + sn + '" data-opt="pause">Pause</button>' +
                '<button type="button" class="bf-btn" data-sn="' + sn + '" data-opt="resume">Resume</button>' +
                '<button type="button" class="bf-btn" data-sn="' + sn + '" data-opt="stop">Stop</button>' +
                '<button type="button" class="bf-btn" data-sn="' + sn + '" data-opt="bed_clean">Bed clean</button>' +
                '</div></article>';
            }).join('');
            grid.querySelectorAll('input.bulk-pick').forEach(cb => {
              cb.addEventListener('change', () => {
                if (cb.checked) selected.add(cb.dataset.sn); else selected.delete(cb.dataset.sn);
                syncBulkBar();
              });
            });
            grid.querySelectorAll('button.printer-open').forEach(btn => {
              btn.addEventListener('click', () => openDrawer(btn.dataset.sn));
            });
            grid.querySelectorAll('button.bf-btn').forEach(btn => {
              btn.addEventListener('click', async () => {
                await fetch(API + '/api/v1/edge/control/' + btn.dataset.sn, {
                  method: 'POST',
                  credentials: 'include',
                  headers: {'Content-Type': 'application/json'},
                  body: JSON.stringify({opt: btn.dataset.opt})
                });
              });
            });
            syncBulkBar();
          }

          async function bootstrap() {
            try {
              const r = await fetch(API + '/api/v1/printers', { credentials: 'include' });
              if (r.ok) {
                const list = await r.json();
                (list || []).forEach(upsert);
              }
              const e = await fetch(API + '/api/v1/edge/status', { credentials: 'include' });
              if (e.ok) {
                const st = await e.json();
                (st.devices || []).forEach(upsert);
              }
            } catch (_) {}
          }

          function connectSSE() {
            const es = new EventSource(API + '/api/v1/events/stream', { withCredentials: true });
            es.addEventListener('ready', () => {
              statusEl.textContent = 'SSE live';
              statusEl.classList.add('live');
            });
            ['printer.state_changed','printer.online','printer.offline','printer.ams_changed'].forEach(ev => {
              es.addEventListener(ev, (msg) => {
                try {
                  const env = JSON.parse(msg.data);
                  upsert(env.payload || {});
                } catch (_) {}
              });
            });
            es.addEventListener('alert.raised', (msg) => {
              try {
                const env = JSON.parse(msg.data);
                const p = env.payload || {};
                const sn = p.external_sn || '?';
                const why = (p.reasons && p.reasons.length) ? p.reasons.join(', ') : (p.gcode_state || 'attention');
                recentAlerts.unshift(esc(p.name || sn) + ' · ' + esc(why));
                if (recentAlerts.length > 5) recentAlerts.pop();
                alertBanner.innerHTML = recentAlerts.map(x => '<div>' + x + '</div>').join('');
                alertBanner.classList.remove('hidden');
                if (p.external_sn) upsert(p);
              } catch (_) {}
            });
            es.onerror = () => {
              statusEl.textContent = 'SSE reconnecting…';
              statusEl.classList.remove('live');
            };
          }

          bootstrap();
          connectSSE();
        </script>
      {{else if eq .Active "artifacts"}}
        <p class="bf-lede">Staged <code>.gcode.3mf</code> plates under <code>ARTIFACTS_DIR</code> (mirrored to object store when configured). Copy an id into Jobs to enqueue.</p>
        <div class="bf-panel mb-4">
          <h2 class="font-display text-lg mb-3">Upload</h2>
          <form id="artifact-form" class="flex flex-wrap items-end gap-3">
            <label class="text-sm">File
              <input class="bf-input block mt-1" id="artifact-file" type="file" accept=".3mf,.gcode.3mf,application/octet-stream" required/>
            </label>
            <label class="text-sm">Est. seconds
              <input class="bf-input block mt-1 w-28" id="artifact-est-sec" type="number" min="0" step="1"/>
            </label>
            <label class="text-sm">Est. grams
              <input class="bf-input block mt-1 w-28" id="artifact-est-g" type="number" min="0" step="0.1"/>
            </label>
            <button type="submit" class="bf-btn bf-btn-primary">Upload</button>
            <p class="text-sm m-0" id="artifact-form-msg"></p>
          </form>
        </div>
        <div class="flex flex-wrap items-center gap-3 mb-3">
          <div id="artifacts-status" class="bf-status m-0">loading…</div>
          <button type="button" class="bf-btn" id="artifacts-refresh">Refresh</button>
        </div>
        <div class="bf-panel overflow-x-auto">
          <table class="bf-table">
            <thead>
              <tr>
                <th>Filename</th>
                <th>ID</th>
                <th>Size</th>
                <th>Est.</th>
                <th>Created</th>
                <th></th>
              </tr>
            </thead>
            <tbody id="artifacts-body"><tr><td colspan="6" class="bf-empty">loading…</td></tr></tbody>
          </table>
        </div>
        <script>
          const API = {{printf "%q" .APIBase}};
          const artStatus = document.getElementById('artifacts-status');
          const artBody = document.getElementById('artifacts-body');
          function esc(s) {
            return String(s == null ? '' : s)
              .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/"/g, '&quot;');
          }
          function fmtBytes(n) {
            const v = Number(n) || 0;
            if (v < 1024) return v + ' B';
            if (v < 1048576) return (v / 1024).toFixed(1) + ' KB';
            return (v / 1048576).toFixed(1) + ' MB';
          }
          async function loadArtifacts() {
            artStatus.textContent = 'loading…';
            artStatus.classList.remove('live');
            try {
              const items = await bfJSON(API + '/api/v1/artifacts');
              if (!items || !items.length) {
                artBody.innerHTML = '<tr><td colspan="6" class="bf-empty">No artifacts yet — upload a plate above.</td></tr>';
                artStatus.textContent = 'empty';
                return;
              }
              artBody.innerHTML = items.map(a => {
                const est = [];
                if (a.est_seconds) est.push(a.est_seconds + 's');
                if (a.est_grams) est.push(a.est_grams + 'g');
                return '<tr>' +
                  '<td>' + esc(a.filename || '—') + '</td>' +
                  '<td class="font-mono text-xs">' + esc(a.id) + '</td>' +
                  '<td class="font-mono text-xs">' + esc(fmtBytes(a.bytes)) + '</td>' +
                  '<td class="text-bf-muted text-xs">' + esc(est.join(' · ') || '—') + '</td>' +
                  '<td class="font-mono text-xs">' + esc(a.created_at ? new Date(a.created_at).toLocaleString() : '—') + '</td>' +
                  '<td class="whitespace-nowrap">' +
                  '<a class="bf-btn" href="' + API + '/api/v1/artifacts/' + encodeURIComponent(a.id) + '/content">Download</a> ' +
                  '<button type="button" class="bf-btn" data-copy="' + esc(a.id) + '">Copy id</button> ' +
                  '<button type="button" class="bf-btn" data-del="' + esc(a.id) + '" data-del-name="' + esc(a.filename || a.id) + '">Delete</button>' +
                  '</td></tr>';
              }).join('');
              artBody.querySelectorAll('[data-copy]').forEach(btn => {
                btn.onclick = async () => {
                  try {
                    await navigator.clipboard.writeText(btn.dataset.copy);
                    artStatus.textContent = 'Copied ' + btn.dataset.copy;
                    artStatus.classList.add('live');
                  } catch (_) {
                    artStatus.textContent = btn.dataset.copy;
                  }
                };
              });
              artBody.querySelectorAll('[data-del]').forEach(btn => {
                btn.onclick = async () => {
                  if (!confirm('Delete artifact ' + (btn.dataset.delName || btn.dataset.del) + '?')) return;
                  try {
                    const r = await fetch(API + '/api/v1/artifacts/' + encodeURIComponent(btn.dataset.del), {
                      method: 'DELETE', credentials: 'include',
                    });
                    if (!r.ok) {
                      const t = await r.text();
                      throw new Error(t || r.statusText);
                    }
                    await loadArtifacts();
                  } catch (e) {
                    artStatus.textContent = String(e.message || e);
                    artStatus.classList.remove('live');
                  }
                };
              });
              artStatus.textContent = 'updated · ' + items.length;
              artStatus.classList.add('live');
            } catch (e) {
              artStatus.textContent = String(e.message || e);
              artBody.innerHTML = '<tr><td colspan="6" class="text-bf-danger">' + esc(e.message || e) + '</td></tr>';
            }
          }
          document.getElementById('artifacts-refresh').onclick = loadArtifacts;
          document.getElementById('artifact-form').onsubmit = async (ev) => {
            ev.preventDefault();
            const msg = document.getElementById('artifact-form-msg');
            msg.textContent = '';
            const file = document.getElementById('artifact-file').files[0];
            if (!file) return;
            const fd = new FormData();
            fd.append('file', file);
            const estSec = parseInt(document.getElementById('artifact-est-sec').value, 10);
            const estG = parseFloat(document.getElementById('artifact-est-g').value);
            if (Number.isFinite(estSec) && estSec > 0) fd.append('est_seconds', String(estSec));
            if (Number.isFinite(estG) && estG > 0) fd.append('est_grams', String(estG));
            try {
              const art = await bfJSON(API + '/api/v1/artifacts', {method: 'POST', credentials: 'include', body: fd});
              msg.textContent = 'Uploaded ' + (art.id || art.filename);
              document.getElementById('artifact-file').value = '';
              await loadArtifacts();
            } catch (e) {
              msg.textContent = String(e.message || e);
            }
          };
          loadArtifacts();
        </script>
      {{else if eq .Active "jobs"}}
        <p class="bf-lede">Upload a <code>.gcode.3mf</code>, enqueue against core. Target a printer SN when known; leave blank to queue unassigned.</p>
        <div class="bf-panel mb-4">
          <h2 class="font-display text-lg mb-3">Create job</h2>
          <form id="job-form" class="grid gap-x-4 gap-y-1 md:grid-cols-2">
            <div>
              <label class="bf-field" for="job-name">Name</label>
              <input class="bf-input" id="job-name" name="name" placeholder="batch-plate-a" required/>
            </div>
            <div>
              <label class="bf-field" for="job-file">Plate file</label>
              <input class="bf-input" id="job-file" name="file" type="file" accept=".3mf,.gcode.3mf,application/octet-stream"/>
            </div>
            <div>
              <label class="bf-field" for="job-artifact-pick">Staged artifact</label>
              <select class="bf-input" id="job-artifact-pick">
                <option value="">— pick or type id below —</option>
              </select>
            </div>
            <div>
              <label class="bf-field" for="job-artifact">Artifact ID <span class="text-bf-muted">(or upload above)</span></label>
              <input class="bf-input" id="job-artifact" name="artifact_id" placeholder="from prior upload"/>
            </div>
            <div>
              <label class="bf-field" for="job-file-id">FM file id <span class="text-bf-muted">(optional f3mf_id)</span></label>
              <input class="bf-input" id="job-file-id" name="file_id" placeholder="skip upload when set"/>
            </div>
            <div>
              <label class="bf-field" for="job-sn">Printer SN <span class="text-bf-muted">(external_sn)</span></label>
              <input class="bf-input" id="job-sn" name="external_sn" placeholder="optional — triggers start when set"/>
            </div>
            <div>
              <label class="bf-field" for="job-model">Target model <span class="text-bf-muted">(optional)</span></label>
              <input class="bf-input" id="job-model" name="target_model" placeholder="P1S, X1C, …"/>
            </div>
            <div>
              <label class="bf-field" for="job-after">After job <span class="text-bf-muted">(wait until it succeeds)</span></label>
              <input class="bf-input" id="job-after" name="after_job_id" placeholder="job id"/>
            </div>
            <div>
              <label class="bf-field" for="job-spool">Spoolman spool <span class="text-bf-muted">(optional id)</span></label>
              <input class="bf-input" id="job-spool" name="spool_id" placeholder="posted on finish when set"/>
            </div>
            <div>
              <label class="bf-field" for="job-copies">Copies</label>
              <input class="bf-input" id="job-copies" name="copies" type="number" min="1" value="1"/>
            </div>
            <div>
              <label class="bf-field" for="job-est-sec">Est. seconds</label>
              <input class="bf-input" id="job-est-sec" name="est_seconds" type="number" min="0" placeholder="0"/>
            </div>
            <div>
              <label class="bf-field" for="job-est-g">Est. grams</label>
              <input class="bf-input" id="job-est-g" name="est_grams" type="number" min="0" step="0.1" placeholder="0"/>
            </div>
            <div class="md:col-span-2 bf-actions">
              <button type="submit" class="bf-btn bf-btn-primary">Create job</button>
              <p class="text-sm self-center" id="job-form-msg"></p>
            </div>
          </form>
        </div>
        <div class="flex flex-wrap items-center gap-3 mb-3">
          <div id="jobs-status" class="bf-status m-0">loading…</div>
          <input class="bf-input max-w-xs" id="jobs-q" type="search" placeholder="Search name, SN, artifact, id"/>
          <select class="bf-input max-w-xs" id="jobs-state">
            <option value="">All states</option>
            <option>queued</option>
            <option>assigned</option>
            <option>dispatched</option>
            <option>running</option>
            <option>paused</option>
            <option>succeeded</option>
            <option>failed</option>
            <option>cancelled</option>
          </select>
          <button type="button" class="bf-btn" id="jobs-refresh">Refresh</button>
        </div>
        <div id="job-events-panel" class="bf-panel mb-3 hidden">
          <div class="flex items-center justify-between gap-2 mb-2">
            <h2 class="font-display text-base m-0">Job history <span class="text-bf-muted font-mono text-xs" id="job-events-id"></span></h2>
            <button type="button" class="bf-btn" id="job-events-close">Close</button>
          </div>
          <div class="overflow-x-auto">
            <table class="bf-table" id="job-events-table">
              <thead><tr><th>When</th><th>From</th><th>To</th><th>Actor</th><th>Reason</th></tr></thead>
              <tbody id="job-events-body"><tr><td colspan="5" class="bf-empty">—</td></tr></tbody>
            </table>
          </div>
        </div>
        <div class="bf-panel overflow-x-auto">
          <table class="bf-table" id="jobs-table">
            <thead>
              <tr>
                <th>Name</th>
                <th>State</th>
                <th>Printer</th>
                <th>Artifact</th>
                <th>Est.</th>
                <th>Cost</th>
                <th>Error</th>
                <th></th>
              </tr>
            </thead>
            <tbody><tr><td colspan="8" class="bf-empty">loading…</td></tr></tbody>
          </table>
        </div>
        <script>
          const API = {{printf "%q" .APIBase}};
          const jobsStatus = document.getElementById('jobs-status');
          const jobsTb = document.querySelector('#jobs-table tbody');

          function esc(s) {
            return String(s == null ? '' : s)
              .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
          }
          function fmtCost(j) {
            const cur = j.currency || '';
            const est = Number(j.cost_estimate) || 0;
            const act = Number(j.cost_actual) || 0;
            if (!est && !act) return '—';
            if (act) return act.toFixed(2) + (cur ? ' ' + cur : '') + ' actual';
            return est.toFixed(2) + (cur ? ' ' + cur : '') + ' est';
          }
          function fmtEst(j) {
            const parts = [];
            if (j.est_seconds) parts.push(Math.round(j.est_seconds / 60) + 'm');
            if (j.est_grams) parts.push(j.est_grams + 'g');
            if (j.copies && j.copies > 1) parts.push('×' + j.copies);
            return parts.length ? parts.join(' / ') : '—';
          }
          function stateClass(st) {
            const s = String(st || '').toLowerCase();
            if (s === 'running' || s === 'dispatched') return 'state-RUNNING';
            if (s === 'failed' || s === 'cancelled') return 'state-FAILED';
            if (s === 'succeeded') return 'state-FINISH';
            if (s === 'paused') return 'state-PAUSE';
            return 'state-IDLE';
          }

          function jobsURL() {
            const p = new URLSearchParams();
            const q = document.getElementById('jobs-q').value.trim();
            const state = document.getElementById('jobs-state').value;
            if (q) p.set('q', q);
            if (state) p.set('state', state);
            const s = p.toString();
            return API + '/api/v1/jobs' + (s ? '?' + s : '');
          }
          async function loadJobs() {
            try {
              const jobs = await bfJSON(jobsURL());
              jobsStatus.textContent = 'Updated ' + new Date().toLocaleTimeString();
              jobsStatus.classList.add('live');
              if (!jobs || !jobs.length) {
                jobsTb.innerHTML = '<tr><td colspan="8" class="bf-empty">No jobs yet — create one above.</td></tr>';
                return;
              }
              jobsTb.innerHTML = jobs.map(j => '<tr>' +
                '<td><div class="font-medium">' + esc(j.name || j.id) + '</div>' +
                  '<div class="text-bf-muted text-xs font-mono">' + esc(j.id) + '</div></td>' +
                '<td><span class="' + stateClass(j.state) + '">' + esc(j.state || '') + '</span></td>' +
                '<td class="text-bf-muted font-mono text-xs">' + esc(j.external_sn || '—') + '</td>' +
                '<td class="text-bf-muted font-mono text-xs">' + esc(j.artifact_id || '—') + '</td>' +
                '<td class="text-bf-muted">' + esc(fmtEst(j)) + '</td>' +
                '<td class="text-bf-muted">' + esc(fmtCost(j)) + '</td>' +
                '<td class="' + (j.error ? 'text-bf-danger text-xs' : 'text-bf-muted') + '">' +
                  esc(j.error || '—') + '</td>' +
                '<td><button type="button" class="bf-btn" data-history="' + esc(j.id) + '">History</button> ' +
                  '<button type="button" class="bf-btn" data-reprint="' + esc(j.id) + '">Reprint</button>' +
                  (['queued','assigned'].indexOf(j.state) >= 0
                    ? ' <button type="button" class="bf-btn" data-dispatch="' + esc(j.id) + '" data-sn="' + esc(j.external_sn || '') + '">Dispatch</button>'
                    : '') +
                  (['queued','assigned','dispatched','running','paused'].indexOf(j.state) >= 0
                    ? ' <button type="button" class="bf-btn" data-cancel="' + esc(j.id) + '">Cancel</button>'
                    : '') +
                  (j.state === 'queued'
                    ? ' <button type="button" class="bf-btn" data-queue="earlier" data-job="' + esc(j.id) + '">Earlier</button>' +
                      ' <button type="button" class="bf-btn" data-queue="later" data-job="' + esc(j.id) + '">Later</button>'
                    : '') +
                '</td>' +
              '</tr>').join('');
            } catch (e) {
              jobsStatus.textContent = 'Error';
              jobsStatus.classList.remove('live');
              jobsTb.innerHTML = '<tr><td colspan="8" class="text-bf-danger">' + esc(e.message || e) + '</td></tr>';
            }
          }

          const eventsPanel = document.getElementById('job-events-panel');
          const eventsBody = document.getElementById('job-events-body');
          const eventsId = document.getElementById('job-events-id');
          async function loadJobEvents(id) {
            eventsId.textContent = id;
            eventsPanel.classList.remove('hidden');
            eventsBody.innerHTML = '<tr><td colspan="5" class="bf-empty">loading…</td></tr>';
            try {
              const items = await bfJSON(API + '/api/v1/jobs/' + encodeURIComponent(id) + '/events');
              if (!items || !items.length) {
                eventsBody.innerHTML = '<tr><td colspan="5" class="bf-empty">No transitions yet</td></tr>';
                return;
              }
              eventsBody.innerHTML = items.map(e => '<tr>' +
                '<td class="font-mono text-xs">' + esc(e.at ? new Date(e.at).toLocaleString() : '—') + '</td>' +
                '<td class="font-mono text-xs">' + esc(e.from_state || '—') + '</td>' +
                '<td class="font-mono text-xs">' + esc(e.to_state || '—') + '</td>' +
                '<td>' + esc(e.actor || '—') + '</td>' +
                '<td class="text-bf-muted text-xs">' + esc(e.reason || '—') + '</td>' +
                '</tr>').join('');
            } catch (err) {
              eventsBody.innerHTML = '<tr><td colspan="5" class="text-bf-danger">' + esc(err.message || err) + '</td></tr>';
            }
          }
          document.getElementById('job-events-close').onclick = () => {
            eventsPanel.classList.add('hidden');
          };

          async function loadJobArtifacts() {
            const sel = document.getElementById('job-artifact-pick');
            try {
              const items = await bfJSON(API + '/api/v1/artifacts');
              const opts = ['<option value="">— pick or type id below —</option>'];
              (items || []).forEach(a => {
                if (!a.id) return;
                const label = (a.filename || a.id) + ' · ' + String(a.id).slice(0, 8);
                opts.push('<option value="' + esc(a.id) + '">' + esc(label) + '</option>');
              });
              sel.innerHTML = opts.join('');
            } catch (_) {}
          }
          document.getElementById('job-artifact-pick').onchange = () => {
            const sel = document.getElementById('job-artifact-pick');
            if (sel.value) document.getElementById('job-artifact').value = sel.value;
          };
          document.getElementById('jobs-refresh').onclick = () => { loadJobs(); loadJobArtifacts(); };
          document.getElementById('jobs-q').oninput = () => { loadJobs(); };
          document.getElementById('jobs-state').onchange = () => { loadJobs(); };
          jobsTb.onclick = async (ev) => {
            const hist = ev.target.closest('[data-history]');
            if (hist) {
              await loadJobEvents(hist.dataset.history);
              return;
            }
            const move = ev.target.closest('[data-queue]');
            if (move && !move.disabled) {
              move.disabled = true;
              try {
                await bfJSON(API + '/api/v1/jobs/' + encodeURIComponent(move.dataset.job) + '/queue', {
                  method: 'POST', credentials: 'include',
                  headers: {'Content-Type': 'application/json'},
                  body: JSON.stringify({direction: move.dataset.queue}),
                });
                await loadJobs();
              } catch (e) {
                jobsStatus.textContent = String(e.message || e);
                jobsStatus.classList.remove('live');
                move.disabled = false;
              }
              return;
            }
            const dispatchBtn = ev.target.closest('[data-dispatch]');
            if (dispatchBtn && !dispatchBtn.disabled) {
              dispatchBtn.disabled = true;
              try {
                const body = {};
                if (dispatchBtn.dataset.sn) body.external_sn = dispatchBtn.dataset.sn;
                await bfJSON(API + '/api/v1/jobs/' + encodeURIComponent(dispatchBtn.dataset.dispatch) + '/dispatch', {
                  method: 'POST', credentials: 'include',
                  headers: {'Content-Type': 'application/json'},
                  body: JSON.stringify(body),
                });
                jobsStatus.textContent = 'Dispatched';
                await loadJobs();
              } catch (e) {
                jobsStatus.textContent = String(e.message || e);
                jobsStatus.classList.remove('live');
                dispatchBtn.disabled = false;
              }
              return;
            }
            const cancelBtn = ev.target.closest('[data-cancel]');
            if (cancelBtn && !cancelBtn.disabled) {
              cancelBtn.disabled = true;
              try {
                await bfJSON(API + '/api/v1/jobs/' + encodeURIComponent(cancelBtn.dataset.cancel) + '/cancel', {
                  method: 'POST', credentials: 'include',
                  headers: {'Content-Type': 'application/json'},
                  body: JSON.stringify({reason: 'operator cancel'}),
                });
                jobsStatus.textContent = 'Cancelled';
                await loadJobs();
              } catch (e) {
                jobsStatus.textContent = String(e.message || e);
                jobsStatus.classList.remove('live');
                cancelBtn.disabled = false;
              }
              return;
            }
            const btn = ev.target.closest('[data-reprint]');
            if (!btn || btn.disabled) return;
            btn.disabled = true;
            try {
              const job = await bfJSON(API + '/api/v1/jobs/' + encodeURIComponent(btn.dataset.reprint) + '/reprint', {
                method: 'POST', credentials: 'include',
              });
              jobsStatus.textContent = 'Queued ' + (job.name || job.id);
              await loadJobs();
            } catch (e) {
              jobsStatus.textContent = String(e.message || e);
              jobsStatus.classList.remove('live');
              btn.disabled = false;
            }
          };
          document.getElementById('job-form').onsubmit = async (ev) => {
            ev.preventDefault();
            const msg = document.getElementById('job-form-msg');
            msg.className = 'text-sm self-center';
            msg.textContent = '';
            const copies = parseInt(document.getElementById('job-copies').value, 10);
            const estSec = parseInt(document.getElementById('job-est-sec').value, 10);
            const estG = parseFloat(document.getElementById('job-est-g').value);
            let artifactId = document.getElementById('job-artifact').value.trim();
            const fileId = document.getElementById('job-file-id').value.trim();
            const fileInput = document.getElementById('job-file');
            try {
              if (fileInput.files && fileInput.files[0]) {
                const fd = new FormData();
                fd.append('file', fileInput.files[0]);
                if (Number.isFinite(estSec) && estSec > 0) fd.append('est_seconds', String(estSec));
                if (Number.isFinite(estG) && estG > 0) fd.append('est_grams', String(estG));
                const art = await bfJSON(API + '/api/v1/artifacts', {
                  method: 'POST', credentials: 'include', body: fd,
                });
                artifactId = art.id;
              }
              const body = {
                name: document.getElementById('job-name').value.trim(),
                artifact_id: artifactId || undefined,
                file_id: fileId || undefined,
                external_sn: document.getElementById('job-sn').value.trim() || undefined,
                target_model: document.getElementById('job-model').value.trim() || undefined,
                after_job_id: document.getElementById('job-after').value.trim() || undefined,
                spool_id: document.getElementById('job-spool').value.trim() || undefined,
                copies: Number.isFinite(copies) && copies > 0 ? copies : 1,
              };
              if (Number.isFinite(estSec) && estSec > 0) body.est_seconds = estSec;
              if (Number.isFinite(estG) && estG > 0) body.est_grams = estG;
              const job = await bfJSON(API + '/api/v1/jobs', {
                method: 'POST',
                credentials: 'include',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify(body),
              });
              msg.className = 'text-bf-accent text-sm self-center';
              msg.textContent = 'Queued ' + (job.name || job.id);
              document.getElementById('job-form').reset();
              document.getElementById('job-copies').value = '1';
              await loadJobs();
            } catch (e) {
              msg.className = 'text-bf-danger text-sm self-center';
              msg.textContent = String(e.message || e);
            }
          };

          let jobsReloadTimer = null;
          function scheduleJobsReload() {
            if (jobsReloadTimer) clearTimeout(jobsReloadTimer);
            jobsReloadTimer = setTimeout(() => { jobsReloadTimer = null; loadJobs(); }, 200);
          }
          function connectJobsSSE() {
            const es = new EventSource(API + '/api/v1/events/stream', { withCredentials: true });
            es.addEventListener('ready', () => {
              jobsStatus.textContent = 'SSE live';
              jobsStatus.classList.add('live');
            });
            ['job.queued','job.assigned','job.dispatched','job.progress','job.finished','job.failed','job.cancelled'].forEach(ev => {
              es.addEventListener(ev, () => { scheduleJobsReload(); });
            });
            es.onerror = () => {
              jobsStatus.textContent = 'SSE reconnecting…';
              jobsStatus.classList.remove('live');
            };
          }

          loadJobs();
          loadJobArtifacts();
          connectJobsSSE();
        </script>
      {{else if eq .Active "presets"}}
        <p class="bf-lede">Versioned filament and print presets stored in Postgres.</p>
        <div class="bf-panel mb-4">
          <h2 class="font-display text-lg mb-3" id="preset-editor-title">New filament preset</h2>
          <form id="preset-editor" class="grid gap-x-4 gap-y-1 md:grid-cols-2">
            <input type="hidden" id="preset-kind" value="filament"/>
            <input type="hidden" id="preset-mode" value="create"/>
            <div>
              <label class="bf-field" for="preset-id">ID</label>
              <input class="bf-input" id="preset-id" name="id" required placeholder="pla-basic"/>
            </div>
            <div>
              <label class="bf-field" for="preset-name">Name</label>
              <input class="bf-input" id="preset-name" name="name" required placeholder="PLA Basic"/>
            </div>
            <div data-filament-only>
              <label class="bf-field" for="preset-material">Material</label>
              <input class="bf-input" id="preset-material" placeholder="PLA"/>
            </div>
            <div data-filament-only>
              <label class="bf-field" for="preset-vendor">Vendor</label>
              <input class="bf-input" id="preset-vendor" placeholder="Generic"/>
            </div>
            <div data-filament-only>
              <label class="bf-field" for="preset-price">Price $/kg</label>
              <input class="bf-input" id="preset-price" type="number" min="0" step="0.01" placeholder="0"/>
            </div>
            <div data-print-only class="hidden">
              <label class="bf-field" for="preset-layer">Layer height mm</label>
              <input class="bf-input" id="preset-layer" type="number" min="0" step="0.01" placeholder="0.2"/>
            </div>
            <div data-print-only class="hidden">
              <label class="bf-field" for="preset-nozzle">Nozzle ø mm</label>
              <input class="bf-input" id="preset-nozzle" type="number" min="0" step="0.01" placeholder="0.4"/>
            </div>
            <div class="md:col-span-2 flex flex-wrap gap-2 items-center mt-2">
              <button type="submit" class="bf-btn bf-btn-primary" id="preset-save">Create</button>
              <button type="button" class="bf-btn" id="preset-new-filament">New filament</button>
              <button type="button" class="bf-btn" id="preset-new-print">New print</button>
              <button type="button" class="bf-btn hidden" id="preset-cancel">Cancel edit</button>
              <p class="text-sm self-center" id="preset-editor-msg"></p>
            </div>
          </form>
        </div>
        <div class="bf-panel mb-4">
          <h2 class="font-display text-lg mb-3">Import / export</h2>
          <p class="text-sm text-bf-muted mb-3">JSON bundle shape: <code>filament_presets</code> + <code>print_presets</code>. Import upserts by id and bumps version.</p>
          <div class="flex flex-wrap gap-2 items-center mb-3">
            <a class="bf-btn" id="presets-export" href="#">Export JSON</a>
            <label class="bf-btn cursor-pointer">
              Import file
              <input type="file" id="presets-file" accept="application/json,.json" class="hidden"/>
            </label>
          </div>
          <label class="bf-field" for="presets-import">Or paste bundle JSON</label>
          <textarea class="bf-input font-mono text-xs min-h-28" id="presets-import" placeholder='{"filament_presets":[],"print_presets":[]}'></textarea>
          <button type="button" class="bf-btn bf-btn-primary mt-2" id="presets-import-btn">Import</button>
          <p class="text-sm mt-2" id="presets-msg"></p>
        </div>
        <div class="grid gap-4 md:grid-cols-2">
          <div class="bf-panel">
            <h2 class="font-display text-lg mb-3">Filament</h2>
            <div id="filament-list" class="bf-empty">loading…</div>
          </div>
          <div class="bf-panel">
            <h2 class="font-display text-lg mb-3">Print</h2>
            <div id="print-list" class="bf-empty">loading…</div>
          </div>
        </div>
        <div class="bf-panel mt-4 hidden" id="presets-history">
          <div class="flex items-center justify-between gap-2 mb-3">
            <h2 class="font-display text-lg m-0" id="presets-history-title">History</h2>
            <button type="button" class="bf-btn" id="presets-history-close">Close</button>
          </div>
          <div id="presets-history-body" class="bf-empty">loading…</div>
          <p class="text-sm mt-2" id="presets-history-msg"></p>
        </div>
        <script>
          const API = {{printf "%q" .APIBase}};
          document.getElementById('presets-export').href = API + '/api/v1/presets/export';
          const historyPanel = document.getElementById('presets-history');
          let historyKind = '';
          let historyID = '';
          function list(el, items, map) {
            if (!items || !items.length) { el.innerHTML = '<p class="bf-empty">None yet</p>'; return; }
            el.innerHTML = '<ul class="space-y-2">' + items.map(map).join('') + '</ul>';
          }
          function setEditorKind(kind, mode) {
            document.getElementById('preset-kind').value = kind;
            document.getElementById('preset-mode').value = mode || 'create';
            const editing = (mode || 'create') === 'edit';
            document.getElementById('preset-id').readOnly = editing;
            document.getElementById('preset-save').textContent = editing ? 'Save' : 'Create';
            document.getElementById('preset-cancel').classList.toggle('hidden', !editing);
            document.getElementById('preset-editor-title').textContent =
              (editing ? 'Edit ' : 'New ') + (kind === 'print' ? 'print' : 'filament') + ' preset';
            document.querySelectorAll('[data-filament-only]').forEach(el => el.classList.toggle('hidden', kind !== 'filament'));
            document.querySelectorAll('[data-print-only]').forEach(el => el.classList.toggle('hidden', kind !== 'print'));
          }
          function resetEditor(kind) {
            document.getElementById('preset-editor').reset();
            document.getElementById('preset-editor-msg').textContent = '';
            setEditorKind(kind || 'filament', 'create');
          }
          async function editPreset(kind, id) {
            try {
              const x = await bfJSON(API + '/api/v1/presets/' + kind + '/' + encodeURIComponent(id));
              setEditorKind(kind, 'edit');
              document.getElementById('preset-id').value = x.id || id;
              document.getElementById('preset-name').value = x.name || '';
              document.getElementById('preset-material').value = x.material || '';
              document.getElementById('preset-vendor').value = x.vendor || '';
              document.getElementById('preset-price').value = x.price_per_kg || '';
              document.getElementById('preset-layer').value = x.layer_height_mm || '';
              document.getElementById('preset-nozzle').value = x.nozzle_diameter_mm || '';
              document.getElementById('preset-editor-msg').textContent = '';
              document.getElementById('preset-editor').scrollIntoView({behavior: 'smooth', block: 'nearest'});
            } catch (e) {
              alert(String(e.message || e));
            }
          }
          function renderPresets(f, p) {
            list(document.getElementById('filament-list'), f, x =>
              '<li class="bf-card"><div class="bf-card-name">' + (x.name||x.id) + '</div>' +
              '<div class="bf-card-meta">' + (x.material||'') + ' · ' + (x.vendor||'') +
              ' · v' + (x.version||1) +
              (x.price_per_kg ? ' · $' + x.price_per_kg + '/kg' : '') + '</div>' +
              '<div class="flex flex-wrap gap-2 mt-2">' +
              '<button type="button" class="bf-btn" data-edit-kind="filament" data-edit-id="' + x.id + '">Edit</button>' +
              '<button type="button" class="bf-btn" data-hist-kind="filament" data-hist-id="' + x.id + '" data-hist-name="' + (x.name||x.id) + '">History</button>' +
              '<button type="button" class="bf-btn" data-del-kind="filament" data-del-id="' + x.id + '" data-del-name="' + (x.name||x.id) + '">Delete</button></div></li>');
            list(document.getElementById('print-list'), p, x =>
              '<li class="bf-card"><div class="bf-card-name">' + (x.name||x.id) + '</div>' +
              '<div class="bf-card-meta">' + (x.layer_height_mm||'') + 'mm · ø' + (x.nozzle_diameter_mm||'') +
              ' · v' + (x.version||1) + '</div>' +
              '<div class="flex flex-wrap gap-2 mt-2">' +
              '<button type="button" class="bf-btn" data-edit-kind="print" data-edit-id="' + x.id + '">Edit</button>' +
              '<button type="button" class="bf-btn" data-hist-kind="print" data-hist-id="' + x.id + '" data-hist-name="' + (x.name||x.id) + '">History</button>' +
              '<button type="button" class="bf-btn" data-del-kind="print" data-del-id="' + x.id + '" data-del-name="' + (x.name||x.id) + '">Delete</button></div></li>');
            document.querySelectorAll('[data-edit-id]').forEach(btn => {
              btn.onclick = () => editPreset(btn.dataset.editKind, btn.dataset.editId);
            });
            document.querySelectorAll('[data-hist-id]').forEach(btn => {
              btn.onclick = () => openHistory(btn.dataset.histKind, btn.dataset.histId, btn.dataset.histName);
            });
            document.querySelectorAll('[data-del-id]').forEach(btn => {
              btn.onclick = () => deletePreset(btn.dataset.delKind, btn.dataset.delId, btn.dataset.delName);
            });
          }
          async function deletePreset(kind, id, name) {
            if (!confirm('Delete preset ' + (name || id) + '? Revisions are removed too.')) return;
            try {
              const r = await fetch(API + '/api/v1/presets/' + kind + '/' + encodeURIComponent(id), {
                method: 'DELETE', credentials: 'include',
              });
              if (!r.ok) {
                const t = await r.text();
                throw new Error(t || r.statusText);
              }
              await loadPresets();
            } catch (e) {
              alert(String(e.message || e));
            }
          }
          async function openHistory(kind, id, name) {
            historyKind = kind;
            historyID = id;
            document.getElementById('presets-history-title').textContent = 'History · ' + (name || id);
            document.getElementById('presets-history-msg').textContent = '';
            historyPanel.classList.remove('hidden');
            const body = document.getElementById('presets-history-body');
            body.textContent = 'loading…';
            try {
              const rows = await bfJSON(API + '/api/v1/presets/' + kind + '/' + encodeURIComponent(id) + '/revisions');
              if (!rows || !rows.length) { body.innerHTML = '<p class="bf-empty">No revisions</p>'; return; }
              body.innerHTML = '<ul class="space-y-2">' + rows.map(r =>
                '<li class="bf-card flex flex-wrap items-center justify-between gap-2">' +
                '<span class="font-mono text-xs">v' + r.version + ' · ' + (r.created_at || '') + '</span>' +
                '<button type="button" class="bf-btn" data-restore="' + r.version + '">Restore</button></li>'
              ).join('') + '</ul>';
              body.querySelectorAll('[data-restore]').forEach(btn => {
                btn.onclick = () => restoreRevision(Number(btn.dataset.restore));
              });
            } catch (e) {
              body.textContent = String(e.message || e);
            }
          }
          async function restoreRevision(version) {
            const msg = document.getElementById('presets-history-msg');
            msg.className = 'text-sm mt-2';
            msg.textContent = '';
            try {
              await bfJSON(API + '/api/v1/presets/' + historyKind + '/' + encodeURIComponent(historyID) + '/restore', {
                method: 'POST', credentials: 'include',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({version}),
              });
              msg.className = 'text-bf-accent text-sm mt-2';
              msg.textContent = 'Restored v' + version + ' as a new tip';
              await loadPresets();
              await openHistory(historyKind, historyID, document.getElementById('presets-history-title').textContent.replace(/^History · /, ''));
            } catch (e) {
              msg.className = 'text-bf-danger text-sm mt-2';
              msg.textContent = String(e.message || e);
            }
          }
          document.getElementById('presets-history-close').onclick = () => historyPanel.classList.add('hidden');
          document.getElementById('preset-new-filament').onclick = () => resetEditor('filament');
          document.getElementById('preset-new-print').onclick = () => resetEditor('print');
          document.getElementById('preset-cancel').onclick = () => resetEditor(document.getElementById('preset-kind').value);
          document.getElementById('preset-editor').onsubmit = async (ev) => {
            ev.preventDefault();
            const msg = document.getElementById('preset-editor-msg');
            msg.className = 'text-sm self-center';
            msg.textContent = '';
            const kind = document.getElementById('preset-kind').value;
            const mode = document.getElementById('preset-mode').value;
            const id = document.getElementById('preset-id').value.trim();
            const name = document.getElementById('preset-name').value.trim();
            if (!id || !name) {
              msg.className = 'text-bf-danger text-sm self-center';
              msg.textContent = 'id and name required';
              return;
            }
            let body;
            if (kind === 'print') {
              body = {
                id, name,
                layer_height_mm: Number(document.getElementById('preset-layer').value) || 0,
                nozzle_diameter_mm: Number(document.getElementById('preset-nozzle').value) || 0,
                active: true,
              };
            } else {
              body = {
                id, name,
                material: document.getElementById('preset-material').value.trim(),
                vendor: document.getElementById('preset-vendor').value.trim(),
                price_per_kg: Number(document.getElementById('preset-price').value) || 0,
                active: true,
              };
            }
            const url = mode === 'edit'
              ? API + '/api/v1/presets/' + kind + '/' + encodeURIComponent(id)
              : API + '/api/v1/presets/' + kind;
            try {
              await bfJSON(url, {
                method: mode === 'edit' ? 'PATCH' : 'POST',
                credentials: 'include',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify(body),
              });
              msg.className = 'text-bf-accent text-sm self-center';
              msg.textContent = mode === 'edit' ? 'Saved' : 'Created';
              resetEditor(kind);
              await loadPresets();
            } catch (e) {
              msg.className = 'text-bf-danger text-sm self-center';
              msg.textContent = String(e.message || e);
            }
          };
          async function loadPresets() {
            const [f, p] = await Promise.all([
              bfJSON(API + '/api/v1/presets/filament'),
              bfJSON(API + '/api/v1/presets/print')
            ]);
            renderPresets(f, p);
          }
          async function importBundle(text) {
            const msg = document.getElementById('presets-msg');
            msg.className = 'text-sm mt-2';
            msg.textContent = '';
            let bundle;
            try { bundle = JSON.parse(text); } catch (e) {
              msg.className = 'text-bf-danger text-sm mt-2';
              msg.textContent = 'Invalid JSON';
              return;
            }
            try {
              const r = await fetch(API + '/api/v1/presets/import', {
                method: 'POST', credentials: 'include',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify(bundle),
              });
              if (!r.ok) {
                let err = r.statusText;
                try { const j = await r.json(); err = j.error || err; } catch (_) {}
                throw new Error(err);
              }
              msg.className = 'text-bf-accent text-sm mt-2';
              msg.textContent = 'Imported';
              document.getElementById('presets-import').value = '';
              await loadPresets();
            } catch (e) {
              msg.className = 'text-bf-danger text-sm mt-2';
              msg.textContent = String(e.message || e);
            }
          }
          document.getElementById('presets-import-btn').onclick = () =>
            importBundle(document.getElementById('presets-import').value);
          document.getElementById('presets-file').onchange = async (ev) => {
            const file = ev.target.files && ev.target.files[0];
            if (!file) return;
            const text = await file.text();
            document.getElementById('presets-import').value = text;
            await importBundle(text);
            ev.target.value = '';
          };
          setEditorKind('filament', 'create');
          loadPresets().catch(e => {
            document.getElementById('filament-list').textContent = String(e.message || e);
          });
        </script>
      {{else if eq .Active "filament"}}
        <p class="bf-lede">Spoolman spools and AMS tray links when <code>SPOOLMAN_URL</code> is set.</p>
        <div class="flex flex-wrap items-center gap-3 mb-3">
          <div id="filament-status" class="bf-status m-0">loading…</div>
          <button type="button" class="bf-btn" id="filament-refresh">Refresh</button>
        </div>
        <div class="bf-panel mb-4">
          <h2 class="font-display text-lg mb-3">Low stock</h2>
          <div id="filament-low" class="bf-empty">loading…</div>
        </div>
        <div class="bf-panel mb-4">
          <h2 class="font-display text-lg mb-3">Spools</h2>
          <div class="overflow-x-auto">
            <table class="bf-table">
              <thead><tr><th>Name</th><th>Material</th><th>Vendor</th><th>Remain</th><th>Location</th><th>ID</th></tr></thead>
              <tbody id="filament-spools-body"><tr><td colspan="6" class="bf-empty">loading…</td></tr></tbody>
            </table>
          </div>
        </div>
        <div class="bf-panel">
          <h2 class="font-display text-lg mb-3">AMS tray → spool</h2>
          <div class="flex flex-wrap items-end gap-3">
            <label class="text-sm">Tray index
              <input class="bf-input block mt-1" id="filament-tray" type="number" min="0" step="1" value="0"/>
            </label>
            <button type="button" class="bf-btn" id="filament-tray-lookup">Lookup</button>
            <p class="text-sm m-0" id="filament-tray-msg">Uses <code>AMS_SLOT_MAP</code>.</p>
          </div>
        </div>
        <script>
          const API = {{printf "%q" .APIBase}};
          const filStatus = document.getElementById('filament-status');
          const filLow = document.getElementById('filament-low');
          const filBody = document.getElementById('filament-spools-body');
          function esc(s) {
            return String(s == null ? '' : s)
              .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/"/g, '&quot;');
          }
          function fmtG(n) {
            const v = Number(n);
            if (!Number.isFinite(v)) return '—';
            return v.toFixed(0) + ' g';
          }
          async function loadFilament() {
            filStatus.textContent = 'loading…';
            filStatus.classList.remove('live');
            try {
              const data = await bfJSON(API + '/api/v1/filament/spools');
              if (!data.configured) {
                filStatus.textContent = 'Spoolman not configured';
                filLow.innerHTML = '<p class="bf-empty m-0">Set SPOOLMAN_URL to list spools.</p>';
                filBody.innerHTML = '<tr><td colspan="6" class="bf-empty">Not configured</td></tr>';
                return;
              }
              const low = data.low_stock || [];
              if (!low.length) {
                filLow.innerHTML = '<p class="bf-empty m-0">None under ' + esc(data.threshold_g ?? '?') + ' g</p>';
              } else {
                filLow.innerHTML = '<div class="flex flex-wrap gap-2">' + low.map(s =>
                  '<span class="bf-chip">' + esc(s.name || s.id) + ' · ' + esc(fmtG(s.remaining_g)) + '</span>'
                ).join('') + '</div>';
              }
              const spools = data.spools || [];
              if (!spools.length) {
                filBody.innerHTML = '<tr><td colspan="6" class="bf-empty">No spools</td></tr>';
              } else {
                filBody.innerHTML = spools.map(s => '<tr>' +
                  '<td>' + esc(s.name || '—') + '</td>' +
                  '<td>' + esc(s.material || '—') + '</td>' +
                  '<td>' + esc(s.vendor || '—') + '</td>' +
                  '<td class="font-mono text-xs">' + esc(fmtG(s.remaining_g)) + '</td>' +
                  '<td class="text-bf-muted text-xs">' + esc(s.location || '—') + '</td>' +
                  '<td class="font-mono text-xs">' + esc(s.id || s.spoolman_id || '—') + '</td>' +
                  '</tr>').join('');
              }
              filStatus.textContent = 'updated · ' + spools.length + ' spools';
              filStatus.classList.add('live');
            } catch (e) {
              filStatus.textContent = String(e.message || e);
              filLow.textContent = String(e.message || e);
              filBody.innerHTML = '<tr><td colspan="6" class="text-bf-danger">' + esc(e.message || e) + '</td></tr>';
            }
          }
          document.getElementById('filament-refresh').onclick = loadFilament;
          document.getElementById('filament-tray-lookup').onclick = async () => {
            const msg = document.getElementById('filament-tray-msg');
            const tray = document.getElementById('filament-tray').value;
            try {
              const data = await bfJSON(API + '/api/v1/filament/links?tray=' + encodeURIComponent(tray));
              msg.textContent = data.spool_id
                ? ('Tray ' + data.tray + ' → spool ' + data.spool_id)
                : ('Tray ' + data.tray + ' has no mapped spool');
            } catch (e) {
              msg.textContent = String(e.message || e);
            }
          };
          loadFilament();
        </script>
      {{else if eq .Active "library"}}
        <p class="bf-lede">Manyfold models when <code>MANYFOLD_URL</code> is set. Import/export create a model by name (plate bytes stay in farm artifacts).</p>
        <div class="flex flex-wrap items-center gap-3 mb-3">
          <div id="library-status" class="bf-status m-0">loading…</div>
          <button type="button" class="bf-btn" id="library-refresh">Refresh</button>
        </div>
        <div class="bf-panel mb-4">
          <h2 class="font-display text-lg mb-3">Import model</h2>
          <form id="library-import-form" class="flex flex-wrap items-end gap-3">
            <label class="text-sm flex-1 min-w-[12rem]">Name
              <input class="bf-input block mt-1 w-full" id="library-import-name" required placeholder="Benchy"/>
            </label>
            <button type="submit" class="bf-btn bf-btn-primary">Create in Manyfold</button>
            <p class="text-sm m-0" id="library-import-msg"></p>
          </form>
        </div>
        <div class="bf-panel mb-4">
          <h2 class="font-display text-lg mb-3">Export artifact → model</h2>
          <form id="library-export-form" class="grid gap-3 md:grid-cols-2">
            <label class="text-sm">Staged artifact
              <select class="bf-input block mt-1 w-full" id="library-export-pick">
                <option value="">— pick or type id below —</option>
              </select>
            </label>
            <label class="text-sm">Artifact id
              <input class="bf-input block mt-1 w-full" id="library-export-artifact" required placeholder="artifact uuid"/>
            </label>
            <label class="text-sm">Model name
              <input class="bf-input block mt-1 w-full" id="library-export-name" required placeholder="Benchy plate"/>
            </label>
            <div class="md:col-span-2 flex flex-wrap items-center gap-3">
              <button type="submit" class="bf-btn bf-btn-primary">Export</button>
              <p class="text-sm m-0" id="library-export-msg"></p>
            </div>
          </form>
        </div>
        <div class="bf-panel">
          <h2 class="font-display text-lg mb-3">Models</h2>
          <div class="overflow-x-auto">
            <table class="bf-table">
              <thead><tr><th>Name</th><th>ID</th><th>URL</th></tr></thead>
              <tbody id="library-models-body"><tr><td colspan="3" class="bf-empty">loading…</td></tr></tbody>
            </table>
          </div>
        </div>
        <script>
          const API = {{printf "%q" .APIBase}};
          const libStatus = document.getElementById('library-status');
          const libBody = document.getElementById('library-models-body');
          function esc(s) {
            return String(s == null ? '' : s)
              .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/"/g, '&quot;');
          }
          async function loadLibrary() {
            libStatus.textContent = 'loading…';
            libStatus.classList.remove('live');
            try {
              const data = await bfJSON(API + '/api/v1/library/models');
              if (!data.configured) {
                libStatus.textContent = 'Manyfold not configured';
                libBody.innerHTML = '<tr><td colspan="3" class="bf-empty">Set MANYFOLD_URL to list models.</td></tr>';
                return;
              }
              const models = data.models || [];
              if (!models.length) {
                libBody.innerHTML = '<tr><td colspan="3" class="bf-empty">No models</td></tr>';
              } else {
                libBody.innerHTML = models.map(m => '<tr>' +
                  '<td>' + esc(m.name || '—') + '</td>' +
                  '<td class="font-mono text-xs">' + esc(m.id || '—') + '</td>' +
                  '<td class="text-xs">' + (m.url
                    ? '<a class="underline" href="' + esc(m.url) + '" target="_blank" rel="noopener">' + esc(m.url) + '</a>'
                    : '—') + '</td>' +
                  '</tr>').join('');
              }
              libStatus.textContent = 'updated · ' + models.length + ' models';
              libStatus.classList.add('live');
            } catch (e) {
              libStatus.textContent = String(e.message || e);
              libBody.innerHTML = '<tr><td colspan="3" class="text-bf-danger">' + esc(e.message || e) + '</td></tr>';
            }
          }
          async function loadLibraryArtifacts() {
            const sel = document.getElementById('library-export-pick');
            try {
              const items = await bfJSON(API + '/api/v1/artifacts');
              const opts = ['<option value="">— pick or type id below —</option>'];
              (items || []).forEach(a => {
                if (!a.id) return;
                const label = (a.filename || a.id) + ' · ' + String(a.id).slice(0, 8);
                opts.push('<option value="' + esc(a.id) + '" data-name="' + esc(a.filename || '') + '">' + esc(label) + '</option>');
              });
              sel.innerHTML = opts.join('');
            } catch (_) {}
          }
          document.getElementById('library-export-pick').onchange = () => {
            const sel = document.getElementById('library-export-pick');
            if (!sel.value) return;
            document.getElementById('library-export-artifact').value = sel.value;
            const opt = sel.options[sel.selectedIndex];
            const nameEl = document.getElementById('library-export-name');
            if (!nameEl.value.trim() && opt && opt.dataset.name) {
              nameEl.value = opt.dataset.name.replace(/\.(gcode\.)?3mf$/i, '') || opt.dataset.name;
            }
          };
          document.getElementById('library-refresh').onclick = () => { loadLibrary(); loadLibraryArtifacts(); };
          document.getElementById('library-import-form').onsubmit = async (ev) => {
            ev.preventDefault();
            const msg = document.getElementById('library-import-msg');
            msg.textContent = '';
            try {
              const out = await bfJSON(API + '/api/v1/library/import', {
                method: 'POST', credentials: 'include',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({name: document.getElementById('library-import-name').value.trim()}),
              });
              msg.textContent = out.imported ? ('Created ' + (out.id || out.name)) : (out.reason || 'not imported');
              await loadLibrary();
            } catch (e) {
              msg.textContent = String(e.message || e);
            }
          };
          document.getElementById('library-export-form').onsubmit = async (ev) => {
            ev.preventDefault();
            const msg = document.getElementById('library-export-msg');
            msg.textContent = '';
            try {
              const out = await bfJSON(API + '/api/v1/library/export', {
                method: 'POST', credentials: 'include',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({
                  artifact_id: document.getElementById('library-export-artifact').value.trim(),
                  name: document.getElementById('library-export-name').value.trim(),
                }),
              });
              msg.textContent = out.exported ? ('Created ' + (out.id || out.name)) : (out.reason || 'not exported');
              await loadLibrary();
            } catch (e) {
              msg.textContent = String(e.message || e);
            }
          };
          loadLibrary();
          loadLibraryArtifacts();
        </script>
      {{else if eq .Active "slice"}}
        <p class="bf-lede">Optional slicer plugin (<code>compose --profile slicer</code>). Orca is operator-supplied via <code>ORCA_BIN</code> and is not bundled.</p>
        <div class="flex flex-wrap items-center gap-3 mb-3">
          <div id="slice-status" class="bf-status m-0">loading…</div>
          <button type="button" class="bf-btn" id="slice-refresh">Refresh status</button>
        </div>
        <div class="bf-panel mb-4">
          <h2 class="font-display text-lg mb-3">Plugin status</h2>
          <pre class="bf-pre text-xs m-0" id="slice-status-out">…</pre>
        </div>
        <div class="bf-panel">
          <h2 class="font-display text-lg mb-3">Submit slice job</h2>
          <form id="slice-form" class="grid gap-3 md:grid-cols-2">
            <label class="text-sm md:col-span-2">Name
              <input class="bf-input block mt-1 w-full" id="slice-name" required placeholder="Benchy"/>
            </label>
            <label class="text-sm md:col-span-2">Staged artifact
              <select class="bf-input block mt-1 w-full" id="slice-artifact">
                <option value="">— pick or type path below —</option>
              </select>
            </label>
            <label class="text-sm md:col-span-2">Input path
              <input class="bf-input block mt-1 w-full font-mono text-xs" id="slice-input" placeholder="relative to SLICE_WORK_DIR / ARTIFACTS_DIR"/>
            </label>
            <label class="text-sm">Print preset
              <select class="bf-input block mt-1 w-full" id="slice-print-preset">
                <option value="">— optional —</option>
              </select>
            </label>
            <label class="text-sm">Filament preset
              <select class="bf-input block mt-1 w-full" id="slice-filament-preset">
                <option value="">— optional —</option>
              </select>
            </label>
            <label class="text-sm">Output path
              <input class="bf-input block mt-1 w-full font-mono text-xs" id="slice-output" placeholder="optional"/>
            </label>
            <div class="md:col-span-2 flex flex-wrap items-center gap-3">
              <button type="submit" class="bf-btn bf-btn-primary">Slice</button>
              <p class="text-sm m-0" id="slice-form-msg"></p>
            </div>
          </form>
          <pre class="bf-pre text-xs mt-3 hidden" id="slice-result"></pre>
        </div>
        <script>
          const API = {{printf "%q" .APIBase}};
          const sliceStatus = document.getElementById('slice-status');
          const sliceOut = document.getElementById('slice-status-out');
          const sliceResult = document.getElementById('slice-result');
          async function loadSliceStatus() {
            sliceStatus.textContent = 'loading…';
            sliceStatus.classList.remove('live');
            try {
              const st = await bfJSON(API + '/api/v1/slice/status');
              sliceOut.textContent = JSON.stringify(st, null, 2);
              sliceStatus.textContent = st.orca_configured ? 'Orca configured' : 'plugin up · ORCA_BIN unset';
              sliceStatus.classList.add('live');
            } catch (e) {
              sliceStatus.textContent = 'slicer not reachable';
              sliceOut.textContent = String(e.message || e) + '\n\nEnable with: podman compose --profile slicer … and SLICER_URL on gateway.';
            }
          }
          function escOpt(s) {
            return String(s == null ? '' : s)
              .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/"/g, '&quot;');
          }
          async function loadSliceArtifacts() {
            const sel = document.getElementById('slice-artifact');
            try {
              const items = await bfJSON(API + '/api/v1/artifacts');
              const opts = ['<option value="">— pick or type path below —</option>'];
              (items || []).forEach(a => {
                const path = a.object_key || a.filename || a.id || '';
                if (!path) return;
                const label = (a.filename || path) + (a.id ? ' · ' + a.id.slice(0, 8) : '');
                opts.push('<option value="' + escOpt(path) + '" data-name="' + escOpt(a.filename || '') + '">' +
                  escOpt(label) + '</option>');
              });
              sel.innerHTML = opts.join('');
            } catch (_) {
              /* leave default option */
            }
          }
          document.getElementById('slice-artifact').onchange = () => {
            const sel = document.getElementById('slice-artifact');
            const opt = sel.options[sel.selectedIndex];
            if (!sel.value) return;
            document.getElementById('slice-input').value = sel.value;
            const nameEl = document.getElementById('slice-name');
            if (!nameEl.value.trim() && opt && opt.dataset.name) {
              nameEl.value = opt.dataset.name.replace(/\.(gcode\.)?3mf$/i, '') || opt.dataset.name;
            }
          };
          async function loadSlicePresets() {
            const fill = (sel, items) => {
              const opts = ['<option value="">— optional —</option>'];
              (items || []).forEach(p => {
                if (!p.id) return;
                opts.push('<option value="' + escOpt(p.id) + '">' + escOpt((p.name || p.id) + ' · ' + p.id) + '</option>');
              });
              sel.innerHTML = opts.join('');
            };
            try {
              const [print, filament] = await Promise.all([
                bfJSON(API + '/api/v1/presets/print'),
                bfJSON(API + '/api/v1/presets/filament'),
              ]);
              fill(document.getElementById('slice-print-preset'), print);
              fill(document.getElementById('slice-filament-preset'), filament);
            } catch (_) {}
          }
          document.getElementById('slice-refresh').onclick = () => {
            loadSliceStatus();
            loadSliceArtifacts();
            loadSlicePresets();
          };
          document.getElementById('slice-form').onsubmit = async (ev) => {
            ev.preventDefault();
            const msg = document.getElementById('slice-form-msg');
            msg.textContent = '';
            sliceResult.classList.add('hidden');
            const body = {
              name: document.getElementById('slice-name').value.trim(),
              input_path: document.getElementById('slice-input').value.trim() || undefined,
              output_path: document.getElementById('slice-output').value.trim() || undefined,
              print_preset_id: document.getElementById('slice-print-preset').value.trim() || undefined,
              filament_preset_id: document.getElementById('slice-filament-preset').value.trim() || undefined,
            };
            try {
              const out = await bfJSON(API + '/api/v1/slice/jobs', {
                method: 'POST', credentials: 'include',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify(body),
              });
              msg.textContent = out.sliced ? ('Sliced → ' + (out.output_path || 'ok')) : (out.reason || 'not sliced');
              sliceResult.textContent = JSON.stringify(out, null, 2);
              sliceResult.classList.remove('hidden');
            } catch (e) {
              msg.textContent = String(e.message || e);
            }
          };
          loadSliceStatus();
          loadSliceArtifacts();
          loadSlicePresets();
        </script>
      {{else if eq .Active "commerce"}}
        <p class="bf-lede">Optional shop-orders plugin (<code>compose --profile commerce</code>). Webhook queues a core job when <code>CORE_URL</code> is set; pack/ship status is echoed (not stored yet).</p>
        <div class="flex flex-wrap items-center gap-3 mb-3">
          <div id="commerce-status" class="bf-status m-0">loading…</div>
          <button type="button" class="bf-btn" id="commerce-refresh">Refresh status</button>
        </div>
        <div class="bf-panel mb-4">
          <h2 class="font-display text-lg mb-3">Plugin status</h2>
          <pre class="bf-pre text-xs m-0" id="commerce-status-out">…</pre>
        </div>
        <div class="bf-panel mb-4">
          <h2 class="font-display text-lg mb-3">Test store webhook</h2>
          <form id="commerce-webhook-form" class="grid gap-3 md:grid-cols-2">
            <label class="text-sm">Order name
              <input class="bf-input block mt-1 w-full" id="commerce-name" placeholder="Benchy order"/>
            </label>
            <label class="text-sm">External id
              <input class="bf-input block mt-1 w-full font-mono text-xs" id="commerce-external" placeholder="shop-123"/>
            </label>
            <label class="text-sm">Staged artifact
              <select class="bf-input block mt-1 w-full" id="commerce-artifact-pick">
                <option value="">— optional pick —</option>
              </select>
            </label>
            <label class="text-sm">Artifact id
              <input class="bf-input block mt-1 w-full font-mono text-xs" id="commerce-artifact" placeholder="optional"/>
            </label>
            <label class="text-sm">Spool id
              <input class="bf-input block mt-1 w-full font-mono text-xs" id="commerce-spool" placeholder="optional"/>
            </label>
            <div class="md:col-span-2 flex flex-wrap items-center gap-3">
              <button type="submit" class="bf-btn bf-btn-primary">Queue job</button>
              <p class="text-sm m-0" id="commerce-webhook-msg"></p>
            </div>
          </form>
        </div>
        <div class="bf-panel">
          <h2 class="font-display text-lg mb-3">Mark order status</h2>
          <form id="commerce-status-form" class="flex flex-wrap items-end gap-3">
            <label class="text-sm">Order id
              <input class="bf-input block mt-1" id="commerce-order-id" required placeholder="order id"/>
            </label>
            <label class="text-sm">Status
              <select class="bf-input block mt-1" id="commerce-order-status">
                <option value="packed">packed</option>
                <option value="shipped">shipped</option>
              </select>
            </label>
            <button type="submit" class="bf-btn">Update</button>
            <p class="text-sm m-0" id="commerce-status-msg"></p>
          </form>
          <pre class="bf-pre text-xs mt-3 hidden" id="commerce-result"></pre>
        </div>
        <script>
          const API = {{printf "%q" .APIBase}};
          const commerceStatus = document.getElementById('commerce-status');
          const commerceOut = document.getElementById('commerce-status-out');
          const commerceResult = document.getElementById('commerce-result');
          async function loadCommerceStatus() {
            commerceStatus.textContent = 'loading…';
            commerceStatus.classList.remove('live');
            try {
              const st = await bfJSON(API + '/api/v1/commerce/status');
              commerceOut.textContent = JSON.stringify(st, null, 2);
              commerceStatus.textContent = 'plugin up';
              commerceStatus.classList.add('live');
            } catch (e) {
              commerceStatus.textContent = 'commerce not reachable';
              commerceOut.textContent = String(e.message || e) + '\n\nEnable with: podman compose --profile commerce … and COMMERCE_URL on gateway.';
            }
          }
          async function loadCommerceArtifacts() {
            const sel = document.getElementById('commerce-artifact-pick');
            try {
              const items = await bfJSON(API + '/api/v1/artifacts');
              const opts = ['<option value="">— optional pick —</option>'];
              (items || []).forEach(a => {
                if (!a.id) return;
                const label = (a.filename || a.id) + ' · ' + String(a.id).slice(0, 8);
                opts.push('<option value="' + String(a.id).replace(/"/g, '&quot;') + '">' +
                  String(label).replace(/&/g, '&amp;').replace(/</g, '&lt;') + '</option>');
              });
              sel.innerHTML = opts.join('');
            } catch (_) {}
          }
          document.getElementById('commerce-artifact-pick').onchange = () => {
            const sel = document.getElementById('commerce-artifact-pick');
            if (sel.value) document.getElementById('commerce-artifact').value = sel.value;
          };
          document.getElementById('commerce-refresh').onclick = () => {
            loadCommerceStatus();
            loadCommerceArtifacts();
          };
          document.getElementById('commerce-webhook-form').onsubmit = async (ev) => {
            ev.preventDefault();
            const msg = document.getElementById('commerce-webhook-msg');
            msg.textContent = '';
            commerceResult.classList.add('hidden');
            try {
              const out = await bfJSON(API + '/api/v1/commerce/webhook', {
                method: 'POST', credentials: 'include',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({
                  name: document.getElementById('commerce-name').value.trim() || undefined,
                  external_id: document.getElementById('commerce-external').value.trim() || undefined,
                  artifact_id: document.getElementById('commerce-artifact').value.trim() || undefined,
                  spool_id: document.getElementById('commerce-spool').value.trim() || undefined,
                }),
              });
              msg.textContent = out.queued ? ('Queued job ' + (out.job_id || '')) : (out.reason || 'not queued');
              commerceResult.textContent = JSON.stringify(out, null, 2);
              commerceResult.classList.remove('hidden');
            } catch (e) {
              msg.textContent = String(e.message || e);
            }
          };
          document.getElementById('commerce-status-form').onsubmit = async (ev) => {
            ev.preventDefault();
            const msg = document.getElementById('commerce-status-msg');
            msg.textContent = '';
            commerceResult.classList.add('hidden');
            try {
              const out = await bfJSON(API + '/api/v1/commerce/orders/status', {
                method: 'POST', credentials: 'include',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({
                  order_id: document.getElementById('commerce-order-id').value.trim(),
                  status: document.getElementById('commerce-order-status').value,
                }),
              });
              msg.textContent = 'Updated';
              commerceResult.textContent = JSON.stringify(out, null, 2);
              commerceResult.classList.remove('hidden');
            } catch (e) {
              msg.textContent = String(e.message || e);
            }
          };
          loadCommerceStatus();
          loadCommerceArtifacts();
        </script>
      {{else if eq .Active "analytics"}}
        <p class="bf-lede">Daily utilization, costs, and failure codes from the analytics consumer.</p>
        <div class="flex flex-wrap items-end gap-3 mb-3">
          <label class="text-sm">From
            <input type="date" class="bf-input block mt-1" id="analytics-from"/>
          </label>
          <label class="text-sm">To
            <input type="date" class="bf-input block mt-1" id="analytics-to"/>
          </label>
          <label class="text-sm">Group utilization
            <select class="bf-input block mt-1" id="analytics-group">
              <option value="printer">Printer</option>
              <option value="site">Site</option>
              <option value="day">Day</option>
            </select>
          </label>
          <button type="button" class="bf-btn bf-btn-primary" id="analytics-apply">Apply</button>
          <div id="analytics-status" class="bf-status m-0">loading…</div>
          <div class="flex flex-wrap gap-2 ml-auto">
            <a class="bf-btn" id="analytics-export-json" href="#">Export JSON</a>
            <a class="bf-btn" id="analytics-export-csv" href="#">Export CSV</a>
            <button type="button" class="bf-btn" id="analytics-refresh">Refresh</button>
          </div>
        </div>
        <div class="bf-panel mb-4">
          <h2 class="font-display text-lg mb-3">Event counters</h2>
          <div id="analytics-counters" class="flex flex-wrap gap-2 text-sm">loading…</div>
        </div>
        <div class="bf-panel mb-4">
          <h2 class="font-display text-lg mb-3">Daily utilization</h2>
          <div id="analytics-util-bars" class="space-y-2 mb-4"></div>
          <div class="overflow-x-auto">
            <table class="bf-table" id="analytics-util-table">
              <thead>
                <tr>
                  <th>Day</th>
                  <th>Site</th>
                  <th>Printer</th>
                  <th>Idle</th>
                  <th>Run</th>
                  <th>Fail</th>
                  <th>Util %</th>
                  <th>Samples</th>
                </tr>
              </thead>
              <tbody id="analytics-util-body">
                <tr><td colspan="8" class="bf-empty">loading…</td></tr>
              </tbody>
            </table>
          </div>
        </div>
        <div class="bf-panel mb-4">
          <h2 class="font-display text-lg mb-3">Daily costs</h2>
          <div id="analytics-cost-bars" class="space-y-2 mb-4"></div>
          <div class="overflow-x-auto">
            <table class="bf-table" id="analytics-cost-table">
              <thead>
                <tr>
                  <th>Day</th>
                  <th>Site</th>
                  <th>Material</th>
                  <th>Machine</th>
                  <th>Labor</th>
                  <th>Total</th>
                  <th>Jobs</th>
                </tr>
              </thead>
              <tbody id="analytics-cost-body">
                <tr><td colspan="7" class="bf-empty">loading…</td></tr>
              </tbody>
            </table>
          </div>
        </div>
        <div class="bf-panel">
          <h2 class="font-display text-lg mb-3">Failure codes</h2>
          <div id="analytics-fail-bars" class="space-y-2 mb-4"></div>
          <div class="overflow-x-auto">
            <table class="bf-table" id="analytics-fail-table">
              <thead>
                <tr>
                  <th>Day</th>
                  <th>Site</th>
                  <th>Code</th>
                  <th>Count</th>
                </tr>
              </thead>
              <tbody id="analytics-fail-body">
                <tr><td colspan="4" class="bf-empty">loading…</td></tr>
              </tbody>
            </table>
          </div>
        </div>
        <script>
          const API = {{printf "%q" .APIBase}};
          const statusEl = document.getElementById('analytics-status');
          const countersEl = document.getElementById('analytics-counters');
          const utilBody = document.getElementById('analytics-util-body');
          const utilBars = document.getElementById('analytics-util-bars');
          const costBody = document.getElementById('analytics-cost-body');
          const costBars = document.getElementById('analytics-cost-bars');
          const failBody = document.getElementById('analytics-fail-body');
          const failBars = document.getElementById('analytics-fail-bars');

          function fmtDuration(sec) {
            const s = Number(sec) || 0;
            if (s < 60) return s + 's';
            if (s < 3600) return Math.floor(s / 60) + 'm ' + (s % 60) + 's';
            const h = Math.floor(s / 3600);
            const m = Math.floor((s % 3600) / 60);
            return h + 'h ' + m + 'm';
          }
          function fmtMoney(n) {
            return '$' + (Number(n) || 0).toFixed(2);
          }
          function utilPct(row) {
            const idle = Number(row.idle_s) || 0;
            const run = Number(row.run_s) || 0;
            const fail = Number(row.fail_s) || 0;
            const total = idle + run + fail;
            if (!total) return '—';
            return ((100 * run) / total).toFixed(1) + '%';
          }
          function barRow(label, segments) {
            const parts = segments.map(s =>
              '<span style="width:' + s.pct + '%;background:' + s.color + '" title="' + s.title + '"></span>'
            ).join('');
            return '<div class="flex items-center gap-2 text-xs">' +
              '<span class="font-mono w-24 shrink-0 truncate" title="' + label + '">' + label + '</span>' +
              '<div class="flex-1 h-3 rounded-bf overflow-hidden bg-bf-0 border border-bf-border flex">' + parts + '</div>' +
              '</div>';
          }
          function renderCounters(counters) {
            const keys = Object.keys(counters || {}).sort();
            if (!keys.length) {
              countersEl.innerHTML = '<span class="bf-empty">No counter data yet</span>';
              return;
            }
            countersEl.innerHTML = keys.map(k =>
              '<span class="bf-chip">' + k.replace(/^events\./, '') +
              ' <strong class="font-mono ml-1">' + counters[k] + '</strong></span>'
            ).join('');
          }
          function renderUtilization(series) {
            const rows = (series && series.series) || [];
            if (!rows.length) {
              utilBody.innerHTML = '<tr><td colspan="8" class="bf-empty">No daily rollups yet</td></tr>';
              utilBars.innerHTML = '<p class="bf-empty m-0">No utilization chart yet</p>';
              return;
            }
            utilBody.innerHTML = rows.map(r =>
              '<tr>' +
              '<td class="font-mono text-xs">' + (r.day || '—') + '</td>' +
              '<td>' + (r.site_id || '—') + '</td>' +
              '<td class="font-mono text-xs">' + (r.external_sn || '—') + '</td>' +
              '<td class="font-mono text-xs">' + fmtDuration(r.idle_s) + '</td>' +
              '<td class="font-mono text-xs">' + fmtDuration(r.run_s) + '</td>' +
              '<td class="font-mono text-xs">' + fmtDuration(r.fail_s) + '</td>' +
              '<td class="font-mono text-xs">' + utilPct(r) + '</td>' +
              '<td class="font-mono text-xs">' + (r.samples ?? 0) + '</td>' +
              '</tr>'
            ).join('');
            utilBars.innerHTML = rows.slice(0, 12).map(r => {
              const idle = Number(r.idle_s) || 0;
              const run = Number(r.run_s) || 0;
              const fail = Number(r.fail_s) || 0;
              const total = idle + run + fail || 1;
              const label = (r.day || '') + ' ' + (r.external_sn || r.site_id || '');
              return barRow(label.trim(), [
                { pct: (100 * run) / total, color: 'var(--color-bf-accent, #3dd68c)', title: 'run ' + fmtDuration(run) },
                { pct: (100 * idle) / total, color: 'rgb(100 116 139 / 0.55)', title: 'idle ' + fmtDuration(idle) },
                { pct: (100 * fail) / total, color: 'rgb(239 68 68 / 0.75)', title: 'fail ' + fmtDuration(fail) }
              ]);
            }).join('');
          }
          function renderCosts(series) {
            const rows = (series && series.series) || [];
            if (!rows.length) {
              costBody.innerHTML = '<tr><td colspan="7" class="bf-empty">No cost rollups yet</td></tr>';
              costBars.innerHTML = '<p class="bf-empty m-0">No cost chart yet</p>';
              return;
            }
            const maxTotal = Math.max(...rows.map(r =>
              (Number(r.material_cost)||0) + (Number(r.machine_cost)||0) + (Number(r.labor_cost)||0)
            ), 0.01);
            costBody.innerHTML = rows.map(r => {
              const mat = Number(r.material_cost) || 0;
              const mac = Number(r.machine_cost) || 0;
              const lab = Number(r.labor_cost) || 0;
              return '<tr>' +
                '<td class="font-mono text-xs">' + (r.day || '—') + '</td>' +
                '<td>' + (r.site_id || '—') + '</td>' +
                '<td class="font-mono text-xs">' + fmtMoney(mat) + '</td>' +
                '<td class="font-mono text-xs">' + fmtMoney(mac) + '</td>' +
                '<td class="font-mono text-xs">' + fmtMoney(lab) + '</td>' +
                '<td class="font-mono text-xs">' + fmtMoney(mat + mac + lab) + '</td>' +
                '<td class="font-mono text-xs">' + (r.jobs ?? 0) + '</td>' +
                '</tr>';
            }).join('');
            costBars.innerHTML = rows.slice(0, 12).map(r => {
              const mat = Number(r.material_cost) || 0;
              const mac = Number(r.machine_cost) || 0;
              const lab = Number(r.labor_cost) || 0;
              const total = mat + mac + lab;
              const scale = (100 * total) / maxTotal;
              const inner = total || 1;
              return barRow((r.day || '') + ' ' + (r.site_id || ''), [
                { pct: scale * mat / inner, color: 'var(--color-bf-accent, #3dd68c)', title: 'material ' + fmtMoney(mat) },
                { pct: scale * mac / inner, color: 'rgb(56 189 248 / 0.8)', title: 'machine ' + fmtMoney(mac) },
                { pct: scale * lab / inner, color: 'rgb(250 204 21 / 0.85)', title: 'labor ' + fmtMoney(lab) }
              ]);
            }).join('');
          }
          function renderFailures(series) {
            const rows = (series && series.series) || [];
            if (!rows.length) {
              failBody.innerHTML = '<tr><td colspan="4" class="bf-empty">No failure codes yet</td></tr>';
              failBars.innerHTML = '<p class="bf-empty m-0">No failure chart yet</p>';
              return;
            }
            const maxCount = Math.max(...rows.map(r => Number(r.count) || 0), 1);
            failBody.innerHTML = rows.map(r =>
              '<tr>' +
              '<td class="font-mono text-xs">' + (r.day || '—') + '</td>' +
              '<td>' + (r.site_id || '—') + '</td>' +
              '<td class="font-mono text-xs">' + (r.code || '—') + '</td>' +
              '<td class="font-mono text-xs">' + (r.count ?? 0) + '</td>' +
              '</tr>'
            ).join('');
            failBars.innerHTML = rows.slice(0, 12).map(r => {
              const n = Number(r.count) || 0;
              const label = (r.code || '—') + ' · ' + (r.day || '');
              return barRow(label, [
                { pct: (100 * n) / maxCount, color: 'rgb(239 68 68 / 0.75)', title: n + ' hits' }
              ]);
            }).join('');
          }
          function analyticsQS(extra) {
            const q = new URLSearchParams(extra || {});
            const from = document.getElementById('analytics-from').value;
            const to = document.getElementById('analytics-to').value;
            const gb = document.getElementById('analytics-group').value;
            if (from && to) {
              q.set('from', from);
              q.set('to', to);
            }
            if (gb) q.set('group_by', gb);
            const s = q.toString();
            return s ? ('?' + s) : '';
          }
          function syncExportLinks() {
            document.getElementById('analytics-export-json').href =
              API + '/api/v1/analytics/export' + analyticsQS({format: 'json'});
            document.getElementById('analytics-export-csv').href =
              API + '/api/v1/analytics/export' + analyticsQS({format: 'csv'});
          }
          async function loadAnalytics() {
            statusEl.textContent = 'loading…';
            statusEl.classList.remove('live');
            syncExportLinks();
            const qs = analyticsQS();
            try {
              const [u, c, costs, fails] = await Promise.all([
                bfJSON(API + '/api/v1/analytics/utilization' + qs),
                bfJSON(API + '/api/v1/analytics/counters'),
                bfJSON(API + '/api/v1/analytics/costs' + qs),
                bfJSON(API + '/api/v1/analytics/failures' + qs)
              ]);
              renderCounters(c);
              renderUtilization(u);
              renderCosts(costs);
              renderFailures(fails);
              statusEl.textContent = 'updated';
              statusEl.classList.add('live');
            } catch (e) {
              const msg = String(e.message || e);
              statusEl.textContent = msg;
              statusEl.classList.remove('live');
              countersEl.textContent = msg;
              utilBody.innerHTML = '<tr><td colspan="8" class="bf-empty">' + msg + '</td></tr>';
              costBody.innerHTML = '<tr><td colspan="7" class="bf-empty">' + msg + '</td></tr>';
              failBody.innerHTML = '<tr><td colspan="4" class="bf-empty">' + msg + '</td></tr>';
              utilBars.innerHTML = costBars.innerHTML = failBars.innerHTML = '';
            }
          }
          document.getElementById('analytics-refresh').onclick = loadAnalytics;
          document.getElementById('analytics-apply').onclick = loadAnalytics;
          syncExportLinks();
          loadAnalytics();
        </script>
      {{else}}
        <p class="bf-lede">Local accounts, stack updates, rate cards, power, and auth mode.</p>
        <div class="bf-panel mb-4">
          <h2 class="font-display text-lg mb-2">System updates</h2>
          <p class="text-sm text-bf-muted mb-3">Keep BambuFarmer in step with Farm Manager: after upgrading FM, update this stack. One command on the host, or request from here while a host watcher is running.</p>
          <pre class="bf-pre text-xs mb-3" id="sys-out">loading…</pre>
          <div class="flex flex-wrap gap-2 items-center">
            <button type="button" class="bf-btn bf-btn-primary" id="sys-update-btn">Request update</button>
            <code class="text-xs text-bf-muted">./scripts/bf update</code>
            <span class="text-xs text-bf-muted">or</span>
            <code class="text-xs text-bf-muted">./scripts/bf update --watch</code>
          </div>
          <p class="text-sm mt-2" id="sys-msg"></p>
        </div>
        <div class="bf-panel mb-4">
          <h2 class="font-display text-lg mb-2">Power</h2>
          <p class="text-sm text-bf-muted mb-3">Smart-plug webhook via <code>POWER_URL</code>. Empty URL accepts the call and does not switch.</p>
          <div id="power-status" class="bf-status m-0 mb-3">loading…</div>
          <div class="flex flex-wrap gap-2 items-center">
            <button type="button" class="bf-btn bf-btn-primary" id="power-on" data-power="on">On</button>
            <button type="button" class="bf-btn" id="power-off" data-power="off">Off</button>
            <p class="text-sm m-0" id="power-msg"></p>
          </div>
        </div>
        <div class="bf-panel mb-4">
          <h2 class="font-display text-lg mb-2">Notify</h2>
          <p class="text-sm text-bf-muted mb-3">Job-finished / alert webhook via <code>NOTIFY_URL</code>. Empty URL accepts the test and does not POST.</p>
          <div id="notify-status" class="bf-status m-0 mb-3">loading…</div>
          <div class="flex flex-wrap gap-2 items-center">
            <button type="button" class="bf-btn bf-btn-primary" id="notify-test">Send test</button>
            <p class="text-sm m-0" id="notify-msg"></p>
          </div>
        </div>
        <div class="bf-panel mb-4">
          <h2 class="font-display text-lg mb-2">Site</h2>
          <p class="text-sm text-bf-muted mb-3">Display name and timezone for the farm site (single-site MVP).</p>
          <form id="site-form" class="grid gap-x-4 gap-y-1 md:grid-cols-2 max-w-xl">
            <input type="hidden" id="site-id"/>
            <div>
              <label class="bf-field" for="site-name">Name</label>
              <input class="bf-input" id="site-name" required/>
            </div>
            <div>
              <label class="bf-field" for="site-tz">Timezone</label>
              <input class="bf-input" id="site-tz" placeholder="UTC" required/>
            </div>
            <div class="md:col-span-2 flex flex-wrap gap-2 items-center mt-2">
              <button type="submit" class="bf-btn bf-btn-primary">Save site</button>
              <p class="text-sm self-center" id="site-msg"></p>
            </div>
          </form>
        </div>
        <div class="bf-panel mb-4">
          <h2 class="font-display text-lg mb-2">Stack &amp; plugins</h2>
          <p class="text-sm text-bf-muted mb-3">Core version plus optional compose profiles (slicer, commerce) and integration URLs.</p>
          <div id="stack-status" class="flex flex-wrap gap-2 text-sm">loading…</div>
          <button type="button" class="bf-btn mt-3" id="stack-refresh">Refresh</button>
        </div>
        <div class="bf-panel mb-4">
          <h2 class="font-display text-lg mb-2">Farm tags</h2>
          <p class="text-sm text-bf-muted mb-3">Circuit/group labels from edge (<code>GET /api/v1/edge/tags</code>). Empty when the plane has no tag source.</p>
          <div id="edge-tags" class="bf-empty">loading…</div>
          <button type="button" class="bf-btn mt-3" id="edge-tags-refresh">Refresh tags</button>
        </div>
        <div class="grid gap-4 md:grid-cols-2">
          <div class="bf-panel">
            <h2 class="font-display text-lg mb-3">Users</h2>
            <div id="users-list" class="bf-empty text-left">loading…</div>
            <form id="user-form" class="mt-4 space-y-2 border-t border-bf-border pt-4">
              <label class="bf-field" for="nu">Username</label>
              <input class="bf-input" id="nu" required/>
              <label class="bf-field" for="np">Password</label>
              <input class="bf-input" id="np" type="password" minlength="8" required/>
              <label class="bf-field" for="nr">Role</label>
              <select class="bf-input" id="nr">
                <option value="viewer">viewer</option>
                <option value="operator">operator</option>
                <option value="admin">admin</option>
              </select>
              <button type="submit" class="bf-btn bf-btn-primary mt-2">Create user</button>
              <p class="text-bf-danger text-sm" id="user-err"></p>
            </form>
          </div>
          <div class="bf-panel space-y-4">
            <div>
              <h2 class="font-display text-lg mb-3">Change my password</h2>
              <form id="pw-form" class="space-y-2">
                <label class="bf-field" for="pw-cur">Current password</label>
                <input class="bf-input" id="pw-cur" type="password" autocomplete="current-password" required/>
                <label class="bf-field" for="pw-new">New password</label>
                <input class="bf-input" id="pw-new" type="password" autocomplete="new-password" minlength="8" required/>
                <label class="bf-field" for="pw-new2">Confirm new password</label>
                <input class="bf-input" id="pw-new2" type="password" autocomplete="new-password" minlength="8" required/>
                <button type="submit" class="bf-btn bf-btn-primary mt-2">Update password</button>
                <p class="text-sm mt-2" id="pw-msg"></p>
              </form>
            </div>
            <div>
              <h2 class="font-display text-lg mb-3">Rate card</h2>
              <p class="text-sm text-bf-muted mb-3">Site costing defaults for job estimates and finish actuals.</p>
              <p class="text-xs text-bf-muted mb-3" id="admin-me">Session…</p>
              <form id="rates-form" class="space-y-2">
                <label class="bf-field" for="rate-currency">Currency</label>
                <input class="bf-input" id="rate-currency" name="currency" placeholder="USD" maxlength="8"/>
                <label class="bf-field" for="rate-machine">Machine $/hr</label>
                <input class="bf-input" id="rate-machine" name="machine_rate_per_hour" type="number" min="0" step="0.01"/>
                <label class="bf-field" for="rate-filament">Filament $/kg</label>
                <input class="bf-input" id="rate-filament" name="filament_price_per_kg" type="number" min="0" step="0.01"/>
                <label class="bf-field" for="rate-kghr">Filament kg/hr</label>
                <input class="bf-input" id="rate-kghr" name="filament_kg_per_hour" type="number" min="0" step="0.001"/>
                <label class="bf-field" for="rate-labor">Labor flat</label>
                <input class="bf-input" id="rate-labor" name="labor_flat" type="number" min="0" step="0.01"/>
                <label class="bf-field" for="rate-fail">Failure surcharge %</label>
                <input class="bf-input" id="rate-fail" name="failure_surcharge_pct" type="number" min="0" step="0.1"/>
                <button type="submit" class="bf-btn bf-btn-primary mt-2">Save rates</button>
                <p class="text-sm mt-2" id="rates-msg"></p>
              </form>
            </div>
          </div>
        </div>
        <script>
          const API = {{printf "%q" .APIBase}};
          const opts = { credentials: 'include' };
          async function patchUser(id, body) {
            await bfJSON(API + '/api/v1/admin/users/' + encodeURIComponent(id), {
              method: 'PATCH', credentials: 'include',
              headers: {'Content-Type': 'application/json'},
              body: JSON.stringify(body),
            });
            await loadUsers();
          }
          async function loadUsers() {
            const users = await bfJSON(API + '/api/v1/admin/users', opts);
            const el = document.getElementById('users-list');
            if (!users || !users.length) { el.innerHTML = '<p class="bf-empty">No users</p>'; return; }
            el.innerHTML = '<ul class="space-y-2">' + users.map(u => {
              const role = (u.roles && u.roles[0]) || 'operator';
              return '<li class="bf-card">' +
              '<div class="bf-card-name">' + u.username + (u.active ? '' : ' (disabled)') + '</div>' +
              '<div class="bf-card-meta">' + (u.roles||[]).join(', ') + '</div>' +
              '<div class="flex flex-wrap gap-1.5 mt-2 items-center">' +
              '<select class="bf-input !w-auto text-xs" data-role="' + u.id + '">' +
              '<option value="operator"' + (role === 'operator' ? ' selected' : '') + '>operator</option>' +
              '<option value="admin"' + (role === 'admin' ? ' selected' : '') + '>admin</option>' +
              '</select>' +
              '<button type="button" class="bf-btn" data-save-role="' + u.id + '">Save role</button>' +
              '<button type="button" class="bf-btn" data-active="' + u.id + '" data-next="' + (u.active ? '0' : '1') + '">' +
                (u.active ? 'Disable' : 'Enable') + '</button>' +
              '<input class="bf-input !w-auto min-w-[8rem] text-xs" type="password" placeholder="new password" data-reset-pw="' + u.id + '" minlength="8"/>' +
              '<button type="button" class="bf-btn" data-reset="' + u.id + '">Reset password</button>' +
              '</div></li>';
            }).join('') + '</ul>';
            el.querySelectorAll('button[data-save-role]').forEach(btn => {
              btn.onclick = async () => {
                const id = btn.dataset.saveRole;
                const sel = el.querySelector('select[data-role="' + id + '"]');
                try {
                  await patchUser(id, {roles: [sel.value]});
                } catch (e) { alert(String(e.message || e)); }
              };
            });
            el.querySelectorAll('button[data-active]').forEach(btn => {
              btn.onclick = async () => {
                try {
                  await patchUser(btn.dataset.active, {active: btn.dataset.next === '1'});
                } catch (e) { alert(String(e.message || e)); }
              };
            });
            el.querySelectorAll('button[data-reset]').forEach(btn => {
              btn.onclick = async () => {
                const id = btn.dataset.reset;
                const input = el.querySelector('input[data-reset-pw="' + id + '"]');
                const pw = (input && input.value) || '';
                if (pw.length < 8) { alert('New password must be at least 8 characters'); return; }
                try {
                  await bfJSON(API + '/api/v1/admin/users/' + id + '/password', {
                    method: 'POST', credentials: 'include',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({ new_password: pw }),
                  });
                  if (input) input.value = '';
                  btn.textContent = 'Reset ✓';
                  setTimeout(() => { btn.textContent = 'Reset password'; }, 1500);
                } catch (e) { alert(String(e.message || e)); }
              };
            });
          }
          document.getElementById('user-form').onsubmit = async (ev) => {
            ev.preventDefault();
            const err = document.getElementById('user-err');
            err.textContent = '';
            try {
              await bfJSON(API + '/api/v1/admin/users', {
                method: 'POST', credentials: 'include',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({
                  username: document.getElementById('nu').value.trim(),
                  password: document.getElementById('np').value,
                  roles: [document.getElementById('nr').value],
                }),
              });
              document.getElementById('user-form').reset();
              await loadUsers();
            } catch (e) { err.textContent = String(e.message || e); }
          };
          document.getElementById('pw-form').onsubmit = async (ev) => {
            ev.preventDefault();
            const msg = document.getElementById('pw-msg');
            msg.className = 'text-sm mt-2';
            msg.textContent = '';
            const a = document.getElementById('pw-new').value;
            const b = document.getElementById('pw-new2').value;
            if (a !== b) { msg.className = 'text-bf-danger text-sm mt-2'; msg.textContent = 'New passwords do not match'; return; }
            try {
              await bfJSON(API + '/api/v1/auth/password', {
                method: 'POST', credentials: 'include',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({
                  current_password: document.getElementById('pw-cur').value,
                  new_password: a,
                }),
              });
              msg.className = 'text-bf-accent text-sm mt-2';
              msg.textContent = 'Password updated';
              document.getElementById('pw-form').reset();
            } catch (e) {
              msg.className = 'text-bf-danger text-sm mt-2';
              msg.textContent = String(e.message || e);
            }
          };
          async function loadSystem() {
            const sys = await bfJSON(API + '/api/v1/admin/system', opts);
            document.getElementById('sys-out').textContent = JSON.stringify(sys, null, 2);
            return sys;
          }
          async function loadPower() {
            const el = document.getElementById('power-status');
            try {
              const st = await bfJSON(API + '/api/v1/power', opts);
              el.textContent = st.configured ? 'POWER_URL configured' : 'POWER_URL unset (dry-run)';
              el.classList.toggle('live', !!st.configured);
            } catch (e) {
              el.textContent = String(e.message || e);
              el.classList.remove('live');
            }
          }
          async function loadNotify() {
            const el = document.getElementById('notify-status');
            try {
              const st = await bfJSON(API + '/api/v1/notify', opts);
              el.textContent = st.configured ? 'NOTIFY_URL configured' : 'NOTIFY_URL unset (dry-run)';
              el.classList.toggle('live', !!st.configured);
            } catch (e) {
              el.textContent = String(e.message || e);
              el.classList.remove('live');
            }
          }
          function stackChip(label, ok, detail) {
            const cls = ok ? 'bf-chip' : 'bf-chip opacity-70';
            return '<span class="' + cls + '" title="' + String(detail || '').replace(/"/g, '&quot;') + '">' +
              label + (ok ? ' · on' : ' · off') + '</span>';
          }
          async function probe(path) {
            try {
              return {ok: true, data: await bfJSON(API + path, opts)};
            } catch (e) {
              return {ok: false, err: String(e.message || e)};
            }
          }
          async function loadStack() {
            const el = document.getElementById('stack-status');
            el.textContent = 'loading…';
            const [ver, sites, integ, power, notify, edge, slice, commerce] = await Promise.all([
              probe('/api/v1/version'),
              probe('/api/v1/sites'),
              probe('/api/v1/integrations/status'),
              probe('/api/v1/power'),
              probe('/api/v1/notify'),
              probe('/api/v1/edge/status'),
              probe('/api/v1/slice/status'),
              probe('/api/v1/commerce/status'),
            ]);
            const chips = [];
            if (ver.ok) {
              const v = ver.data || {};
              chips.push(stackChip('core ' + (v.version || v.Version || 'dev'), true,
                [v.commit || v.Commit, v.built_at || v.BuiltAt].filter(Boolean).join(' · ')));
            } else {
              chips.push(stackChip('core', false, ver.err));
            }
            if (sites.ok) {
              const list = Array.isArray(sites.data) ? sites.data : [];
              const label = list.length === 1
                ? ('site ' + (list[0].name || list[0].id))
                : ('sites ×' + list.length);
              chips.push(stackChip(label, list.length > 0, list.map(s => s.id).join(', ')));
            } else {
              chips.push(stackChip('sites', false, sites.err));
            }
            if (integ.ok) {
              const s = integ.data || {};
              chips.push(stackChip('spoolman', !!(s.spoolman && s.spoolman.configured), JSON.stringify(s.spoolman || {})));
              chips.push(stackChip('manyfold', !!(s.manyfold && s.manyfold.configured), JSON.stringify(s.manyfold || {})));
            } else {
              chips.push(stackChip('integrations', false, integ.err));
            }
            chips.push(stackChip('power', !!(power.ok && power.data && power.data.configured), power.ok ? '' : power.err));
            chips.push(stackChip('notify', !!(notify.ok && notify.data && notify.data.configured), notify.ok ? '' : notify.err));
            if (edge.ok) {
              const e = edge.data || {};
              chips.push(stackChip('edge ' + (e.transport || e.farm_mode || ''), !!e.connected,
                [e.note, e.last_error, 'devices=' + (e.device_count ?? '?')].filter(Boolean).join(' · ')));
            } else {
              chips.push(stackChip('edge', false, edge.err));
            }
            chips.push(stackChip('slicer', !!(slice.ok && slice.data), slice.ok
              ? ('orca=' + !!(slice.data && slice.data.orca_configured))
              : (slice.err || 'profile off')));
            chips.push(stackChip('commerce', !!(commerce.ok && commerce.data), commerce.ok
              ? String((commerce.data && commerce.data.status) || 'up')
              : (commerce.err || 'profile off')));
            el.innerHTML = chips.join('');
          }
          async function loadSite() {
            const msg = document.getElementById('site-msg');
            msg.textContent = '';
            try {
              const list = await bfJSON(API + '/api/v1/sites', opts);
              const site = Array.isArray(list) && list.length ? list[0] : null;
              if (!site) {
                msg.className = 'text-bf-muted text-sm self-center';
                msg.textContent = 'No site row yet';
                return;
              }
              document.getElementById('site-id').value = site.id || '';
              document.getElementById('site-name').value = site.name || '';
              document.getElementById('site-tz').value = site.timezone || 'UTC';
            } catch (e) {
              msg.className = 'text-bf-danger text-sm self-center';
              msg.textContent = String(e.message || e);
            }
          }
          document.getElementById('site-form').onsubmit = async (ev) => {
            ev.preventDefault();
            const msg = document.getElementById('site-msg');
            msg.className = 'text-sm self-center';
            msg.textContent = '';
            const id = document.getElementById('site-id').value.trim();
            if (!id) {
              msg.className = 'text-bf-danger text-sm self-center';
              msg.textContent = 'No site id';
              return;
            }
            try {
              await bfJSON(API + '/api/v1/sites/' + encodeURIComponent(id), {
                method: 'PATCH', credentials: 'include',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({
                  name: document.getElementById('site-name').value.trim(),
                  timezone: document.getElementById('site-tz').value.trim() || 'UTC',
                }),
              });
              msg.className = 'text-bf-accent text-sm self-center';
              msg.textContent = 'Saved';
              await loadStack();
            } catch (e) {
              msg.className = 'text-bf-danger text-sm self-center';
              msg.textContent = String(e.message || e);
            }
          };
          async function loadEdgeTags() {
            const el = document.getElementById('edge-tags');
            el.className = 'bf-empty';
            el.textContent = 'loading…';
            try {
              const data = await bfJSON(API + '/api/v1/edge/tags', opts);
              const tags = (data && data.tags) || [];
              if (!tags.length) {
                el.textContent = 'No tags';
                return;
              }
              el.className = 'overflow-x-auto';
              el.innerHTML = '<table class="bf-table"><thead><tr><th>Name</th><th>ID</th><th>Devices</th></tr></thead><tbody>' +
                tags.map(t => {
                  const members = (t.devices && t.devices.length) ? t.devices : (t.dev_ids || []);
                  return '<tr><td>' + String(t.name || '—') + '</td>' +
                    '<td class="font-mono text-xs">' + String(t.id || '—') + '</td>' +
                    '<td class="font-mono text-xs">' + (members.length ? members.join(', ') : '—') + '</td></tr>';
                }).join('') + '</tbody></table>';
            } catch (e) {
              el.className = 'text-bf-danger text-sm';
              el.textContent = String(e.message || e);
            }
          }
          document.getElementById('stack-refresh').onclick = loadStack;
          document.getElementById('edge-tags-refresh').onclick = loadEdgeTags;
          loadStack();
          loadSite();
          loadEdgeTags();
          document.querySelectorAll('[data-power]').forEach(btn => {
            btn.onclick = async () => {
              const msg = document.getElementById('power-msg');
              msg.textContent = '';
              try {
                const out = await bfJSON(API + '/api/v1/power', {
                  method: 'POST', credentials: 'include',
                  headers: {'Content-Type': 'application/json'},
                  body: JSON.stringify({action: btn.dataset.power}),
                });
                msg.textContent = out.switched
                  ? ('Switched ' + out.action)
                  : (out.reason || 'not switched');
              } catch (e) {
                msg.textContent = String(e.message || e);
              }
            };
          });
          loadPower();
          document.getElementById('notify-test').onclick = async () => {
            const msg = document.getElementById('notify-msg');
            msg.textContent = '';
            try {
              const out = await bfJSON(API + '/api/v1/notify/test', {
                method: 'POST', credentials: 'include',
              });
              msg.textContent = out.sent
                ? ('Sent test · ' + (out.job_id || 'notify-test'))
                : (out.reason || 'not sent');
            } catch (e) {
              msg.textContent = String(e.message || e);
            }
          };
          loadNotify();
          document.getElementById('sys-update-btn').onclick = async () => {
            const msg = document.getElementById('sys-msg');
            msg.className = 'text-sm mt-2';
            msg.textContent = '';
            try {
              const res = await bfJSON(API + '/api/v1/admin/system/update', {
                method: 'POST', credentials: 'include',
              });
              msg.className = 'text-bf-accent text-sm mt-2';
              msg.textContent = res.hint || 'Update requested — host watcher will apply it.';
              await loadSystem();
            } catch (e) {
              msg.className = 'text-bf-danger text-sm mt-2';
              msg.textContent = String(e.message || e);
            }
          };
          function fillRates(rc) {
            if (!rc || rc.error) return;
            document.getElementById('rate-currency').value = rc.currency || 'USD';
            document.getElementById('rate-machine').value = rc.machine_rate_per_hour ?? 0;
            document.getElementById('rate-filament').value = rc.filament_price_per_kg ?? 0;
            document.getElementById('rate-kghr').value = rc.filament_kg_per_hour ?? 0;
            document.getElementById('rate-labor').value = rc.labor_flat ?? 0;
            document.getElementById('rate-fail').value = rc.failure_surcharge_pct ?? 0;
          }
          document.getElementById('rates-form').onsubmit = async (ev) => {
            ev.preventDefault();
            const msg = document.getElementById('rates-msg');
            msg.className = 'text-sm mt-2';
            msg.textContent = '';
            const body = {
              currency: document.getElementById('rate-currency').value.trim() || 'USD',
              machine_rate_per_hour: Number(document.getElementById('rate-machine').value) || 0,
              filament_price_per_kg: Number(document.getElementById('rate-filament').value) || 0,
              filament_kg_per_hour: Number(document.getElementById('rate-kghr').value) || 0,
              labor_flat: Number(document.getElementById('rate-labor').value) || 0,
              failure_surcharge_pct: Number(document.getElementById('rate-fail').value) || 0,
            };
            try {
              const out = await bfJSON(API + '/api/v1/admin/rate-cards', {
                method: 'PUT', credentials: 'include',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify(body),
              });
              fillRates(out);
              msg.className = 'text-bf-accent text-sm mt-2';
              msg.textContent = 'Rates saved';
            } catch (e) {
              msg.className = 'text-bf-danger text-sm mt-2';
              msg.textContent = String(e.message || e);
            }
          };
          Promise.all([
            bfJSON(API + '/api/v1/auth/me', opts),
            bfJSON(API + '/api/v1/admin/rate-cards', opts),
            loadUsers().catch(e => { document.getElementById('users-list').textContent = String(e); }),
            loadSystem().catch(e => { document.getElementById('sys-out').textContent = String(e); })
          ]).then(([me, rates]) => {
            const meEl = document.getElementById('admin-me');
            meEl.textContent = me && me.username
              ? ('Signed in as ' + me.username + (me.roles && me.roles.length ? ' · ' + me.roles.join(', ') : ''))
              : 'Session unknown';
            fillRates(rates);
          }).catch(e => {
            document.getElementById('rates-msg').className = 'text-bf-danger text-sm mt-2';
            document.getElementById('rates-msg').textContent = String(e.message || e);
          });
        </script>
      {{end}}
    </main>
  </div>
</body>
</html>
`

const loginHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1"/>
  <title>Login · BambuFarmer</title>
  <link rel="preconnect" href="https://fonts.googleapis.com"/>
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin/>
  <link href="https://fonts.googleapis.com/css2?family=IBM+Plex+Sans:wght@400;500;600;700&family=Syne:wght@600;700&display=swap" rel="stylesheet"/>
  <link rel="stylesheet" href="/static/css/app.css"/>
</head>
<body class="bf-wizard">
  <div class="bf-wizard-shell">
    <div class="bf-brand text-[1.35rem]">BambuFarmer<span>Sign in</span></div>
    <h1 class="bf-title mt-5">Login</h1>
    <p class="bf-lede">LAN mode local account. Forgot password? Ask an admin to reset it on the Admin page (no email reset on LAN).</p>
    <form class="bf-panel" id="login-form">
      <label class="bf-field" for="username">Username</label>
      <input class="bf-input" id="username" name="username" autocomplete="username" required/>
      <label class="bf-field" for="password">Password</label>
      <input class="bf-input" id="password" name="password" type="password" autocomplete="current-password" required/>
      <div class="bf-actions">
        <button type="submit" class="bf-btn bf-btn-primary">Sign in</button>
      </div>
      <p class="text-bf-danger text-sm mt-3" id="err"></p>
    </form>
  </div>
  <script>
    const API = {{printf "%q" .APIBase}};
    document.getElementById('login-form').onsubmit = async (ev) => {
      ev.preventDefault();
      const err = document.getElementById('err');
      err.textContent = '';
      try {
        const r = await fetch(API + '/api/v1/auth/login', {
          method: 'POST',
          credentials: 'include',
          headers: {'Content-Type': 'application/json'},
          body: JSON.stringify({
            username: document.getElementById('username').value.trim(),
            password: document.getElementById('password').value,
          }),
        });
        const text = await r.text();
        let data = null;
        try { data = text ? JSON.parse(text) : null; } catch (_) {}
        if (!r.ok) throw new Error((data && data.error) || text || r.statusText);
        location.href = '/';
      } catch (e) {
        err.textContent = String(e.message || e);
      }
    };
  </script>
</body>
</html>
`
