package httpserver

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/jjheffernan/BambuFarmer/packages/api"
)

// Options configures a stub HTTP server.
type Options struct {
	Service string
	Addr    string
	Mux     *http.ServeMux // optional; if nil, a new mux is created
}

// NewMux returns a mux with /healthz and /readyz registered.
func NewMux(service string) *http.ServeMux {
	mux := http.NewServeMux()
	RegisterHealth(mux, service)
	return mux
}

// RegisterHealth adds standard health endpoints.
func RegisterHealth(mux *http.ServeMux, service string) {
	write := func(w http.ResponseWriter, status int) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(api.HealthResponse{Status: "ok", Service: service})
	}
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		write(w, http.StatusOK)
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, _ *http.Request) {
		write(w, http.StatusOK)
	})
}

// ListenAndServe starts the HTTP server with sane defaults.
func ListenAndServe(opts Options) error {
	if opts.Addr == "" {
		opts.Addr = ":8080"
	}
	if opts.Service == "" {
		opts.Service = "unknown"
	}
	mux := opts.Mux
	if mux == nil {
		mux = NewMux(opts.Service)
	}
	// Callers that pass Mux should already have called NewMux or RegisterHealth.

	srv := &http.Server{
		Addr:              opts.Addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	slog.Info("listening", "service", opts.Service, "addr", opts.Addr)
	return srv.ListenAndServe()
}

// DefaultAddr from env HTTP_ADDR.
func DefaultAddr() string {
	if v := os.Getenv("HTTP_ADDR"); v != "" {
		return v
	}
	return ":8080"
}