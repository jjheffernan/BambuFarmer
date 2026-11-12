// Command commerce is the optional shop-orders plugin (compose profile: commerce).
// Stub: health + status only. Wire store webhooks → jobs → pack/ship in build-out P4.
package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/jjheffernan/BambuFarmer/packages/httpserver"
)

func main() {
	service := "commerce"
	mux := httpserver.NewMux(service)
	mux.HandleFunc("GET /api/v1/commerce/status", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "stub",
			"service": service,
			"note":    "enable with compose --profile commerce; order APIs land in build-out P4",
		})
	})

	if err := httpserver.ListenAndServe(httpserver.Options{
		Service: service,
		Addr:    getenv("HTTP_ADDR", ":8080"),
		Mux:     mux,
	}); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
