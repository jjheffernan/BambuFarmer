package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/jjheffernan/BambuFarmer/packages/httpserver"
)

func main() {
	service := "edge"
	mux := httpserver.NewMux(service)
	mux.HandleFunc("GET /api/v1/edge/status", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"site_id":             getenv("SITE_ID", "default"),
			"farm_mode":           getenv("FARM_MODE", "farm_manager"),
			"farm_manager_url":    getenv("FARM_MANAGER_URL", ""),
			"transport":           getenv("FARM_MANAGER_TRANSPORT", "fm_native"),
			"connected":           false,
			"note":                "scaffold stub — FM-native adapter not wired yet (CDP escape hatch supported)",
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

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
