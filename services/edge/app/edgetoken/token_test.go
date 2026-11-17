package edgetoken

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddlewarePassthroughWhenEmpty(t *testing.T) {
	h := Middleware("", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/edge/status", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("%d", rr.Code)
	}
}

func TestMiddlewareRequiresToken(t *testing.T) {
	h := Middleware("secret", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/edge/status", nil))
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("want 401 got %d", rr.Code)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/edge/status", nil)
	req.Header.Set(Header, "secret")
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("header auth %d", rr.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/edge/status", nil)
	req.Header.Set("Authorization", "Bearer secret")
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("bearer auth %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("healthz %d", rr.Code)
	}
}
