package httpx

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

// WriteJSON encodes v as application/json.
func WriteJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

// RequireDB returns 503 when pool is nil; otherwise runs next.
func RequireDB(w http.ResponseWriter, pool *pgxpool.Pool, next func()) {
	if pool == nil {
		http.Error(w, "database not configured", http.StatusServiceUnavailable)
		return
	}
	next()
}
