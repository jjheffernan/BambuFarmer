package cdp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestReadyProbesVersion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/json/version" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"Browser":"Chrome/test","webSocketDebuggerUrl":"ws://x"}`))
	}))
	defer srv.Close()

	p := New(srv.URL)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := p.Ready(ctx); err != nil {
		t.Fatal(err)
	}
	if err := p.Connect(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestReadyRequiresURL(t *testing.T) {
	p := New("")
	if err := p.Ready(context.Background()); err == nil {
		t.Fatal("expected error")
	}
}
