package users

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/db"
	"github.com/jjheffernan/BambuFarmer/services/core/internal/httpx"
)

// Register mounts admin user CRUD (gateway enforces admin role).
func Register(mux *http.ServeMux, pool *pgxpool.Pool) {
	mux.HandleFunc("GET /api/v1/admin/users", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			items, err := db.ListUsers(r.Context(), pool)
			if err != nil {
				httpx.WriteError(w, err)
				return
			}
			if items == nil {
				items = []api.User{}
			}
			httpx.WriteJSON(w, items)
		})
	})
	mux.HandleFunc("POST /api/v1/admin/users", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			var req api.CreateUserRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				httpx.WriteError(w, api.BadRequest("invalid JSON body"))
				return
			}
			u, err := db.CreateUser(r.Context(), pool, req.Username, req.Password, req.DisplayName, req.Roles)
			if err != nil {
				httpx.WriteError(w, err)
				return
			}
			httpx.WriteJSON(w, u)
		})
	})
	mux.HandleFunc("PATCH /api/v1/admin/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			var req api.UpdateUserRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				httpx.WriteError(w, api.BadRequest("invalid JSON body"))
				return
			}
			u, err := db.UpdateUser(r.Context(), pool, r.PathValue("id"), req.Roles, req.Active)
			if err != nil {
				httpx.WriteError(w, err)
				return
			}
			httpx.WriteJSON(w, u)
		})
	})
	mux.HandleFunc("POST /api/v1/admin/users/{id}/password", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			var req api.AdminResetPasswordRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				httpx.WriteError(w, api.BadRequest("invalid JSON body"))
				return
			}
			if req.NewPassword == "" {
				httpx.WriteError(w, api.BadRequest("new_password required"))
				return
			}
			if err := db.SetPassword(r.Context(), pool, r.PathValue("id"), req.NewPassword); err != nil {
				httpx.WriteError(w, err)
				return
			}
			httpx.WriteJSON(w, map[string]string{"status": "ok"})
		})
	})
}
