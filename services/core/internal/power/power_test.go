package power

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSwitchEmptyURL(t *testing.T) {
	if err := Switch(t.Context(), "  ", nil, "on"); err != nil {
		t.Fatal(err)
	}
}

func TestSwitchPostsAction(t *testing.T) {
	var got map[string]string
	var method string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	if err := Switch(t.Context(), srv.URL, srv.Client(), "off"); err != nil {
		t.Fatal(err)
	}
	if method != http.MethodPost || got["action"] != "off" {
		t.Fatalf("%s %+v", method, got)
	}
	if err := Switch(t.Context(), srv.URL, srv.Client(), "maybe"); err == nil {
		t.Fatal("expected bad action")
	}
}
