package rates

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/db"
	"github.com/jjheffernan/BambuFarmer/services/core/internal/httpx"
)

// Register mounts rate-card admin routes.
func Register(mux *http.ServeMux, pool *pgxpool.Pool, siteID string) {
	mux.HandleFunc("GET /api/v1/admin/rate-cards", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			rc, err := db.GetRateCard(r.Context(), pool, siteID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			httpx.WriteJSON(w, rc)
		})
	})
	mux.HandleFunc("PUT /api/v1/admin/rate-cards", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			var rc api.RateCard
			if err := json.NewDecoder(r.Body).Decode(&rc); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			out, err := db.UpsertRateCard(r.Context(), pool, siteID, rc)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			httpx.WriteJSON(w, out)
		})
	})
}
