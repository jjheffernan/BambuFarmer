package main

import (
	"html/template"
	"net/http"
)

func registerSetup(mux *http.ServeMux, apiBase string) {
	tpl := template.Must(template.New("setup").Parse(setupHTML))
	mux.HandleFunc("GET /setup", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = tpl.Execute(w, pageData{Title: "Setup", Active: "setup", APIBase: apiBase})
	})
}

const setupHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1"/>
  <title>Install · BambuFarmer</title>
  <style>
    :root {
      --bf-bg-0: #0e1114; --bf-bg-1: #161b20; --bf-bg-2: #1e252c;
      --bf-text: #e8eef2; --bf-muted: #8b9aab; --bf-accent: #00ae42;
      --bf-warn: #f0a202; --bf-border: #2a343d; --bf-radius: 6px;
      --bf-font-ui: "IBM Plex Sans", "Segoe UI", sans-serif;
      --bf-info: #3aa0ff;
    }
    * { box-sizing: border-box; }
    body {
      margin: 0; min-height: 100vh; font-family: var(--bf-font-ui);
      background: radial-gradient(1200px 600px at 10% -10%, #1a2a22 0%, var(--bf-bg-0) 55%);
      color: var(--bf-text); display: flex; align-items: flex-start; justify-content: center;
      padding: 2.5rem 1rem;
    }
    .shell { width: min(720px, 100%); }
    .brand { color: var(--bf-accent); font-weight: 700; font-size: 1.35rem; letter-spacing: .04em; }
    .brand span { display: block; color: var(--bf-muted); font-size: .8rem; font-weight: 500; margin-top: .25rem; }
    h1 { margin: 1.25rem 0 .35rem; font-size: 1.6rem; }
    .lede { color: var(--bf-muted); margin: 0 0 1.25rem; }
    .steps { display: flex; flex-wrap: wrap; gap: .4rem; margin-bottom: 1rem; }
    .step-pill {
      border: 1px solid var(--bf-border); border-radius: 999px; padding: .2rem .65rem;
      font-size: .75rem; color: var(--bf-muted);
    }
    .step-pill.on { border-color: var(--bf-accent); color: var(--bf-accent); }
    .step-pill.done { color: var(--bf-text); border-color: #3a4a3f; }
    .panel {
      background: var(--bf-bg-1); border: 1px solid var(--bf-border);
      border-radius: var(--bf-radius); padding: 1.25rem 1.35rem;
    }
    label { display: block; font-size: .8rem; color: var(--bf-muted); margin: .85rem 0 .3rem; }
    input, select {
      width: 100%; padding: .55rem .65rem; border-radius: var(--bf-radius);
      border: 1px solid var(--bf-border); background: var(--bf-bg-0); color: var(--bf-text);
    }
    .row { display: grid; grid-template-columns: 1fr 1fr; gap: .75rem; }
    @media (max-width: 600px) { .row { grid-template-columns: 1fr; } }
    .hint { font-size: .8rem; color: var(--bf-muted); margin-top: .75rem; line-height: 1.4; }
    .hint code { color: var(--bf-info); }
    .actions { display: flex; gap: .5rem; margin-top: 1.25rem; flex-wrap: wrap; }
    button {
      border: 1px solid var(--bf-border); background: var(--bf-bg-2); color: var(--bf-text);
      border-radius: var(--bf-radius); padding: .55rem 1rem; cursor: pointer; font: inherit;
    }
    button.primary { background: var(--bf-accent); color: #04140a; border-color: var(--bf-accent); font-weight: 600; }
    button:disabled { opacity: .5; cursor: not-allowed; }
    .err { color: #e23d28; font-size: .85rem; margin-top: .75rem; }
    .ok { color: var(--bf-accent); }
    .check { display: flex; align-items: center; gap: .5rem; margin-top: 1rem; color: var(--bf-muted); font-size: .9rem; }
    .check input { width: auto; }
    .ver { font-size: .75rem; color: var(--bf-muted); margin-top: 1.5rem; }
  </style>
</head>
<body>
  <div class="shell">
    <div class="brand">BambuFarmer<span>First-run install wizard</span></div>
    <h1 id="title">Welcome</h1>
    <p class="lede" id="lede">Configure this farm install. Secrets like passwords stay in <code>.env</code>.</p>
    <div class="steps" id="pills"></div>
    <div class="panel" id="panel"></div>
    <p class="ver" id="ver"></p>
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
        let cls = 'step-pill';
        if (i === idx) cls += ' on';
        else if (i < idx) cls += ' done';
        return '<span class="' + cls + '">' + (i+1) + '. ' + s.label + '</span>';
      }).join('');
    }

    function field(label, id, value, type) {
      type = type || 'text';
      return '<label for="' + id + '">' + label + '</label><input id="' + id + '" type="' + type + '" value="' +
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
          '<p>You will set site name, auth mode (<strong>lan</strong> vs <strong>oidc</strong>), costing rates, and optional Farm Manager / Spoolman / Manyfold URLs.</p>' +
          '<p class="hint">LAN mode is for private networks only. See <code>docs/DEPLOYMENT_MODES.md</code>.</p>' +
          actions(true, false);
      } else if (s.id === 'site') {
        title.textContent = 'Site';
        lede.textContent = 'Single-site MVP — one Farm Manager edge.';
        panel.innerHTML =
          field('Site name', 'site_name', state.site_name) +
          field('Timezone (IANA)', 'timezone', state.timezone) +
          field('Admin username (LAN)', 'admin_username', state.admin_username) +
          '<p class="hint">Admin password is not stored here — use LAN login stub for now; real user store lands next.</p>' +
          actions(true, true);
      } else if (s.id === 'auth') {
        title.textContent = 'Auth mode';
        lede.textContent = 'Choose how operators sign in.';
        panel.innerHTML =
          '<label for="auth_mode">AUTH_MODE</label><select id="auth_mode">' +
          '<option value="lan"' + (state.auth_mode==='lan'?' selected':'') + '>lan — insecure LAN (default)</option>' +
          '<option value="oidc"' + (state.auth_mode==='oidc'?' selected':'') + '>oidc — URL / SSO deploy</option>' +
          '</select>' +
          '<p class="hint">Set <code>AUTH_MODE</code> and <code>OIDC_*</code> in <code>.env</code> to match. Wizard records intent; gateway reads env at boot.</p>' +
          actions(true, true);
      } else if (s.id === 'rates') {
        title.textContent = 'Costing';
        lede.textContent = '$/hr · $/kg · kg/hr rate card.';
        const rc = state.rate_card;
        panel.innerHTML =
          '<div class="row">' +
          field('Currency', 'currency', rc.currency) +
          field('Machine $/hr', 'machine_rate_per_hour', rc.machine_rate_per_hour, 'number') +
          '</div><div class="row">' +
          field('Filament $/kg', 'filament_price_per_kg', rc.filament_price_per_kg, 'number') +
          field('Filament kg/hr', 'filament_kg_per_hour', rc.filament_kg_per_hour, 'number') +
          '</div>' + actions(true, true);
      } else if (s.id === 'farm') {
        title.textContent = 'Farm Manager';
        lede.textContent = 'Optional now — edge can run fixtures until FM is reachable.';
        panel.innerHTML =
          field('Farm Manager URL', 'farm_manager_url', state.farm_manager_url) +
          '<label for="farm_manager_transport">Transport</label><select id="farm_manager_transport">' +
          '<option value="fm_native"' + (state.farm_manager_transport==='fm_native'?' selected':'') + '>fm_native</option>' +
          '<option value="cdp"' + (state.farm_manager_transport==='cdp'?' selected':'') + '>cdp (escape hatch)</option>' +
          '</select>' +
          '<p class="hint">Put <code>FARM_MANAGER_USER</code> / <code>FARM_MANAGER_PASSWORD</code> in <code>.env</code> on the edge host. Leave URL empty for fixture mode.</p>' +
          actions(true, true);
      } else if (s.id === 'integrations') {
        title.textContent = 'Integrations';
        lede.textContent = 'Optional Spoolman / Manyfold live URLs.';
        panel.innerHTML =
          field('Spoolman URL', 'spoolman_url', state.spoolman_url) +
          field('Manyfold URL', 'manyfold_url', state.manyfold_url) +
          '<p class="hint">Tokens/secrets stay in <code>.env</code>. Empty URL disables the adapter.</p>' +
          actions(true, true);
      } else {
        title.textContent = 'Finish';
        lede.textContent = 'Write config and open the fleet UI.';
        panel.innerHTML =
          '<div class="check"><input type="checkbox" id="import_example_presets"' + (state.import_example_presets?' checked':'') + '/>' +
          '<label for="import_example_presets" style="margin:0">Import starter filament/print presets</label></div>' +
          '<p class="hint">After finish, set matching env vars and restart gateway/edge if you changed auth or FM URL.</p>' +
          '<div id="msg"></div>' + actions(false, true, true);
      }

      bindActions();
    }

    function actions(next, back, finish) {
      let html = '<div class="actions">';
      if (back) html += '<button type="button" id="back">Back</button>';
      if (next) html += '<button type="button" class="primary" id="next">Continue</button>';
      if (finish) html += '<button type="button" class="primary" id="finish">Complete setup</button>';
      html += '</div><div class="err" id="err"></div>';
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
          if (msg) msg.innerHTML = '<p class="ok">Setup complete. Opening fleet…</p>';
          setTimeout(() => { location.href = '/'; }, 600);
        } catch (e) {
          err.textContent = String(e.message || e);
          finish.disabled = false;
        }
      };
    }

    load().catch(e => {
      document.getElementById('panel').innerHTML =
        '<p class="err">Could not load setup API. Is core/Postgres up?<br/>' + String(e) + '</p>';
    });
  </script>
</body>
</html>
`