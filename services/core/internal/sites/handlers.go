package sites

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/db"
	"github.com/jjheffernan/BambuFarmer/services/core/internal/httpx"
)

// Register mounts site list/get routes.
func Register(mux *http.ServeMux, pool *pgxpool.Pool) {
	mux.HandleFunc("GET /api/v1/sites", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			items, err := db.ListSites(r.Context(), pool)
			if err != nil {
				httpx.WriteError(w, err)
				return
			}
			if items == nil {
				items = []api.Site{}
			}
			httpx.WriteJSON(w, items)
		})
	})
	mux.HandleFunc("GET /api/v1/sites/{id}", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			site, err := db.GetSite(r.Context(), pool, r.PathValue("id"))
			if err != nil {
				httpx.WriteError(w, err)
				return
			}
			httpx.WriteJSON(w, site)
		})
	})
}
