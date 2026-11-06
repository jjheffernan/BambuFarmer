package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jjheffernan/BambuFarmer/packages/httpserver"
	"github.com/jjheffernan/BambuFarmer/services/web/static"
)

func TestStaticCSSEmbedded(t *testing.T) {
	mux := httpserver.NewMux("web")
	mux.Handle("GET /static/", static.Handler())
	req := httptest.NewRequest(http.MethodGet, "/static/css/app.css", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, ".bf-shell") && !strings.Contains(body, "bf-shell") {
		t.Fatalf("app.css missing bf-shell component: %.80s", body)
	}
	if !strings.Contains(body, "--color-bf-accent") && !strings.Contains(body, "bf-accent") {
		t.Fatalf("app.css missing accent token")
	}
}
