package export

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jjheffernan/BambuFarmer/services/notify/internal/store"
)

func TestPostChannel(t *testing.T) {
	var gotAuth string
	var gotBody Body
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := &Client{HTTP: srv.Client()}
	ch := store.Channel{ID: "nch_1", URL: srv.URL, AuthHeader: "Bearer secret"}
	body := Body{Type: "job.finished", TS: time.Now().UTC(), Payload: map[string]any{"job_id": "j1"}}
	if err := c.PostChannel(context.Background(), ch, body); err != nil {
		t.Fatal(err)
	}
	if gotAuth != "Bearer secret" {
		t.Fatalf("auth=%q", gotAuth)
	}
	if gotBody.Type != "job.finished" || gotBody.Payload["job_id"] != "j1" {
		t.Fatalf("%+v", gotBody)
	}
}

func TestPostChannelErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()
	c := &Client{HTTP: srv.Client()}
	err := c.PostChannel(context.Background(), store.Channel{URL: srv.URL}, Body{Type: "x", Payload: map[string]any{}})
	if err == nil {
		t.Fatal("expected error")
	}
}
