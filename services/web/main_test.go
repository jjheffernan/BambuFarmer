package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jjheffernan/BambuFarmer/packages/httpserver"
)

func TestSetupRouteRegistered(t *testing.T) {
	mux := httpserver.NewMux("web")
	registerSetup(mux, "http://127.0.0.1:18080")
	req := httptest.NewRequest(http.MethodGet, "/setup", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Install") && !strings.Contains(body, "wizard") && !strings.Contains(body, "BambuFarmer") {
		t.Fatalf("unexpected body prefix: %.120s", body)
	}
}
