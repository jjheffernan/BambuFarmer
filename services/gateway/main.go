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

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/db"
	"github.com/jjheffernan/BambuFarmer/packages/events"
	"github.com/jjheffernan/BambuFarmer/packages/httpserver"
	"github.com/jjheffernan/BambuFarmer/packages/version"
	"github.com/jjheffernan/BambuFarmer/services/gateway/auth"
	"github.com/jjheffernan/BambuFarmer/services/gateway/ops"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	service := "gateway"
	authMode := auth.NormalizeMode(getenv("AUTH_MODE", "lan"))
	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" || sessionSecret == "change-me-to-a-long-random-string" {
		slog.Warn("SESSION_SECRET is empty or still the example value — set a strong secret before production")
	}
	coreURL := mustURL(getenv("CORE_URL", "http://core:8080"))
	webURL := mustURL(getenv("WEB_URL", "http://web:8080"))
	analyticsURL := mustURL(getenv("ANALYTICS_URL", "http://analytics:8080"))
	integrationsURL := mustURL(getenv("INTEGRATIONS_URL", "http://integrations:8080"))
	edgeURL := mustURL(getenv("EDGE_URL", "http://edge:8080"))
	edgeToken := strings.TrimSpace(os.Getenv("EDGE_TOKEN"))
	commerceURLRaw := os.Getenv("COMMERCE_URL")           // optional plugin; empty = not proxied
	slicerURLRaw := os.Getenv("SLICER_URL")               // optional plugin; empty = not proxied
	studioBridgeURLRaw := os.Getenv("STUDIO_BRIDGE_URL") // optional plugin; empty = not proxied
	redisURL := os.Getenv("REDIS_URL")
	databaseURL := os.Getenv("DATABASE_URL")
	oidcIssuer := os.Getenv("OIDC_ISSUER")
	oidcClientID := os.Getenv("OIDC_CLIENT_ID")
	oidcClientSecret := os.Getenv("OIDC_CLIENT_SECRET")
	oidcRedirect := os.Getenv("OIDC_REDIRECT_URL")
	oidcDefaultRole := getenv("OIDC_DEFAULT_ROLE", "viewer")
	oidcAdminGroups := splitCSV(os.Getenv("OIDC_ADMIN_GROUPS"))
	secureCookies := envBool("SESSION_SECURE", false)

	var oidcClient *auth.OIDCClient
	if authMode == "oidc" {
		c, err := auth.NewOIDCClient(ctx, auth.OIDCConfig{
			Issuer:       oidcIssuer,
			ClientID:     oidcClientID,
			ClientSecret: oidcClientSecret,
			RedirectURL:  oidcRedirect,
			DefaultRole:  oidcDefaultRole,
			AdminGroups:  oidcAdminGroups,
		})
		if err != nil {
			slog.Error("OIDC client init failed", "err", err)
			os.Exit(1)
		}
		oidcClient = c
		slog.Info("OIDC client ready", "issuer", oidcIssuer, "redirect", oidcRedirect)
	}

	var pool *pgxpool.Pool
	if databaseURL != "" {
		var err error
		pool, err = db.Open(ctx, databaseURL)
		if err != nil {
			slog.Error("database open failed", "err", err)
			os.Exit(1)
		}
		defer pool.Close()
		if err := db.Migrate(ctx, pool); err != nil {
			slog.Warn("migrate via gateway failed (core usually owns migrations)", "err", err)
		}
	} else {
		slog.Warn("DATABASE_URL unset; LAN login will not work")
	}

	coreProxy := httputil.NewSingleHostReverseProxy(coreURL)
	webProxy := httputil.NewSingleHostReverseProxy(webURL)
	analyticsProxy := httputil.NewSingleHostReverseProxy(analyticsURL)
	integrationsProxy := httputil.NewSingleHostReverseProxy(integrationsURL)
	edgeProxy := httputil.NewSingleHostReverseProxy(edgeURL)
	if edgeToken != "" {
		origDirector := edgeProxy.Director
		edgeProxy.Director = func(req *http.Request) {
			origDirector(req)
			req.Header.Set("X-Edge-Token", edgeToken)
		}
		slog.Info("EDGE_TOKEN set — injecting X-Edge-Token on edge proxy")
	}
	var commerceProxy *httputil.ReverseProxy
	if commerceURLRaw != "" {
		commerceProxy = httputil.NewSingleHostReverseProxy(mustURL(commerceURLRaw))
	}
	var slicerProxy *httputil.ReverseProxy
	if slicerURLRaw != "" {
		slicerProxy = httputil.NewSingleHostReverseProxy(mustURL(slicerURLRaw))
	}
	var studioBridgeProxy *httputil.ReverseProxy
	if studioBridgeURLRaw != "" {
		studioBridgeProxy = httputil.NewSingleHostReverseProxy(mustURL(studioBridgeURLRaw))
	}

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
	mux.HandleFunc("GET /api/v1/auth/me", func(w http.ResponseWriter, r *http.Request) {
		// Public path: optionally resolve session (Protect does not inject user here).
		if pool != nil && sessionSecret != "" {
			if uid, err := auth.ParseSession(r, sessionSecret); err == nil {
				if u, err := db.GetUserByID(r.Context(), pool, uid); err == nil && u.Active {
					writeJSON(w, map[string]any{
						"auth_mode":     authMode,
						"authenticated": true,
						"user":          u,
						"roles":         u.Roles,
						"warning":       auth.LanWarning(authMode),
					})
					return
				}
			}
		}
		writeJSON(w, map[string]any{
			"auth_mode":     authMode,
			"authenticated": false,
			"warning":       auth.LanWarning(authMode),
		})
	})
	mux.HandleFunc("POST /api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if authMode != "lan" {
			httpserver.WriteError(w, api.BadRequest("local login only in AUTH_MODE=lan"))
			return
		}
		if pool == nil {
			httpserver.WriteError(w, api.NotConfigured("database not configured"))
			return
		}
		if sessionSecret == "" {
			httpserver.WriteError(w, api.NotConfigured("SESSION_SECRET not set"))
			return
		}
		var body api.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			httpserver.WriteError(w, api.BadRequest("invalid JSON body"))
			return
		}
		u, err := db.AuthenticateUser(r.Context(), pool, body.Username, body.Password)
		if err != nil {
			httpserver.WriteError(w, err)
			return
		}
		if err := auth.IssueSession(w, sessionSecret, u.ID, secureCookies); err != nil {
			httpserver.WriteError(w, api.Internal("session issue failed"))
			return
		}
		writeJSON(w, api.LoginResponse{
			AuthMode: authMode,
			User:     u,
			Roles:    u.Roles,
			Warning:  auth.LanWarning(authMode),
		})
	})
	mux.HandleFunc("POST /api/v1/auth/logout", func(w http.ResponseWriter, _ *http.Request) {
		auth.ClearSession(w, secureCookies)
		writeJSON(w, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("POST /api/v1/auth/password", func(w http.ResponseWriter, r *http.Request) {
		if authMode != "lan" {
			httpserver.WriteError(w, api.BadRequest("password change only in AUTH_MODE=lan"))
			return
		}
		if pool == nil {
			httpserver.WriteError(w, api.NotConfigured("database not configured"))
			return
		}
		u, ok := auth.UserFromContext(r.Context())
		if !ok {
			// Protect should have injected; fall back to session parse.
			uid, err := auth.ParseSession(r, sessionSecret)
			if err != nil {
				httpserver.WriteError(w, api.Unauthorized("login required"))
				return
			}
			u, err = db.GetUserByID(r.Context(), pool, uid)
			if err != nil || !u.Active {
				httpserver.WriteError(w, api.Unauthorized("login required"))
				return
			}
		}
		var body api.ChangePasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			httpserver.WriteError(w, api.BadRequest("invalid JSON body"))
			return
		}
		if body.CurrentPassword == "" || body.NewPassword == "" {
			httpserver.WriteError(w, api.BadRequest("current_password and new_password required"))
			return
		}
		if err := db.ChangePassword(r.Context(), pool, u.ID, body.CurrentPassword, body.NewPassword); err != nil {
			httpserver.WriteError(w, err)
			return
		}
		writeJSON(w, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("GET /api/v1/auth/oidc/login", func(w http.ResponseWriter, r *http.Request) {
		if authMode != "oidc" {
			httpserver.WriteError(w, api.BadRequest("OIDC login only when AUTH_MODE=oidc"))
			return
		}
		if oidcClient == nil {
			httpserver.WriteError(w, api.NotConfigured("OIDC client not ready"))
			return
		}
		if sessionSecret == "" {
			httpserver.WriteError(w, api.NotConfigured("SESSION_SECRET not set"))
			return
		}
		if err := oidcClient.BeginLogin(w, r, secureCookies); err != nil {
			httpserver.WriteError(w, api.Internal("OIDC login start failed: "+err.Error()))
		}
	})
	mux.HandleFunc("GET /api/v1/auth/oidc/callback", func(w http.ResponseWriter, r *http.Request) {
		if authMode != "oidc" {
			httpserver.WriteError(w, api.BadRequest("OIDC callback only when AUTH_MODE=oidc"))
			return
		}
		if oidcClient == nil || pool == nil {
			httpserver.WriteError(w, api.NotConfigured("OIDC / database not ready"))
			return
		}
		if sessionSecret == "" {
			httpserver.WriteError(w, api.NotConfigured("SESSION_SECRET not set"))
			return
		}
		claims, err := oidcClient.FinishLogin(r.Context(), r)
		auth.ClearOIDCState(w, secureCookies)
		if err != nil {
			httpserver.WriteError(w, api.Unauthorized("OIDC login failed: "+err.Error()))
			return
		}
		username := claims.Username()
		if username == "" {
			httpserver.WriteError(w, api.Unauthorized("OIDC token missing subject/username"))
			return
		}
		n, _ := db.CountUsers(r.Context(), pool)
		roles := oidcClient.DefaultRoles(claims, n)
		u, err := db.UpsertOIDCUser(r.Context(), pool, username, claims.DisplayName(), roles)
		if err != nil {
			httpserver.WriteError(w, err)
			return
		}
		if err := auth.IssueSession(w, sessionSecret, u.ID, secureCookies); err != nil {
			httpserver.WriteError(w, api.Internal("session issue failed"))
			return
		}
		http.Redirect(w, r, "/", http.StatusFound)
	})

	mux.HandleFunc("GET /api/v1/events/stream", hub.ServeHTTP)

	adminUsers := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		coreProxy.ServeHTTP(w, r)
	})
	mux.Handle("GET /api/v1/admin/users", auth.RequireAdmin(adminUsers))
	mux.Handle("POST /api/v1/admin/users", auth.RequireAdmin(adminUsers))
	mux.Handle("PATCH /api/v1/admin/users/{id}", auth.RequireAdmin(adminUsers))
	mux.Handle("POST /api/v1/admin/users/{id}/password", auth.RequireAdmin(adminUsers))

	mux.Handle("GET /api/v1/admin/system", auth.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]any{
			"version": version.Get(),
			"update":  ops.GetStatus(),
		})
	})))
	mux.Handle("POST /api/v1/admin/system/update", auth.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		by := "admin"
		if u, ok := auth.UserFromContext(r.Context()); ok {
			by = u.Username
		}
		st, err := ops.RequestUpdate(by)
		if err != nil {
			httpserver.WriteError(w, api.Internal("could not write update request: "+err.Error()))
			return
		}
		writeJSON(w, map[string]any{
			"status":  "requested",
			"version": version.Get(),
			"update":  st,
			"hint":    "Host must run: ./scripts/bf update --watch (or apply once with ./scripts/bf update)",
		})
	})))

	mux.Handle("/api/v1/presets/", coreProxy)
	mux.Handle("/api/v1/jobs", coreProxy)
	mux.Handle("/api/v1/jobs/", coreProxy)
	mux.Handle("/api/v1/power", coreProxy)
	mux.Handle("/api/v1/power/", coreProxy)
	mux.Handle("/api/v1/notify", coreProxy)
	mux.Handle("/api/v1/notify/", coreProxy)
	mux.Handle("/api/v1/artifacts", coreProxy)
	mux.Handle("/api/v1/artifacts/", coreProxy)
	mux.Handle("/api/v1/printers", coreProxy)
	mux.Handle("GET /api/v1/printers/{id}/liveview", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		r.URL.Path = "/api/v1/edge/printers/" + id + "/liveview"
		r.URL.RawPath = ""
		edgeProxy.ServeHTTP(w, r)
	}))
	mux.Handle("/api/v1/printers/", coreProxy)
	mux.Handle("/api/v1/admin/", coreProxy)
	mux.Handle("/api/v1/setup", coreProxy)
	mux.Handle("/api/v1/setup/", coreProxy)
	mux.Handle("/api/v1/version", coreProxy)
	mux.Handle("/api/v1/sites", coreProxy)
	mux.Handle("/api/v1/sites/", coreProxy)
	mux.Handle("/api/v1/analytics/", analyticsProxy)
	mux.Handle("/api/v1/integrations/", integrationsProxy)
	mux.Handle("/api/v1/filament/", integrationsProxy)
	mux.Handle("/api/v1/library/", integrationsProxy)
	mux.Handle("/api/v1/edge/", edgeProxy)
	if commerceProxy != nil {
		mux.Handle("/api/v1/commerce", commerceProxy)
		mux.Handle("/api/v1/commerce/", commerceProxy)
	}
	if slicerProxy != nil {
		mux.Handle("/api/v1/slice", slicerProxy)
		mux.Handle("/api/v1/slice/", slicerProxy)
	}
	if studioBridgeProxy != nil {
		mux.Handle("/api/v1/studio-bridge", studioBridgeProxy)
		mux.Handle("/api/v1/studio-bridge/", studioBridgeProxy)
	}

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
	var handler http.Handler = mux
	if !skipSetup {
		handler = setupGate(coreURL.String(), mux)
	}

	installDone := func() bool {
		if pool == nil {
			return true
		}
		st, err := db.GetInstallState(ctx, pool)
		if err != nil {
			return true
		}
		return st.Completed
	}
	handler = auth.Protect(auth.GateOptions{
		Pool:          pool,
		Secret:        sessionSecret,
		AuthMode:      authMode,
		InstallDone:   installDone,
		SecureCookies: secureCookies,
	}, handler)

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
	for _, stream := range []string{events.StreamPrinter, events.StreamJob, events.StreamAlert, events.StreamSlice, events.StreamStudioBridge} {
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
		}
	}
}

func (h *sseHub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		httpserver.WriteError(w, api.Internal("streaming unsupported"))
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
			completed = true
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
		if path == "/setup" || path == "/login" ||
			strings.HasPrefix(path, "/static/") ||
			strings.HasPrefix(path, "/api/v1/setup") ||
			path == "/api/v1/version" ||
			path == "/healthz" || path == "/readyz" ||
			strings.HasPrefix(path, "/api/v1/auth/") {
			next.ServeHTTP(w, r)
			return
		}
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

func splitCSV(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
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
