package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/db"
	"github.com/jjheffernan/BambuFarmer/packages/httpserver"
)

type ctxKey int

const userKey ctxKey = 1

// UserFromContext returns the authenticated user, if any.
func UserFromContext(ctx context.Context) (api.User, bool) {
	u, ok := ctx.Value(userKey).(api.User)
	return u, ok
}

// GateOptions configures auth enforcement.
type GateOptions struct {
	Pool          *pgxpool.Pool
	Secret        string
	AuthMode      string
	InstallDone   func() bool // nil = assume done (enforce when users exist)
	SecureCookies bool
}

// Protect wraps next: when LAN mode, install completed, and users exist, require a valid session
// except for the public allowlist.
func Protect(opts GateOptions, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if opts.AuthMode != "lan" || opts.Pool == nil {
			next.ServeHTTP(w, r)
			return
		}
		if isPublicPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		installDone := true
		if opts.InstallDone != nil {
			installDone = opts.InstallDone()
		}
		if !installDone {
			next.ServeHTTP(w, r)
			return
		}
		n, err := db.CountUsers(r.Context(), opts.Pool)
		if err != nil || n == 0 {
			next.ServeHTTP(w, r)
			return
		}

		uid, err := ParseSession(r, opts.Secret)
		if err != nil {
			writeAuthRequired(w, r)
			return
		}
		u, err := db.GetUserByID(r.Context(), opts.Pool, uid)
		if err != nil || !u.Active {
			ClearSession(w, opts.SecureCookies)
			writeAuthRequired(w, r)
			return
		}
		ctx := context.WithValue(r.Context(), userKey, u)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAdmin rejects non-admin users (JSON 403). Call after Protect.
func RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := UserFromContext(r.Context())
		if !ok || !HasRole(u, db.RoleAdmin) {
			httpserver.WriteError(w, api.Forbidden("admin role required"))
			return
		}
		next(w, r)
	}
}

// HasRole reports whether u has role.
func HasRole(u api.User, role string) bool {
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}
	return false
}

func isPublicPath(path string) bool {
	switch {
	case path == "/healthz", path == "/readyz":
		return true
	case path == "/login", path == "/setup":
		return true
	case strings.HasPrefix(path, "/static/"):
		return true
	case path == "/api/v1/auth/login", path == "/api/v1/auth/logout", path == "/api/v1/auth/me":
		return true
	case strings.HasPrefix(path, "/api/v1/auth/oidc"):
		return true
	case strings.HasPrefix(path, "/api/v1/setup"):
		return true
	case path == "/api/v1/version":
		return true
	default:
		return false
	}
}

func writeAuthRequired(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		httpserver.WriteError(w, api.Unauthorized("login required"))
		return
	}
	http.Redirect(w, r, "/login", http.StatusFound)
}
