package presets

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/db"
	"github.com/jjheffernan/BambuFarmer/services/core/internal/httpx"
)

// Register mounts filament/print preset routes.
func Register(mux *http.ServeMux, pool *pgxpool.Pool, siteID string) {
	mux.HandleFunc("GET /api/v1/presets/filament", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			items, err := db.ListFilamentPresets(r.Context(), pool, siteID)
			if err != nil {
				httpx.WriteError(w, err)
				return
			}
			if items == nil {
				items = []api.FilamentPreset{}
			}
			httpx.WriteJSON(w, items)
		})
	})
	mux.HandleFunc("POST /api/v1/presets/filament", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			var p api.FilamentPreset
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				httpx.WriteError(w, api.BadRequest("invalid JSON body"))
				return
			}
			if p.ID == "" || p.Name == "" {
				httpx.WriteError(w, api.BadRequest("id and name required"))
				return
			}
			out, err := db.UpsertFilamentPreset(r.Context(), pool, siteID, p)
			if err != nil {
				httpx.WriteError(w, err)
				return
			}
			httpx.WriteJSON(w, out)
		})
	})
	mux.HandleFunc("GET /api/v1/presets/filament/{id}", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			out, err := db.GetFilamentPreset(r.Context(), pool, siteID, r.PathValue("id"))
			if err != nil {
				httpx.WriteError(w, err)
				return
			}
			httpx.WriteJSON(w, out)
		})
	})
	mux.HandleFunc("PATCH /api/v1/presets/filament/{id}", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			var p api.FilamentPreset
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				httpx.WriteError(w, api.BadRequest("invalid JSON body"))
				return
			}
			p.ID = r.PathValue("id")
			if p.Name == "" {
				httpx.WriteError(w, api.BadRequest("name required"))
				return
			}
			out, err := db.UpsertFilamentPreset(r.Context(), pool, siteID, p)
			if err != nil {
				httpx.WriteError(w, err)
				return
			}
			httpx.WriteJSON(w, out)
		})
	})
	mux.HandleFunc("DELETE /api/v1/presets/filament/{id}", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			if err := db.DeleteFilamentPreset(r.Context(), pool, siteID, r.PathValue("id")); err != nil {
				httpx.WriteError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})
	})
	mux.HandleFunc("GET /api/v1/presets/print", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			items, err := db.ListPrintPresets(r.Context(), pool, siteID)
			if err != nil {
				httpx.WriteError(w, err)
				return
			}
			if items == nil {
				items = []api.PrintPreset{}
			}
			httpx.WriteJSON(w, items)
		})
	})
	mux.HandleFunc("POST /api/v1/presets/print", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			var p api.PrintPreset
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				httpx.WriteError(w, api.BadRequest("invalid JSON body"))
				return
			}
			if p.ID == "" || p.Name == "" {
				httpx.WriteError(w, api.BadRequest("id and name required"))
				return
			}
			out, err := db.UpsertPrintPreset(r.Context(), pool, siteID, p)
			if err != nil {
				httpx.WriteError(w, err)
				return
			}
			httpx.WriteJSON(w, out)
		})
	})
	mux.HandleFunc("GET /api/v1/presets/print/{id}", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			out, err := db.GetPrintPreset(r.Context(), pool, siteID, r.PathValue("id"))
			if err != nil {
				httpx.WriteError(w, err)
				return
			}
			httpx.WriteJSON(w, out)
		})
	})
	mux.HandleFunc("PATCH /api/v1/presets/print/{id}", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			var p api.PrintPreset
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				httpx.WriteError(w, api.BadRequest("invalid JSON body"))
				return
			}
			p.ID = r.PathValue("id")
			if p.Name == "" {
				httpx.WriteError(w, api.BadRequest("name required"))
				return
			}
			out, err := db.UpsertPrintPreset(r.Context(), pool, siteID, p)
			if err != nil {
				httpx.WriteError(w, err)
				return
			}
			httpx.WriteJSON(w, out)
		})
	})
	mux.HandleFunc("DELETE /api/v1/presets/print/{id}", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			if err := db.DeletePrintPreset(r.Context(), pool, siteID, r.PathValue("id")); err != nil {
				httpx.WriteError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})
	})
	mux.HandleFunc("GET /api/v1/presets/filament/{id}/revisions", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			items, err := db.ListFilamentPresetRevisions(r.Context(), pool, siteID, r.PathValue("id"))
			if err != nil {
				httpx.WriteError(w, err)
				return
			}
			if items == nil {
				items = []db.PresetRevision{}
			}
			httpx.WriteJSON(w, items)
		})
	})
	mux.HandleFunc("POST /api/v1/presets/filament/{id}/restore", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			var body struct {
				Version int `json:"version"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				httpx.WriteError(w, api.BadRequest("invalid JSON body"))
				return
			}
			out, err := db.RestoreFilamentPresetRevision(r.Context(), pool, siteID, r.PathValue("id"), body.Version)
			if err != nil {
				httpx.WriteError(w, err)
				return
			}
			httpx.WriteJSON(w, out)
		})
	})
	mux.HandleFunc("GET /api/v1/presets/print/{id}/revisions", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			items, err := db.ListPrintPresetRevisions(r.Context(), pool, siteID, r.PathValue("id"))
			if err != nil {
				httpx.WriteError(w, err)
				return
			}
			if items == nil {
				items = []db.PresetRevision{}
			}
			httpx.WriteJSON(w, items)
		})
	})
	mux.HandleFunc("POST /api/v1/presets/print/{id}/restore", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			var body struct {
				Version int `json:"version"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				httpx.WriteError(w, api.BadRequest("invalid JSON body"))
				return
			}
			out, err := db.RestorePrintPresetRevision(r.Context(), pool, siteID, r.PathValue("id"), body.Version)
			if err != nil {
				httpx.WriteError(w, err)
				return
			}
			httpx.WriteJSON(w, out)
		})
	})
	mux.HandleFunc("GET /api/v1/presets/export", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			bundle, err := db.ExportPresetBundle(r.Context(), pool, siteID)
			if err != nil {
				httpx.WriteError(w, err)
				return
			}
			httpx.WriteJSON(w, bundle)
		})
	})
	mux.HandleFunc("POST /api/v1/presets/import", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			var bundle api.PresetBundle
			if err := json.NewDecoder(r.Body).Decode(&bundle); err != nil {
				httpx.WriteError(w, api.BadRequest("invalid JSON body"))
				return
			}
			if err := db.ImportPresetBundle(r.Context(), pool, siteID, bundle); err != nil {
				httpx.WriteError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})
	})
}
