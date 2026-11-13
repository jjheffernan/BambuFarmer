// Command slicer is the optional slice plugin (compose profile: slicer).
// ORCA_BIN empty accepts a job and does not invoke a slicer.
// When ORCA_BIN is set, POST /api/v1/slice/jobs runs the operator-supplied CLI.
// When REDIS_URL is set, slice.queued / slice.succeeded|failed are published.
package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jjheffernan/BambuFarmer/packages/events"
	"github.com/jjheffernan/BambuFarmer/packages/httpserver"
	"github.com/jjheffernan/BambuFarmer/services/slicer/internal/slice"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	service := "slicer"
	bin := os.Getenv("ORCA_BIN")
	workDir := getenv("SLICE_WORK_DIR", getenv("ARTIFACTS_DIR", ""))
	siteID := getenv("SITE_ID", "default")
	redisURL := os.Getenv("REDIS_URL")

	var bus *events.Bus
	if redisURL != "" {
		b, err := events.Dial(ctx, redisURL)
		if err != nil {
			slog.Warn("redis dial failed; slice events disabled", "err", err)
		} else {
			bus = b
			defer bus.Close()
		}
	}

	mux := httpserver.NewMux(service)
	mux.HandleFunc("POST /api/v1/slice/jobs", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Name         string `json:"name"`
			InputPath    string `json:"input_path"`
			OutputPath   string `json:"output_path"`
			PrinterJSON  string `json:"printer_json"`
			ProcessJSON  string `json:"process_json"`
			FilamentJSON string `json:"filament_json"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
			return
		}
		req := slice.Request{
			Name:         body.Name,
			InputPath:    body.InputPath,
			OutputPath:   body.OutputPath,
			PrinterJSON:  body.PrinterJSON,
			ProcessJSON:  body.ProcessJSON,
			FilamentJSON: body.FilamentJSON,
			WorkDir:      workDir,
			Timeout:      30 * time.Minute,
		}
		res, err := slice.Job(r.Context(), bin, req)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		slice.Emit(r.Context(), bus, siteID, req, res)
		writeJSON(w, http.StatusOK, res)
	})
	mux.HandleFunc("GET /api/v1/slice/status", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"service":         service,
			"orca_configured": strings.TrimSpace(bin) != "",
			"events":          bus != nil,
			"work_dir":        workDir,
			"note":            "POST /api/v1/slice/jobs with input_path invokes ORCA_BIN when set. Slice events publish when REDIS_URL is set. The Orca binary is not bundled.",
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

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
