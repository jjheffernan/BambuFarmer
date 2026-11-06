package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"sync"

	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/httpserver"
)

// In-memory preset store for scaffold / local smoke tests.
// Replaced by Postgres in Phase 0+.
type presetStore struct {
	mu       sync.RWMutex
	filament []api.FilamentPreset
	print    []api.PrintPreset
}

func main() {
	service := "core"
	store := &presetStore{}
	mux := httpserver.NewMux(service)

	mux.HandleFunc("GET /api/v1/presets/filament", func(w http.ResponseWriter, _ *http.Request) {
		store.mu.RLock()
		defer store.mu.RUnlock()
		writeJSON(w, store.filament)
	})
	mux.HandleFunc("GET /api/v1/presets/print", func(w http.ResponseWriter, _ *http.Request) {
		store.mu.RLock()
		defer store.mu.RUnlock()
		writeJSON(w, store.print)
	})
	mux.HandleFunc("GET /api/v1/presets/export", func(w http.ResponseWriter, _ *http.Request) {
		store.mu.RLock()
		defer store.mu.RUnlock()
		writeJSON(w, api.PresetBundle{FilamentPresets: store.filament, PrintPresets: store.print})
	})
	mux.HandleFunc("POST /api/v1/presets/import", func(w http.ResponseWriter, r *http.Request) {
		var bundle api.PresetBundle
		if err := json.NewDecoder(r.Body).Decode(&bundle); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		store.mu.Lock()
		store.filament = bundle.FilamentPresets
		store.print = bundle.PrintPresets
		store.mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
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

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
