// Package edgetoken gates edge HTTP with an optional shared secret (EDGE_TOKEN).
package edgetoken

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/httpserver"
)

const Header = "X-Edge-Token"

// Middleware requires EDGE_TOKEN when non-empty. Health endpoints stay open.
func Middleware(token string, next http.Handler) http.Handler {
	token = strings.TrimSpace(token)
	if token == "" {
		return next
	}
	want := []byte(token)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/healthz", "/readyz":
			next.ServeHTTP(w, r)
			return
		}
		got := strings.TrimSpace(r.Header.Get(Header))
		if got == "" {
			if ah := r.Header.Get("Authorization"); strings.HasPrefix(ah, "Bearer ") {
				got = strings.TrimSpace(strings.TrimPrefix(ah, "Bearer "))
			}
		}
		if subtle.ConstantTimeCompare([]byte(got), want) != 1 {
			httpserver.WriteError(w, api.Unauthorized("edge token required"))
			return
		}
		next.ServeHTTP(w, r)
	})
}
