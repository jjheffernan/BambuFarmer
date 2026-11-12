package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/jjheffernan/BambuFarmer/packages/httpserver"
	"github.com/jjheffernan/BambuFarmer/services/integrations/app/manyfold"
	"github.com/jjheffernan/BambuFarmer/services/integrations/app/spoolman"
	"github.com/jjheffernan/BambuFarmer/services/integrations/app/status"
)

// Deps are optional upstream clients (nil = disabled).
type Deps struct {
	Spoolman      *spoolman.Client
	Manyfold      *manyfold.Client
	SpoolURL      string
	ManyURL       string
	LowStockGrams float64
}

// Register mounts integration HTTP routes.
func Register(mux *http.ServeMux, d Deps) {
	mux.HandleFunc("GET /api/v1/integrations/status", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]any{
			"spoolman": status.URL(d.SpoolURL),
			"manyfold": status.URL(d.ManyURL),
			"mode":     "optional_live_url",
			"refs": map[string]string{
				"spoolman": "https://github.com/Donkie/Spoolman",
				"manyfold": "https://github.com/manyfold3d/manyfold",
			},
		})
	})

	mux.HandleFunc("GET /api/v1/filament/spools", func(w http.ResponseWriter, r *http.Request) {
		if d.Spoolman == nil {
			writeJSON(w, map[string]any{"configured": false, "spools": []any{}, "low_stock": []any{}})
			return
		}
		spools, err := d.Spoolman.ListSpools(r.Context(), 1000)
		if err != nil {
			httpserver.WriteError(w, err)
			return
		}
		writeJSON(w, map[string]any{
			"configured":  true,
			"threshold_g": d.LowStockGrams,
			"spools":      spools,
			"low_stock":   spoolman.LowStock(spools, d.LowStockGrams),
		})
	})

	mux.HandleFunc("GET /api/v1/library/models", func(w http.ResponseWriter, r *http.Request) {
		if d.Manyfold == nil {
			writeJSON(w, map[string]any{"configured": false, "models": []any{}})
			return
		}
		models, err := d.Manyfold.ListModels(r.Context(), 1, "name")
		if err != nil {
			httpserver.WriteError(w, err)
			return
		}
		writeJSON(w, map[string]any{"configured": true, "models": models})
	})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
