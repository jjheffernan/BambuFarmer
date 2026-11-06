package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/jjheffernan/BambuFarmer/packages/httpserver"
)

func main() {
	service := "integrations"
	spoolman := os.Getenv("SPOOLMAN_URL")
	manyfold := os.Getenv("MANYFOLD_URL")

	mux := httpserver.NewMux(service)
	mux.HandleFunc("GET /api/v1/integrations/status", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]any{
			"spoolman": statusURL(spoolman),
			"manyfold": statusURL(manyfold),
			"mode":     "optional_live_url",
		})
	})
	mux.HandleFunc("GET /api/v1/filament/spools", func(w http.ResponseWriter, _ *http.Request) {
		if spoolman == "" {
			writeJSON(w, map[string]any{"configured": false, "spools": []any{}})
			return
		}
		// Live client lands later — empty list keeps UI/API contract stable.
		writeJSON(w, map[string]any{"configured": true, "spools": []any{}, "note": "Spoolman client stub"})
	})
	mux.HandleFunc("GET /api/v1/library/models", func(w http.ResponseWriter, _ *http.Request) {
		if manyfold == "" {
			writeJSON(w, map[string]any{"configured": false, "models": []any{}})
			return
		}
		writeJSON(w, map[string]any{"configured": true, "models": []any{}, "note": "Manyfold client stub"})
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

func statusURL(u string) map[string]any {
	if u == "" {
		return map[string]any{"configured": false}
	}
	return map[string]any{"configured": true, "url": u}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
