package printers

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/db"
	"github.com/jjheffernan/BambuFarmer/services/core/internal/httpx"
)

// Register mounts printer projection list routes.
func Register(mux *http.ServeMux, pool *pgxpool.Pool, siteID string) {
	mux.HandleFunc("GET /api/v1/printers", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			items, err := db.ListPrinters(r.Context(), pool, siteID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if items == nil {
				items = []api.PrinterSummary{}
			}
			httpx.WriteJSON(w, items)
		})
	})
}
