package setup

import (
	"html/template"
	"net/http"
)

// PageData is the install-wizard template model.
type PageData struct {
	Title   string
	Active  string
	APIBase string
}

// Register mounts GET /setup (first-run install wizard).
func Register(mux *http.ServeMux, apiBase string) {
	tpl := template.Must(template.New("setup").Parse(setupHTML))
	mux.HandleFunc("GET /setup", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = tpl.Execute(w, PageData{Title: "Setup", Active: "setup", APIBase: apiBase})
	})
}

const setupHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1"/>
  <title>Install · BambuFarmer</title>
  <link rel="preconnect" href="https://fonts.googleapis.com"/>
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin/>
  <link href="https://fonts.googleapis.com/css2?family=IBM+Plex+Mono:wght@400;500&family=IBM+Plex+Sans:wght@400;500;600;700&family=Syne:wght@600;700&display=swap" rel="stylesheet"/>
  <link rel="stylesheet" href="/static/css/app.css"/>
</head>
<body class="bf-wizard">
  <div class="bf-wizard-shell">
    <div class="bf-brand text-[1.35rem]">BambuFarmer<span>First-run install wizard</span></div>
    <h1 class="bf-title mt-5" id="title">Welcome</h1>
    <p class="bf-lede" id="lede">Configure this farm install. Secrets like passwords stay in <code>.env</code>.</p>
    <div class="flex flex-wrap gap-1.5 mb-4" id="pills"></div>
    <div class="bf-panel" id="panel"></div>
    <p class="text-xs text-bf-muted mt-6" id="ver"></p>
  </div>
  <script>
    const API = {{printf "%q" .APIBase}};
    const STEPS = [
      { id: 'welcome', label: 'Welcome' },
      { id: 'site', label: 'Site' },
      { id: 'auth', label: 'Auth' },
      { id: 'rates', label: 'Costing' },
      { id: 'farm', label: 'Farm Manager' },
      { id: 'integrations', label: 'Integrations' },
      { id: 'finish', label: 'Finish' },
    ];
    let idx = 0;
    let state = {
      site_name: 'Default Site',
      timezone: 'UTC',
      auth_mode: 'lan',
      farm_manager_url: '',
      farm_manager_transport: 'fm_native',
      spoolman_url: '',
      manyfold_url: '',
      admin_username: 'admin',
      import_example_presets: true,
      rate_card: {
        currency: 'USD',
        machine_rate_per_hour: 0,
        filament_price_per_kg: 0,
        filament_kg_per_hour: 0,
        labor_flat: 0,
        failure_surcharge_pct: 0,
      },
    };
    let version = {};

    async function load() {
      const r = await fetch(API + '/api/v1/setup');
      if (!r.ok) throw new Error('setup status ' + r.status);
      const data = await r.json();
      version = data.version || {};
      const inst = data.install || {};
      if (inst.completed) {
        location.href = '/';
        return;
      }
      Object.assign(state, {
        site_name: inst.site_name || state.site_name,
        timezone: inst.timezone || state.timezone,
        auth_mode: inst.auth_mode || state.auth_mode,
        farm_manager_url: inst.farm_manager_url || '',
        farm_manager_transport: inst.farm_manager_transport || 'fm_native',
        spoolman_url: inst.spoolman_url || '',
        manyfold_url: inst.manyfold_url || '',
        admin_username: inst.admin_username || 'admin',
      });
      if (data.rate_card) state.rate_card = Object.assign(state.rate_card, data.rate_card);
      const step = inst.step;
      const found = STEPS.findIndex(s => s.id === step);
      if (found > 0) idx = found;
      render();
    }

    function pills() {
      return STEPS.map((s, i) => {
        let cls = 'bf-step';
        if (i === idx) cls += ' on';
        else if (i < idx) cls += ' done';
        return '<span class="' + cls + '">' + (i+1) + '. ' + s.label + '</span>';
      }).join('');
    }

    function field(label, id, value, type) {
      type = type || 'text';
      return '<label class="bf-field" for="' + id + '">' + label + '</label><input class="bf-input" id="' + id + '" type="' + type + '" value="' +
        String(value ?? '').replace(/"/g, '&quot;') + '"/>';
    }

    function render() {
      document.getElementById('pills').innerHTML = pills();
      document.getElementById('ver').textContent =
        'BambuFarmer ' + (version.version || '') + (version.commit && version.commit !== 'unknown' ? ' · ' + version.commit.slice(0,7) : '');
      const panel = document.getElementById('panel');
      const title = document.getElementById('title');
      const lede = document.getElementById('lede');
      const s = STEPS[idx];

      if (s.id === 'welcome') {
        title.textContent = 'Welcome';
        lede.textContent = 'This wizard configures site defaults. Put SESSION_SECRET, DB, and FM passwords in .env.';
        panel.innerHTML =
          '<p class="text-bf-text leading-relaxed">You will set site name, auth mode (<strong>lan</strong> vs <strong>oidc</strong>), costing rates, and optional Farm Manager / Spoolman / Manyfold URLs.</p>' +
          '<p class="text-sm text-bf-muted mt-3 leading-relaxed">LAN mode is for private networks only. See <code>docs/DEPLOYMENT_MODES.md</code>.</p>' +
          actions(true, false);
      } else if (s.id === 'site') {
        title.textContent = 'Site';
        lede.textContent = 'Single-site MVP — one Farm Manager edge.';
        panel.innerHTML =
          field('Site name', 'site_name', state.site_name) +
          field('Timezone (IANA)', 'timezone', state.timezone) +
          field('Admin username (LAN)', 'admin_username', state.admin_username) +
          '<p class="text-sm text-bf-muted mt-3 leading-relaxed">Admin password is not stored here — use LAN login stub for now; real user store lands next.</p>' +
          actions(true, true);
      } else if (s.id === 'auth') {
        title.textContent = 'Auth mode';
        lede.textContent = 'Choose how operators sign in.';
        panel.innerHTML =
          '<label class="bf-field" for="auth_mode">AUTH_MODE</label><select class="bf-input" id="auth_mode">' +
          '<option value="lan"' + (state.auth_mode==='lan'?' selected':'') + '>lan — insecure LAN (default)</option>' +
          '<option value="oidc"' + (state.auth_mode==='oidc'?' selected':'') + '>oidc — URL / SSO deploy</option>' +
          '</select>' +
          '<p class="text-sm text-bf-muted mt-3 leading-relaxed">Set <code>AUTH_MODE</code> and <code>OIDC_*</code> in <code>.env</code> to match. Wizard records intent; gateway reads env at boot.</p>' +
          actions(true, true);
      } else if (s.id === 'rates') {
        title.textContent = 'Costing';
        lede.textContent = '$/hr · $/kg · kg/hr rate card.';
        const rc = state.rate_card;
        panel.innerHTML =
          '<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">' +
          field('Currency', 'currency', rc.currency) +
          field('Machine $/hr', 'machine_rate_per_hour', rc.machine_rate_per_hour, 'number') +
          field('Filament $/kg', 'filament_price_per_kg', rc.filament_price_per_kg, 'number') +
          field('Filament kg/hr', 'filament_kg_per_hour', rc.filament_kg_per_hour, 'number') +
          '</div>' + actions(true, true);
      } else if (s.id === 'farm') {
        title.textContent = 'Farm Manager';
        lede.textContent = 'Optional now — edge can run fixtures until FM is reachable.';
        panel.innerHTML =
          field('Farm Manager URL', 'farm_manager_url', state.farm_manager_url) +
          '<label class="bf-field" for="farm_manager_transport">Transport</label><select class="bf-input" id="farm_manager_transport">' +
          '<option value="fm_native"' + (state.farm_manager_transport==='fm_native'?' selected':'') + '>fm_native</option>' +
          '<option value="cdp"' + (state.farm_manager_transport==='cdp'?' selected':'') + '>cdp (escape hatch)</option>' +
          '</select>' +
          '<p class="text-sm text-bf-muted mt-3 leading-relaxed">Put <code>FARM_MANAGER_USER</code> / <code>FARM_MANAGER_PASSWORD</code> in <code>.env</code> on the edge host. Leave URL empty for fixture mode.</p>' +
          actions(true, true);
      } else if (s.id === 'integrations') {
        title.textContent = 'Integrations';
        lede.textContent = 'Optional Spoolman / Manyfold live URLs.';
        panel.innerHTML =
          field('Spoolman URL', 'spoolman_url', state.spoolman_url) +
          field('Manyfold URL', 'manyfold_url', state.manyfold_url) +
          '<p class="text-sm text-bf-muted mt-3 leading-relaxed">Tokens/secrets stay in <code>.env</code>. Empty URL disables the adapter.</p>' +
          actions(true, true);
      } else {
        title.textContent = 'Finish';
        lede.textContent = 'Write config and open the fleet UI.';
        panel.innerHTML =
          '<div class="flex items-center gap-2 mt-1 text-sm text-bf-muted"><input type="checkbox" id="import_example_presets"' + (state.import_example_presets?' checked':'') + '/>' +
          '<label for="import_example_presets" class="m-0">Import starter filament/print presets</label></div>' +
          '<p class="text-sm text-bf-muted mt-3 leading-relaxed">After finish, set matching env vars and restart gateway/edge if you changed auth or FM URL.</p>' +
          '<div id="msg"></div>' + actions(false, true, true);
      }

      bindActions();
    }

    function actions(next, back, finish) {
      let html = '<div class="bf-actions">';
      if (back) html += '<button type="button" class="bf-btn" id="back">Back</button>';
      if (next) html += '<button type="button" class="bf-btn bf-btn-primary" id="next">Continue</button>';
      if (finish) html += '<button type="button" class="bf-btn bf-btn-primary" id="finish">Complete setup</button>';
      html += '</div><div class="text-bf-danger text-sm mt-3" id="err"></div>';
      return html;
    }

    function readForm() {
      const g = id => document.getElementById(id);
      if (g('site_name')) state.site_name = g('site_name').value.trim();
      if (g('timezone')) state.timezone = g('timezone').value.trim() || 'UTC';
      if (g('admin_username')) state.admin_username = g('admin_username').value.trim() || 'admin';
      if (g('auth_mode')) state.auth_mode = g('auth_mode').value;
      if (g('currency')) {
        state.rate_card.currency = g('currency').value || 'USD';
        state.rate_card.machine_rate_per_hour = Number(g('machine_rate_per_hour').value || 0);
        state.rate_card.filament_price_per_kg = Number(g('filament_price_per_kg').value || 0);
        state.rate_card.filament_kg_per_hour = Number(g('filament_kg_per_hour').value || 0);
      }
      if (g('farm_manager_url')) {
        state.farm_manager_url = g('farm_manager_url').value.trim();
        state.farm_manager_transport = g('farm_manager_transport').value;
      }
      if (g('spoolman_url')) {
        state.spoolman_url = g('spoolman_url').value.trim();
        state.manyfold_url = g('manyfold_url').value.trim();
      }
      if (g('import_example_presets')) state.import_example_presets = g('import_example_presets').checked;
    }

    async function save(complete) {
      readForm();
      const body = Object.assign({}, state, {
        step: STEPS[idx].id,
        complete: !!complete,
      });
      const r = await fetch(API + '/api/v1/setup', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify(body),
      });
      if (!r.ok) throw new Error(await r.text());
      return r.json();
    }

    function bindActions() {
      const err = document.getElementById('err');
      const back = document.getElementById('back');
      const next = document.getElementById('next');
      const finish = document.getElementById('finish');
      if (back) back.onclick = () => { idx = Math.max(0, idx - 1); render(); };
      if (next) next.onclick = async () => {
        try {
          err.textContent = '';
          await save(false);
          idx = Math.min(STEPS.length - 1, idx + 1);
          render();
        } catch (e) { err.textContent = String(e.message || e); }
      };
      if (finish) finish.onclick = async () => {
        try {
          err.textContent = '';
          finish.disabled = true;
          await save(true);
          const msg = document.getElementById('msg');
          if (msg) msg.innerHTML = '<p class="text-bf-accent mt-3">Setup complete. Opening fleet…</p>';
          setTimeout(() => { location.href = '/'; }, 600);
        } catch (e) {
          err.textContent = String(e.message || e);
          finish.disabled = false;
        }
      };
    }

    load().catch(e => {
      document.getElementById('panel').innerHTML =
        '<p class="text-bf-danger">Could not load setup API. Is core/Postgres up?<br/>' + String(e) + '</p>';
    });
  </script>
</body>
</html>
`