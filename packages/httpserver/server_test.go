package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jjheffernan/BambuFarmer/packages/api"
)

func TestHealthEndpoints(t *testing.T) {
	mux := NewMux("probe")
	for _, path := range []string{"/healthz", "/readyz"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: status %d", path, rec.Code)
		}
		var body api.HealthResponse
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.Status != "ok" || body.Service != "probe" {
			t.Fatalf("%s: %+v", path, body)
		}
	}
}

func TestDefaultAddr(t *testing.T) {
	t.Setenv("HTTP_ADDR", "")
	if DefaultAddr() != ":8080" {
		t.Fatal(DefaultAddr())
	}
	t.Setenv("HTTP_ADDR", ":19090")
	if DefaultAddr() != ":19090" {
		t.Fatal(DefaultAddr())
	}
}
