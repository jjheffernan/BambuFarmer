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
			items, err := db.ListPrinterSnapshots(r.Context(), pool, siteID)
			if err != nil {
				httpx.WriteError(w, err)
				return
			}
			q := r.URL.Query()
			items = db.FilterPrinterSnapshots(items, db.PrinterListFilter{
				Tag:    q.Get("tag"),
				State:  q.Get("state"),
				Model:  q.Get("model"),
				Query:  q.Get("q"),
				Online: q.Get("online"),
			})
			if items == nil {
				items = []api.PrinterSnapshot{}
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
	mux.HandleFunc("POST /api/v1/printers/{sn}/maintenance", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			var body struct {
				Maintenance *bool `json:"maintenance"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				httpx.WriteError(w, api.BadRequest("invalid JSON body"))
				return
			}
			if body.Maintenance == nil {
				httpx.WriteError(w, api.BadRequest("maintenance required"))
				return
			}
			p, err := db.SetPrinterMaintenance(r.Context(), pool, siteID, r.PathValue("sn"), *body.Maintenance)
			if err != nil {
				httpx.WriteError(w, err)
				return
			}
			httpx.WriteJSON(w, p)
		})
	})
}
