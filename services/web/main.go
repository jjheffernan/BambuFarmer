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
	mux.HandleFunc("GET /presets", func(w http.ResponseWriter, _ *http.Request) {
		render(w, "Presets", "presets")
	})
	mux.HandleFunc("GET /analytics", func(w http.ResponseWriter, _ *http.Request) {
		render(w, "Analytics", "analytics")
	})
	mux.HandleFunc("GET /admin", func(w http.ResponseWriter, _ *http.Request) {
		render(w, "Admin", "admin")
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
</head>
<body>
  <div class="bf-shell">
    <aside class="bf-aside">
      <div class="bf-brand">BambuFarmer<span>LAN farm ops</span></div>
      <nav class="bf-nav">
        <a href="/" {{if eq .Active "fleet"}}class="active"{{end}}>Fleet</a>
        <a href="/jobs" {{if eq .Active "jobs"}}class="active"{{end}}>Jobs</a>
        <a href="/presets" {{if eq .Active "presets"}}class="active"{{end}}>Presets</a>
        <a href="/analytics" {{if eq .Active "analytics"}}class="active"{{end}}>Analytics</a>
        <a href="/admin" {{if eq .Active "admin"}}class="active"{{end}}>Admin</a>
      </nav>
      <p class="mt-auto text-xs text-bf-muted hidden md:block">Gateway API<br/><code class="text-[0.7rem] break-all">{{.APIBase}}</code></p>
    </aside>
    <main class="bf-main">
      <h1 class="bf-title">{{.Title}}</h1>
      {{if eq .Active "fleet"}}
        <p class="bf-lede">Live fleet via gateway SSE — no browser polling of Farm Manager.</p>
        <div id="sse-status" class="bf-status">connecting…</div>
        <div class="flex flex-wrap gap-2 mb-4" id="bucket-chips">
          <span class="bf-chip on" data-bucket="all">All</span>
          <span class="bf-chip" data-bucket="ready_to_start">Ready to Start</span>
          <span class="bf-chip" data-bucket="attention_needed">Attention Needed</span>
          <span class="bf-chip" data-bucket="completed_task">Completed</span>
          <span class="bf-chip" data-bucket="not_in_task">Not in Task</span>
          <span class="bf-chip" data-bucket="printing">Printing</span>
        </div>
        <div id="printer-grid" class="grid gap-3 grid-cols-[repeat(auto-fill,minmax(220px,1fr))]"></div>
        <script>
          const API = {{printf "%q" .APIBase}};
          const printers = new Map();
          let filter = 'all';
          const statusEl = document.getElementById('sse-status');
          const grid = document.getElementById('printer-grid');

          document.querySelectorAll('#bucket-chips .bf-chip').forEach(chip => {
            chip.addEventListener('click', () => {
              document.querySelectorAll('#bucket-chips .bf-chip').forEach(c => c.classList.remove('on'));
              chip.classList.add('on');
              filter = chip.dataset.bucket;
              render();
            });
          });

          function upsert(p) {
            if (!p || !p.external_sn) return;
            printers.set(p.external_sn, p);
            render();
          }

          function render() {
            const items = [...printers.values()].filter(p => filter === 'all' || p.bucket === filter);
            if (!items.length) {
              grid.innerHTML = '<div class="bf-panel bf-empty">Waiting for edge deltas (fixture or Farm Manager)…</div>';
              return;
            }
            grid.innerHTML = items.map(p => {
              const sn = p.external_sn;
              return '<article class="bf-card" id="printer-' + sn + '">' +
                '<div class="bf-card-name">' + (p.name || sn) + '</div>' +
                '<div class="bf-card-meta">' + (p.model || '') + ' · <span class="state-' + (p.gcode_state||'') + '">' + (p.gcode_state||'?') + '</span>' +
                (p.mc_percent ? ' · ' + p.mc_percent + '%' : '') + '</div>' +
                '<div class="flex flex-wrap gap-1.5 mt-3">' +
                '<button type="button" class="bf-btn" data-sn="' + sn + '" data-opt="pause">Pause</button>' +
                '<button type="button" class="bf-btn" data-sn="' + sn + '" data-opt="resume">Resume</button>' +
                '<button type="button" class="bf-btn" data-sn="' + sn + '" data-opt="stop">Stop</button>' +
                '<button type="button" class="bf-btn" data-sn="' + sn + '" data-opt="bed_clean">Bed clean</button>' +
                '</div></article>';
            }).join('');
            grid.querySelectorAll('button.bf-btn').forEach(btn => {
              btn.addEventListener('click', async () => {
                await fetch(API + '/api/v1/edge/control/' + btn.dataset.sn, {
                  method: 'POST',
                  headers: {'Content-Type': 'application/json'},
                  body: JSON.stringify({opt: btn.dataset.opt})
                });
              });
            });
          }

          async function bootstrap() {
            try {
              const r = await fetch(API + '/api/v1/printers');
              if (r.ok) {
                const list = await r.json();
                (list || []).forEach(upsert);
              }
              const e = await fetch(API + '/api/v1/edge/status');
              if (e.ok) {
                const st = await e.json();
                (st.devices || []).forEach(upsert);
              }
            } catch (_) {}
          }

          function connectSSE() {
            const es = new EventSource(API + '/api/v1/events/stream');
            es.addEventListener('ready', () => {
              statusEl.textContent = 'SSE live';
              statusEl.classList.add('live');
            });
            ['printer.state_changed','printer.online','printer.offline'].forEach(ev => {
              es.addEventListener(ev, (msg) => {
                try {
                  const env = JSON.parse(msg.data);
                  upsert(env.payload || {});
                } catch (_) {}
              });
            });
            es.onerror = () => {
              statusEl.textContent = 'SSE reconnecting…';
              statusEl.classList.remove('live');
            };
          }

          bootstrap();
          connectSSE();
        </script>
      {{else if eq .Active "jobs"}}
        <p class="bf-lede">Queue owned by core. Create jobs via API, then watch state here.</p>
        <div class="bf-panel overflow-x-auto">
          <table class="bf-table" id="jobs-table">
            <thead><tr><th>Name</th><th>State</th><th>Printer</th><th>Est.</th><th>Cost</th></tr></thead>
            <tbody><tr><td colspan="5" class="bf-empty">loading…</td></tr></tbody>
          </table>
        </div>
        <script>
          const API = {{printf "%q" .APIBase}};
          fetch(API + '/api/v1/jobs').then(r => r.json()).then(jobs => {
            const tb = document.querySelector('#jobs-table tbody');
            if (!jobs || !jobs.length) {
              tb.innerHTML = '<tr><td colspan="5" class="bf-empty">No jobs yet — POST /api/v1/jobs to enqueue.</td></tr>';
              return;
            }
            tb.innerHTML = jobs.map(j => '<tr>' +
              '<td>' + (j.name || j.id) + '</td>' +
              '<td>' + (j.state || '') + '</td>' +
              '<td class="text-bf-muted">' + (j.external_sn || '—') + '</td>' +
              '<td class="text-bf-muted">' + (j.est_seconds ? Math.round(j.est_seconds/60) + 'm' : '—') +
              (j.est_grams ? ' / ' + j.est_grams + 'g' : '') + '</td>' +
              '<td class="text-bf-muted">' + (j.cost_estimate ? j.cost_estimate.toFixed(2) + ' ' + (j.currency||'') : '—') + '</td>' +
              '</tr>').join('');
          }).catch(e => {
            document.querySelector('#jobs-table tbody').innerHTML =
              '<tr><td colspan="5" class="text-bf-danger">' + e + '</td></tr>';
          });
        </script>
      {{else if eq .Active "presets"}}
        <p class="bf-lede">Versioned filament and print presets stored in Postgres.</p>
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
        <script>
          const API = {{printf "%q" .APIBase}};
          function list(el, items, map) {
            if (!items || !items.length) { el.innerHTML = '<p class="bf-empty">None yet</p>'; return; }
            el.innerHTML = '<ul class="space-y-2">' + items.map(map).join('') + '</ul>';
          }
          Promise.all([
            fetch(API + '/api/v1/presets/filament').then(r => r.json()),
            fetch(API + '/api/v1/presets/print').then(r => r.json())
          ]).then(([f, p]) => {
            list(document.getElementById('filament-list'), f, x =>
              '<li class="bf-card"><div class="bf-card-name">' + x.name + '</div>' +
              '<div class="bf-card-meta">' + (x.material||'') + ' · ' + (x.vendor||'') +
              (x.price_per_kg ? ' · $' + x.price_per_kg + '/kg' : '') + '</div></li>');
            list(document.getElementById('print-list'), p, x =>
              '<li class="bf-card"><div class="bf-card-name">' + x.name + '</div>' +
              '<div class="bf-card-meta">' + (x.layer_height_mm||'') + 'mm · ø' + (x.nozzle_diameter_mm||'') + '</div></li>');
          }).catch(e => {
            document.getElementById('filament-list').textContent = String(e);
          });
        </script>
      {{else if eq .Active "analytics"}}
        <p class="bf-lede">Utilization and counters from the analytics consumer.</p>
        <div class="bf-panel">
          <pre class="bf-pre" id="analytics-out">loading…</pre>
        </div>
        <script>
          const API = {{printf "%q" .APIBase}};
          Promise.all([
            fetch(API + '/api/v1/analytics/utilization').then(r => r.json()),
            fetch(API + '/api/v1/analytics/counters').then(r => r.json())
          ]).then(([u, c]) => {
            document.getElementById('analytics-out').textContent = JSON.stringify({utilization: u, counters: c}, null, 2);
          }).catch(e => { document.getElementById('analytics-out').textContent = String(e); });
        </script>
      {{else}}
        <p class="bf-lede">Site identity, rate cards, and auth mode.</p>
        <div class="bf-panel">
          <p class="text-sm text-bf-muted mb-4">Auth modes: see <code>docs/DEPLOYMENT_MODES.md</code> (LAN vs OIDC).</p>
          <pre class="bf-pre" id="admin-out">loading…</pre>
        </div>
        <script>
          const API = {{printf "%q" .APIBase}};
          Promise.all([
            fetch(API + '/api/v1/auth/me').then(r => r.json()),
            fetch(API + '/api/v1/admin/rate-cards').then(r => r.json())
          ]).then(([me, rates]) => {
            document.getElementById('admin-out').textContent = JSON.stringify({me, rates}, null, 2);
          }).catch(e => { document.getElementById('admin-out').textContent = String(e); });
        </script>
      {{end}}
    </main>
  </div>
</body>
</html>
`
