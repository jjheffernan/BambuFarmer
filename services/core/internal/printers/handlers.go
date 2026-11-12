package printers

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/db"
	"github.com/jjheffernan/BambuFarmer/packages/events"
	"github.com/jjheffernan/BambuFarmer/services/core/internal/httpx"
)

// Register mounts printer projection list and bulk action routes.
func Register(mux *http.ServeMux, pool *pgxpool.Pool, bus *events.Bus, siteID string) {
	mux.HandleFunc("GET /api/v1/printers", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			items, err := db.ListPrinters(r.Context(), pool, siteID)
			if err != nil {
				httpx.WriteError(w, err)
				return
			}
			if items == nil {
				items = []api.PrinterSummary{}
			}
			httpx.WriteJSON(w, items)
		})
	})
	mux.HandleFunc("POST /api/v1/printers/actions", func(w http.ResponseWriter, r *http.Request) {
		var req api.PrinterActionsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, api.BadRequest("invalid JSON body"))
			return
		}
		serials := NormalizeSerials(req)
		res, err := PublishControl(r.Context(), bus, siteID, req.Action, serials)
		if err != nil {
			httpx.WriteError(w, err)
			return
		}
		httpx.WriteJSON(w, res)
	})
}
