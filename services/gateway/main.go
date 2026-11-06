package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/jjheffernan/BambuFarmer/packages/events"
	"github.com/jjheffernan/BambuFarmer/packages/httpserver"
	"github.com/jjheffernan/BambuFarmer/services/gateway/internal/auth"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	service := "gateway"
	authMode := auth.NormalizeMode(getenv("AUTH_MODE", "lan"))
	coreURL := mustURL(getenv("CORE_URL", "http://core:8080"))
	webURL := mustURL(getenv("WEB_URL", "http://web:8080"))
	analyticsURL := mustURL(getenv("ANALYTICS_URL", "http://analytics:8080"))
	integrationsURL := mustURL(getenv("INTEGRATIONS_URL", "http://integrations:8080"))
	edgeURL := mustURL(getenv("EDGE_URL", "http://edge:8080"))
	redisURL := os.Getenv("REDIS_URL")
	oidcIssuer := os.Getenv("OIDC_ISSUER")
	oidcClientID := os.Getenv("OIDC_CLIENT_ID")
	oidcRedirect := os.Getenv("OIDC_REDIRECT_URL")

	coreProxy := httputil.NewSingleHostReverseProxy(coreURL)
	webProxy := httputil.NewSingleHostReverseProxy(webURL)
	analyticsProxy := httputil.NewSingleHostReverseProxy(analyticsURL)
	integrationsProxy := httputil.NewSingleHostReverseProxy(integrationsURL)
	edgeProxy := httputil.NewSingleHostReverseProxy(edgeURL)

	hub := newSSEHub()
	var bus *events.Bus
	if redisURL != "" {
		var err error
		bus, err = events.Dial(ctx, redisURL)
		if err != nil {
			slog.Warn("redis unavailable; SSE will have no events", "err", err)
		} else {
			defer bus.Close()
			go fanout(ctx, bus, hub)
		}
	}

	mux := httpserver.NewMux(service)
	mux.HandleFunc("GET /api/v1/auth/me", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]any{
			"auth_mode": authMode,
			"user":      "stub",
			"roles":     []string{"admin"},
			"warning":   auth.LanWarning(authMode),
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
		writeJSON(w, map[string]any{
			"token":     "lan-stub-token",
			"auth_mode": "lan",
			"user":      body.Username,
		})
	})

	// OIDC stubs (Phase 0): real exchange lands later.
	mux.HandleFunc("GET /api/v1/auth/oidc/login", func(w http.ResponseWriter, r *http.Request) {
		if authMode != "oidc" {
			http.Error(w, "OIDC login only when AUTH_MODE=oidc", http.StatusBadRequest)
			return
		}
		if oidcIssuer == "" || oidcClientID == "" || oidcRedirect == "" {
			http.Error(w, "OIDC_ISSUER / OIDC_CLIENT_ID / OIDC_REDIRECT_URL required", http.StatusServiceUnavailable)
			return
		}
		// Stub: redirect shape only — full OIDC code flow not wired.
		target := fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&response_type=code&scope=openid%%20profile",
			strings.TrimRight(oidcIssuer, "/")+"/protocol/openid-connect/auth",
			url.QueryEscape(oidcClientID),
			url.QueryEscape(oidcRedirect),
		)
		writeJSON(w, map[string]any{
			"status":      "stub",
			"auth_url":    target,
			"note":        "OIDC authorize URL stub — complete code exchange not implemented yet",
			"auth_mode":   "oidc",
		})
	})
	mux.HandleFunc("GET /api/v1/auth/oidc/callback", func(w http.ResponseWriter, r *http.Request) {
		if authMode != "oidc" {
			http.Error(w, "OIDC callback only when AUTH_MODE=oidc", http.StatusBadRequest)
			return
		}
		code := r.URL.Query().Get("code")
		writeJSON(w, map[string]any{
			"status":  "stub",
			"code":    code != "",
			"note":    "token exchange not implemented — wire IdP client next",
			"auth_mode": "oidc",
		})
	})

	mux.HandleFunc("GET /api/v1/events/stream", hub.ServeHTTP)

	mux.Handle("/api/v1/presets/", coreProxy)
	mux.Handle("/api/v1/jobs", coreProxy)
	mux.Handle("/api/v1/jobs/", coreProxy)
	mux.Handle("/api/v1/printers", coreProxy)
	mux.Handle("/api/v1/printers/", coreProxy)
	mux.Handle("/api/v1/admin/", coreProxy)
	mux.Handle("/api/v1/setup", coreProxy)
	mux.Handle("/api/v1/setup/", coreProxy)
	mux.Handle("/api/v1/version", coreProxy)
	mux.Handle("/api/v1/analytics/", analyticsProxy)
	mux.Handle("/api/v1/integrations/", integrationsProxy)
	mux.Handle("/api/v1/filament/", integrationsProxy)
	mux.Handle("/api/v1/library/", integrationsProxy)
	mux.Handle("/api/v1/edge/", edgeProxy)

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

	skipSetup := envBool("SKIP_SETUP_REDIRECT", false)
	handler := http.Handler(mux)
	if !skipSetup {
		handler = setupGate(coreURL.String(), mux)
	}

	srv := &http.Server{
		Addr:              httpserver.DefaultAddr(),
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		slog.Info("listening", "service", service, "addr", srv.Addr, "auth_mode", authMode)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server stopped", "err", err)
			os.Exit(1)
		}
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}

