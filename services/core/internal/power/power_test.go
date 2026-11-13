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

func TestGetPowerStatus(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, "")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/power", nil))
	if rr.Code != http.StatusOK {
		t.Fatal(rr.Code)
	}
	var body map[string]any
	_ = json.NewDecoder(rr.Body).Decode(&body)
	if body["configured"] != false {
		t.Fatalf("%+v", body)
	}
	mux2 := http.NewServeMux()
	Register(mux2, "http://plug.local/switch")
	rr2 := httptest.NewRecorder()
	mux2.ServeHTTP(rr2, httptest.NewRequest(http.MethodGet, "/api/v1/power", nil))
	_ = json.NewDecoder(rr2.Body).Decode(&body)
	if body["configured"] != true {
		t.Fatalf("%+v", body)
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
