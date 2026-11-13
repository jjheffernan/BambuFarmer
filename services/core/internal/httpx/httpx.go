package httpx

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/httpserver"
)

// WriteJSON encodes v as application/json.
func WriteJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

// WriteError writes a typed JSON API error (see packages/api.Error).
func WriteError(w http.ResponseWriter, err error) {
	httpserver.WriteError(w, err)
}

// RequireDB returns 503 when pool is nil; otherwise runs next.
func RequireDB(w http.ResponseWriter, pool *pgxpool.Pool, next func()) {
	if pool == nil {
		WriteError(w, api.NotConfigured("database not configured"))
		return
	}
	next()
}
