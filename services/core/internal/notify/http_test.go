package notify

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetNotifyStatus(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, "")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/notify", nil))
	if rr.Code != http.StatusOK {
		t.Fatal(rr.Code)
	}
	var body map[string]any
	_ = json.NewDecoder(rr.Body).Decode(&body)
	if body["configured"] != false {
		t.Fatalf("%+v", body)
	}
	mux2 := http.NewServeMux()
	Register(mux2, "http://hook.local/notify")
	rr2 := httptest.NewRecorder()
	mux2.ServeHTTP(rr2, httptest.NewRequest(http.MethodGet, "/api/v1/notify", nil))
	_ = json.NewDecoder(rr2.Body).Decode(&body)
	if body["configured"] != true {
		t.Fatalf("%+v", body)
	}
}

func TestNotifyTestPostsWhenConfigured(t *testing.T) {
	var got Finished
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	mux := http.NewServeMux()
	Register(mux, srv.URL)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/notify/test", nil))
	if rr.Code != http.StatusOK {
		t.Fatal(rr.Code, rr.Body.String())
	}
	var body map[string]any
	_ = json.NewDecoder(rr.Body).Decode(&body)
	if body["sent"] != true || got.JobID != "notify-test" || got.State != "test" {
		t.Fatalf("resp=%+v got=%+v", body, got)
	}
}

func TestNotifyTestDryRunWhenUnset(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, "  ")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/notify/test", nil))
	if rr.Code != http.StatusOK {
		t.Fatal(rr.Code)
	}
	var body map[string]any
	_ = json.NewDecoder(rr.Body).Decode(&body)
	if body["sent"] != false {
		t.Fatalf("%+v", body)
	}
}
