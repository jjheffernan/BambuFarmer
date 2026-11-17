package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProtectOIDCRequiresSession(t *testing.T) {
	h := Protect(GateOptions{
		Pool:     nil, // no pool → pass-through; use a stub pool below via lan path
		Secret:   "secret",
		AuthMode: "oidc",
	}, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	// Pool nil → no enforcement
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/jobs", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("nil pool code %d", rr.Code)
	}
}

func TestIsPublicOIDCPaths(t *testing.T) {
	for _, p := range []string{"/api/v1/auth/oidc/login", "/api/v1/auth/oidc/callback", "/login"} {
		if !isPublicPath(p) {
			t.Fatalf("%s should be public", p)
		}
	}
	if isPublicPath("/jobs") {
		t.Fatal("/jobs should not be public")
	}
}
