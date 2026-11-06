package setup

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/db"
	"github.com/jjheffernan/BambuFarmer/packages/version"
	"github.com/jjheffernan/BambuFarmer/services/core/internal/httpx"
)

// Register mounts install-wizard setup routes.
func Register(mux *http.ServeMux, pool *pgxpool.Pool, siteID string) {
	mux.HandleFunc("GET /api/v1/setup", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			st, err := db.GetInstallState(r.Context(), pool)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			rc, _ := db.GetRateCard(r.Context(), pool, siteID)
			httpx.WriteJSON(w, map[string]any{
				"install":   st,
				"rate_card": rc,
				"version":   version.Get(),
				"site_id":   siteID,
			})
		})
	})
	mux.HandleFunc("POST /api/v1/setup", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			var req api.SetupRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			var (
				st  api.InstallState
				err error
			)
			if req.Complete {
				st, err = db.CompleteInstall(r.Context(), pool, siteID, version.Version, req)
			} else {
				st, err = db.SaveInstallDraft(r.Context(), pool, siteID, req)
			}
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			httpx.WriteJSON(w, st)
		})
	})
	mux.HandleFunc("POST /api/v1/setup/reset", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			if err := db.ResetInstall(r.Context(), pool); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			st, err := db.GetInstallState(r.Context(), pool)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			httpx.WriteJSON(w, st)
		})
	})
}