func fanout(ctx context.Context, bus *events.Bus, hub *sseHub) {
	for _, stream := range []string{events.StreamPrinter, events.StreamJob, events.StreamAlert} {
		s := stream
		go func() {
			consumer := "gateway-" + getenv("HOSTNAME", "1") + "-" + s
			_ = bus.ConsumeLoop(ctx, s, events.GroupGatewayFanout, consumer, "$", func(_ context.Context, id string, env events.Envelope) error {
				hub.Broadcast(env)
				return nil
			})
		}()
	}
	<-ctx.Done()
}

type sseHub struct {
	mu   sync.Mutex
	subs map[chan events.Envelope]struct{}
}

func newSSEHub() *sseHub {
	return &sseHub{subs: make(map[chan events.Envelope]struct{})}
}

func (h *sseHub) Broadcast(env events.Envelope) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.subs {
		select {
		case ch <- env:
		default:
			// drop if slow client
		}
	}
}

func (h *sseHub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ch := make(chan events.Envelope, 16)
	h.mu.Lock()
	h.subs[ch] = struct{}{}
	h.mu.Unlock()
	defer func() {
		h.mu.Lock()
		delete(h.subs, ch)
		h.mu.Unlock()
		close(ch)
	}()

	fmt.Fprintf(w, "event: ready\ndata: {\"ok\":true}\n\n")
	flusher.Flush()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			fmt.Fprintf(w, ": ping\n\n")
			flusher.Flush()
		case env := <-ch:
			b, _ := json.Marshal(env)
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", env.Type, b)
			flusher.Flush()
		}
	}
}

// setupGate redirects browser UI to /setup until install_state.completed.
func setupGate(coreBase string, next http.Handler) http.Handler {
	client := &http.Client{Timeout: 2 * time.Second}
	var (
		mu        sync.Mutex
		cached    bool
		cachedAt  time.Time
		completed bool
	)
	refresh := func() bool {
		mu.Lock()
		defer mu.Unlock()
		if time.Since(cachedAt) < 3*time.Second && cached {
			return completed
		}
		resp, err := client.Get(strings.TrimRight(coreBase, "/") + "/api/v1/setup")
		if err != nil {
			cached = true
			cachedAt = time.Now()
			completed = true // fail open so ops UI still works if core is down mid-boot
			return completed
		}
		defer resp.Body.Close()
		var body struct {
			Install struct {
				Completed bool `json:"completed"`
			} `json:"install"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&body)
		completed = body.Install.Completed
		cached = true
		cachedAt = time.Now()
		return completed
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/setup" ||
			strings.HasPrefix(path, "/api/v1/setup") ||
			path == "/api/v1/version" ||
			path == "/healthz" || path == "/readyz" ||
			strings.HasPrefix(path, "/api/v1/auth/") {
			next.ServeHTTP(w, r)
			return
		}
		// Only gate HTML navigations; leave API usable for the wizard itself.
		if strings.HasPrefix(path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}
		if !refresh() {
			http.Redirect(w, r, "/setup", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func envBool(k string, d bool) bool {
	v := os.Getenv(k)
	if v == "" {
		return d
	}
	switch strings.ToLower(v) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return d
	}
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
