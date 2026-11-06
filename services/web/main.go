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
    .chips { display: flex; flex-wrap: wrap; gap: .5rem; margin-top: 1rem; }
    .chip {
      border: 1px solid var(--bf-border); border-radius: 999px;
      padding: .25rem .7rem; font-size: .8rem; color: var(--bf-muted);
    }
    .chip.on { border-color: var(--bf-accent); color: var(--bf-accent); }
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
    <p class="lede">Scaffold UI — live FM fleet and HTMX fragments land next.</p>
    <div class="panel">
      {{if eq .Active "fleet"}}
        <p>Printer grid will stream from gateway SSE (<code>{{.APIBase}}/api/v1/events/stream</code>).</p>
        <div class="chips">
          <span class="chip on">Ready to Start</span>
          <span class="chip">Attention Needed</span>
          <span class="chip">Completed</span>
          <span class="chip">Not in Task</span>
        </div>
      {{else if eq .Active "jobs"}}
        <p>Job queue / composer will call core via gateway. Circuit limits are Phase 2.</p>
      {{else if eq .Active "presets"}}
        <p>Versioned filament + print presets persist in core Postgres.</p>
        <p>Try: <code>curl {{.APIBase}}/api/v1/presets/export</code></p>
      {{else if eq .Active "analytics"}}
        <p>Dedicated analytics module — utilization, fail rates, $/hr · $/kg · kg/hr.</p>
      {{else}}
        <p>Site, LAN auth users, rate cards, and adapter URLs.</p>
        <p>Auth mode default: <code>lan</code> (insecure private network).</p>
      {{end}}
    </div>
  </main>
</body>
</html>
`