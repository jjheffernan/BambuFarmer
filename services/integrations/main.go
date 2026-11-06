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
	mux := httpserver.NewMux(service)
	mux.HandleFunc("GET /api/v1/integrations/status", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"spoolman":  statusURL(os.Getenv("SPOOLMAN_URL")),
			"manyfold":  statusURL(os.Getenv("MANYFOLD_URL")),
			"mode":      "stub_until_configured",
		})
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
