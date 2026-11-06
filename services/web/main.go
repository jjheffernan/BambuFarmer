package main

import (
	"html/template"
	"log/slog"
	"net/http"
	"os"

	"github.com/jjheffernan/BambuFarmer/packages/httpserver"
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
	registerSetup(mux, apiBase)

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
  <script src="https://unpkg.com/htmx.org@2.0.4"></script>
  <style>
    :root {
      --bf-bg-0: #0e1114;
      --bf-bg-1: #161b20;
      --bf-bg-2: #1e252c;
      --bf-text: #e8eef2;
      --bf-muted: #8b9aab;
      --bf-accent: #00ae42;
      --bf-warn: #f0a202;
      --bf-danger: #e23d28;
      --bf-info: #3aa0ff;
      --bf-border: #2a343d;
      --bf-radius: 6px;
      --bf-font-ui: "IBM Plex Sans", "Segoe UI", sans-serif;
    }
    * { box-sizing: border-box; }
    body {
      margin: 0; min-height: 100vh;
      font-family: var(--bf-font-ui);
      background: var(--bf-bg-0); color: var(--bf-text);
      display: grid; grid-template-columns: 220px 1fr;
    }
    aside {
      background: var(--bf-bg-1); border-right: 1px solid var(--bf-border);
      padding: 1.25rem 1rem; display: flex; flex-direction: column; gap: 1.5rem;
    }
    .brand { color: var(--bf-accent); font-weight: 700; letter-spacing: .04em; font-size: 1.1rem; }
    .brand span { color: var(--bf-muted); font-weight: 500; font-size: .75rem; display: block; margin-top: .2rem; }
    nav a {
      display: block; padding: .55rem .75rem; margin-bottom: .25rem;
      color: var(--bf-muted); text-decoration: none; border-radius: var(--bf-radius);
    }
    nav a:hover { background: var(--bf-bg-2); color: var(--bf-text); }
    nav a.active { background: var(--bf-bg-2); color: var(--bf-accent); }
    main { padding: 1.5rem 2rem; }
    h1 { margin: 0 0 .35rem; font-size: 1.5rem; }
    .lede { color: var(--bf-muted); margin: 0 0 1.5rem; }
    .panel {
      background: var(--bf-bg-1); border: 1px solid var(--bf-border);
      border-radius: var(--bf-radius); padding: 1.25rem;
    }
    .panel code { color: var(--bf-info); }
    .chips { display: flex; flex-wrap: wrap; gap: .5rem; margin: 1rem 0; }
    .chip {
      border: 1px solid var(--bf-border); border-radius: 999px;
      padding: .25rem .7rem; font-size: .8rem; color: var(--bf-muted); cursor: pointer;
    }
    .chip.on { border-color: var(--bf-accent); color: var(--bf-accent); }
    #printer-grid { display: grid; gap: .75rem; grid-template-columns: repeat(auto-fill, minmax(220px, 1fr)); }
    .printer-card {
      border: 1px solid var(--bf-border); border-radius: var(--bf-radius);
      padding: .85rem 1rem; background: var(--bf-bg-2);
    }
    .printer-card .name { font-weight: 600; }
    .printer-card .meta { color: var(--bf-muted); font-size: .85rem; margin-top: .35rem; }
    .state-RUNNING, .state-PREPARE { color: var(--bf-accent); }
    .state-PAUSE, .state-FAILED { color: var(--bf-danger); }
    .state-FINISH { color: var(--bf-warn); }
    .state-IDLE { color: var(--bf-muted); }
    #sse-status { font-size: .8rem; color: var(--bf-muted); margin-bottom: .75rem; }
    #sse-status.live { color: var(--bf-accent); }
    table { width: 100%; border-collapse: collapse; }
    th, td { text-align: left; padding: .5rem .4rem; border-bottom: 1px solid var(--bf-border); font-size: .9rem; }
    button.ctrl {
      background: var(--bf-bg-0); color: var(--bf-text); border: 1px solid var(--bf-border);
      border-radius: var(--bf-radius); padding: .2rem .5rem; margin-right: .25rem; cursor: pointer; font-size: .75rem;
    }
  </style>
</head>
<body>
  <aside>
    <div class="brand">BambuFarmer<span>LAN farm ops</span></div>
    <nav>
      <a href="/" {{if eq .Active "fleet"}}class="active"{{end}}>Fleet</a>
      <a href="/jobs" {{if eq .Active "jobs"}}class="active"{{end}}>Jobs</a>
      <a href="/presets" {{if eq .Active "presets"}}class="active"{{end}}>Presets</a>
      <a href="/analytics" {{if eq .Active "analytics"}}class="active"{{end}}>Analytics</a>
      <a href="/admin" {{if eq .Active "admin"}}class="active"{{end}}>Admin</a>
    </nav>
  </aside>
  <main>
    <h1>{{.Title}}</h1>
    {{if eq .Active "fleet"}}
      <p class="lede">Live fleet via gateway SSE — no browser polling of Farm Manager.</p>
      <div id="sse-status">connecting…</div>
      <div class="chips" id="bucket-chips">
        <span class="chip on" data-bucket="all">All</span>
        <span class="chip" data-bucket="ready_to_start">Ready to Start</span>
        <span class="chip" data-bucket="attention_needed">Attention Needed</span>
        <span class="chip" data-bucket="completed_task">Completed</span>
        <span class="chip" data-bucket="not_in_task">Not in Task</span>
        <span class="chip" data-bucket="printing">Printing</span>
      </div>
      <div id="printer-grid" class="panel"></div>
      <script>
        const API = {{printf "%q" .APIBase}};
        const printers = new Map();
        let filter = 'all';
        const statusEl = document.getElementById('sse-status');
        const grid = document.getElementById('printer-grid');

        document.querySelectorAll('#bucket-chips .chip').forEach(chip => {
          chip.addEventListener('click', () => {
            document.querySelectorAll('#bucket-chips .chip').forEach(c => c.classList.remove('on'));
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
            grid.innerHTML = '<p class="meta">Waiting for edge deltas (fixture or FM)…</p>';
            return;
          }
          grid.innerHTML = items.map(p => {
            const sn = p.external_sn;
            return '<div class="printer-card" id="printer-' + sn + '">' +
              '<div class="name">' + (p.name || sn) + '</div>' +
              '<div class="meta">' + (p.model || '') + ' · <span class="state-' + (p.gcode_state||'') + '">' + (p.gcode_state||'?') + '</span>' +
              (p.mc_percent ? ' · ' + p.mc_percent + '%' : '') + '</div>' +
              '<div class="meta" style="margin-top:.6rem">' +
              '<button class="ctrl" data-sn="' + sn + '" data-opt="pause">Pause</button>' +
              '<button class="ctrl" data-sn="' + sn + '" data-opt="resume">Resume</button>' +
              '<button class="ctrl" data-sn="' + sn + '" data-opt="stop">Stop</button>' +
              '<button class="ctrl" data-sn="' + sn + '" data-opt="bed_clean">Bed clean</button>' +
              '</div></div>';
          }).join('');
          grid.querySelectorAll('button.ctrl').forEach(btn => {
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
      <p class="lede">Job queue via core. Create jobs with curl or wire a form next.</p>
      <div class="panel">
        <p><code>GET {{.APIBase}}/api/v1/jobs</code></p>
        <p><code>POST {{.APIBase}}/api/v1/jobs</code> with <code>{"name":"demo","external_sn":"…","est_seconds":3600,"est_grams":100}</code></p>
        <pre id="jobs-list">loading…</pre>
      </div>
      <script>
        fetch({{printf "%q" .APIBase}} + '/api/v1/jobs').then(r => r.json()).then(j => {
          document.getElementById('jobs-list').textContent = JSON.stringify(j, null, 2);
        }).catch(e => { document.getElementById('jobs-list').textContent = String(e); });
      </script>
    {{else if eq .Active "presets"}}
      <p class="lede">Versioned filament + print presets in Postgres.</p>
      <div class="panel">
        <p>Export: <code>curl {{.APIBase}}/api/v1/presets/export</code></p>
        <pre id="presets-export">loading…</pre>
      </div>
      <script>
        fetch({{printf "%q" .APIBase}} + '/api/v1/presets/export').then(r => r.json()).then(j => {
          document.getElementById('presets-export').textContent = JSON.stringify(j, null, 2);
        }).catch(e => { document.getElementById('presets-export').textContent = String(e); });
      </script>
    {{else if eq .Active "analytics"}}
      <p class="lede">Utilization + event counters from analytics consumer.</p>
      <div class="panel">
        <pre id="analytics-out">loading…</pre>
      </div>
      <script>
        Promise.all([
          fetch({{printf "%q" .APIBase}} + '/api/v1/analytics/utilization').then(r => r.json()),
          fetch({{printf "%q" .APIBase}} + '/api/v1/analytics/counters').then(r => r.json())
        ]).then(([u, c]) => {
          document.getElementById('analytics-out').textContent = JSON.stringify({utilization: u, counters: c}, null, 2);
        }).catch(e => { document.getElementById('analytics-out').textContent = String(e); });
      </script>
    {{else}}
      <p class="lede">Site, rate cards, auth mode.</p>
      <div class="panel">
        <p>Auth: see <code>docs/DEPLOYMENT_MODES.md</code> (LAN vs OIDC).</p>
        <pre id="admin-out">loading…</pre>
      </div>
      <script>
        Promise.all([
          fetch({{printf "%q" .APIBase}} + '/api/v1/auth/me').then(r => r.json()),
          fetch({{printf "%q" .APIBase}} + '/api/v1/admin/rate-cards').then(r => r.json())
        ]).then(([me, rates]) => {
          document.getElementById('admin-out').textContent = JSON.stringify({me, rates}, null, 2);
        }).catch(e => { document.getElementById('admin-out').textContent = String(e); });
      </script>
    {{end}}
  </main>
</body>
</html>
`