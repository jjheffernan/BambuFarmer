package auth

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/httpserver"
)

// CSRF rejects cross-site state-changing requests that carry a session cookie.
// Same-origin fetch() always sends Origin; curl/scripts without a session pass.
func CSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
			next.ServeHTTP(w, r)
			return
		}
		if !hasSessionCookie(r) {
			next.ServeHTTP(w, r)
			return
		}
		if !sameOrigin(r) {
			httpserver.WriteError(w, api.Forbidden("CSRF: cross-origin request blocked"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func hasSessionCookie(r *http.Request) bool {
	c, err := r.Cookie(SessionCookie)
	return err == nil && c != nil && c.Value != ""
}

func sameOrigin(r *http.Request) bool {
	host := r.Host
	if o := strings.TrimSpace(r.Header.Get("Origin")); o != "" {
		u, err := url.Parse(o)
		return err == nil && hostEqual(u.Host, host)
	}
	if ref := strings.TrimSpace(r.Header.Get("Referer")); ref != "" {
		u, err := url.Parse(ref)
		return err == nil && hostEqual(u.Host, host)
	}
	// Session cookie present but no Origin/Referer — atypical for browsers; deny.
	return false
}

func hostEqual(a, b string) bool {
	return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b))
}
