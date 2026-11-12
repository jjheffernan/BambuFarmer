package shop

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestQueueJobSkipsWhenCoreUnset(t *testing.T) {
	res, code, err := QueueJob(t.Context(), "  ", nil, Order{Name: "plate"})
	if err != nil || code != http.StatusOK || res.Queued {
		t.Fatalf("%d %+v %v", code, res, err)
	}
}

func TestQueueJobPostsCore(t *testing.T) {
	var got map[string]string
	var path string
	core := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"id": "job-9"})
	}))
	defer core.Close()

	mux := http.NewServeMux()
	Register(mux, core.URL)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/commerce/webhook", strings.NewReader(`{"external_id":"ord-1","artifact_id":"art"}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d %s", rec.Code, rec.Body.String())
	}
	if path != "/api/v1/jobs" || got["name"] != "ord-1" || got["artifact_id"] != "art" {
		t.Fatalf("path %s body %#v", path, got)
	}
	var out Result
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil || !out.Queued || out.JobID != "job-9" {
		t.Fatalf("%+v %v", out, err)
	}
}
