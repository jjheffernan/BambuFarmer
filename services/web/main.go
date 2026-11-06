package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/jjheffernan/BambuFarmer/packages/httpserver"
)

func main() {
	service := "web"
	mux := httpserver.NewMux(service)
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(indexHTML))
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

const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1"/>
  <title>BambuFarmer</title>
  <style>
    :root {
      --bf-bg-0: #0e1114;
      --bf-bg-1: #161b20;
      --bf-text: #e8eef2;
      --bf-muted: #8b9aab;
      --bf-accent: #00ae42;
      --bf-border: #2a343d;
    }
    body {
      margin: 0; font-family: "IBM Plex Sans", "Segoe UI", sans-serif;
      background: var(--bf-bg-0); color: var(--bf-text);
      min-height: 100vh; display: grid; place-items: center;
    }
    main {
      background: var(--bf-bg-1); border: 1px solid var(--bf-border);
      border-radius: 6px; padding: 2rem 2.5rem; max-width: 36rem;
    }
    h1 { margin: 0 0 .5rem; color: var(--bf-accent); letter-spacing: .02em; }
    p { margin: 0; color: var(--bf-muted); line-height: 1.5; }
    code { color: var(--bf-text); }
  </style>
</head>
<body>
  <main>
    <h1>BambuFarmer</h1>
    <p>Scaffold shell — Go + HTMX UI lands next. See <code>docs/DESIGN_SPEC.md</code> and <code>docs/MVP_CHECKLIST.md</code>.</p>
  </main>
</body>
</html>
`