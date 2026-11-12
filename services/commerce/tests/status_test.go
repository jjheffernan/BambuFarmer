package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jjheffernan/BambuFarmer/packages/httpserver"
)

func TestCommerceStatusRoute(t *testing.T) {
	mux := httpserver.NewMux("commerce")
	mux.HandleFunc("GET /api/v1/commerce/status", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "stub"})
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/commerce/status", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
}
