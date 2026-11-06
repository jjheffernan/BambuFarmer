package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jjheffernan/BambuFarmer/services/core/internal/httpx"
)

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	httpx.WriteJSON(rec, map[string]string{"ok": "yes"})
	if rec.Code != http.StatusOK {
		t.Fatal(rec.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["ok"] != "yes" {
		t.Fatal(body)
	}
}

func TestRequireDBNil(t *testing.T) {
	rec := httptest.NewRecorder()
	called := false
	httpx.RequireDB(rec, nil, func() { called = true })
	if called || rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("called=%v code=%d", called, rec.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["code"] != "not_configured" {
		t.Fatalf("body=%v", body)
	}
}
