package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/jjheffernan/BambuFarmer/packages/httpserver"
)

func main() {
	service := "gateway"
	authMode := normalizeAuthMode(getenv("AUTH_MODE", "lan"))
	coreURL := mustURL(getenv("CORE_URL", "http://core:8080"))
	webURL := mustURL(getenv("WEB_URL", "http://web:8080"))
	analyticsURL := mustURL(getenv("ANALYTICS_URL", "http://analytics:8080"))
	integrationsURL := mustURL(getenv("INTEGRATIONS_URL", "http://integrations:8080"))
	edgeURL := mustURL(getenv("EDGE_URL", "http://edge:8080"))

	coreProxy := httputil.NewSingleHostReverseProxy(coreURL)
	webProxy := httputil.NewSingleHostReverseProxy(webURL)
	analyticsProxy := httputil.NewSingleHostReverseProxy(analyticsURL)
	integrationsProxy := httputil.NewSingleHostReverseProxy(integrationsURL)
	edgeProxy := httputil.NewSingleHostReverseProxy(edgeURL)

	mux := httpserver.NewMux(service)
	mux.HandleFunc("GET /api/v1/auth/me", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]any{
			"auth_mode": authMode,
			"user":      "stub",
			"roles":     []string{"admin"},
			"warning":   lanWarning(authMode),
		})
	})
	mux.HandleFunc("POST /api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if authMode != "lan" {
			http.Error(w, "local login only in AUTH_MODE=lan", http.StatusBadRequest)
			return
		}
		var body struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.Username == "" {
			http.Error(w, "username required", http.StatusBadRequest)
			return
		}
		// Stub: accept any non-empty user in LAN mode until real user store lands.
		writeJSON(w, map[string]any{
			"token":     "lan-stub-token",
			"auth_mode": "lan",
			"user":      body.Username,
		})
	})

	mux.Handle("/api/v1/presets/", coreProxy)
	mux.Handle("/api/v1/analytics/", analyticsProxy)
	mux.Handle("/api/v1/integrations/", integrationsProxy)
	mux.Handle("/api/v1/edge/", edgeProxy)

	// UI through gateway so operators hit one port when desired.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		webProxy.ServeHTTP(w, r)
	})

	if authMode == "lan" {
		slog.Warn("AUTH_MODE=lan (insecure LAN mode) — bind to private networks only")
	}

	srv := &http.Server{
		Addr:              httpserver.DefaultAddr(),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	slog.Info("listening", "service", service, "addr", srv.Addr, "auth_mode", authMode)
	if err := srv.ListenAndServe(); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

func lanWarning(mode string) string {
	if mode == "lan" {
		return "insecure LAN mode — do not expose publicly"
	}
	return ""
}

func normalizeAuthMode(mode string) string {
	if mode == "local_insecure" {
		return "lan"
	}
	return mode
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func mustURL(raw string) *url.URL {
	u, err := url.Parse(raw)
	if err != nil {
		slog.Error("bad url", "url", raw, "err", err)
		os.Exit(1)
	}
	return u
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
