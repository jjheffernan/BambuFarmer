package static

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStaticServesPWAAssets(t *testing.T) {
	h := Handler()
	for _, path := range []string{
		"/static/css/app.css",
		"/static/manifest.webmanifest",
		"/static/sw.js",
		"/static/icons/icon.svg",
		"/static/icons/icon-192.png",
		"/static/icons/apple-touch-icon.png",
	} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status=%d", path, rec.Code)
		}
		body, _ := io.ReadAll(rec.Body)
		if len(body) == 0 {
			t.Fatalf("%s empty body", path)
		}
	}
	req := httptest.NewRequest(http.MethodGet, "/static/sw.js", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if got := rec.Header().Get("Service-Worker-Allowed"); got != "/" {
		t.Fatalf("Service-Worker-Allowed=%q want /", got)
	}
}
