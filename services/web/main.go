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
        <a href="/presets" {{if eq .Active "presets"}}class="active"{{end}}>Presets</a>
        <a href="/analytics" {{if eq .Active "analytics"}}class="active"{{end}}>Analytics</a>
        <a href="/admin" {{if eq .Active "admin"}}class="active"{{end}}>Admin</a>
      </nav>
      <div class="mt-auto text-xs text-bf-muted space-y-2">
        <p id="bf-user-label" class="hidden"></p>
        <button type="button" id="bf-logout" class="bf-btn hidden w-full">Log out</button>
        <p class="hidden md:block">Gateway API<br/><code class="text-[0.7rem] break-all">{{.APIBase}}</code></p>
      </div>
    </aside>
    <script>
      (async () => {
        const API = {{printf "%q" .APIBase}};
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
                  credentials: 'include',
                  headers: {'Content-Type': 'application/json'},
                  body: JSON.stringify({opt: btn.dataset.opt})
                });
              });
            });
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
          bfJSON(API + '/api/v1/jobs').then(jobs => {
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
            bfJSON(API + '/api/v1/presets/filament'),
            bfJSON(API + '/api/v1/presets/print')
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
            bfJSON(API + '/api/v1/analytics/utilization'),
            bfJSON(API + '/api/v1/analytics/counters')
          ]).then(([u, c]) => {
            document.getElementById('analytics-out').textContent = JSON.stringify({utilization: u, counters: c}, null, 2);
          }).catch(e => { document.getElementById('analytics-out').textContent = String(e); });
        </script>
      {{else}}
        <p class="bf-lede">Local accounts, password reset, rate cards, and auth mode.</p>
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
              <h2 class="font-display text-lg mb-3">Session / rates</h2>
              <pre class="bf-pre" id="admin-out">loading…</pre>
            </div>
          </div>
        </div>
        <script>
          const API = {{printf "%q" .APIBase}};
          const opts = { credentials: 'include' };
          async function loadUsers() {
            const users = await bfJSON(API + '/api/v1/admin/users', opts);
            const el = document.getElementById('users-list');
            if (!users || !users.length) { el.innerHTML = '<p class="bf-empty">No users</p>'; return; }
            el.innerHTML = '<ul class="space-y-2">' + users.map(u =>
              '<li class="bf-card">' +
              '<div class="bf-card-name">' + u.username + (u.active ? '' : ' (disabled)') + '</div>' +
              '<div class="bf-card-meta">' + (u.roles||[]).join(', ') + '</div>' +
              '<div class="flex flex-wrap gap-1.5 mt-2 items-center">' +
              '<input class="bf-input !w-auto min-w-[8rem] text-xs" type="password" placeholder="new password" data-reset-pw="' + u.id + '" minlength="8"/>' +
              '<button type="button" class="bf-btn" data-reset="' + u.id + '">Reset password</button>' +
              '</div></li>'
            ).join('') + '</ul>';
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
          Promise.all([
            bfJSON(API + '/api/v1/auth/me', opts),
            bfJSON(API + '/api/v1/admin/rate-cards', opts).catch(e => ({error: String(e)})),
            loadUsers().catch(e => { document.getElementById('users-list').textContent = String(e); })
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
