// Command commerce is the optional shop-orders plugin (compose profile: commerce).
// Store webhook queues a core job when CORE_URL is set.
package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/jjheffernan/BambuFarmer/packages/httpserver"
	"github.com/jjheffernan/BambuFarmer/services/commerce/internal/shop"
)

func main() {
	service := "commerce"
	mux := httpserver.NewMux(service)
	shop.Register(mux, os.Getenv("CORE_URL"))
	mux.HandleFunc("GET /api/v1/commerce/status", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "stub",
			"service": service,
			"note":    "POST /api/v1/commerce/webhook queues a core job when CORE_URL is set; POST /api/v1/commerce/orders/status echoes packed or shipped",
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
