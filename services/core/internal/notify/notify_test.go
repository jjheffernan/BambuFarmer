package notify

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPostFinishedEmptyURL(t *testing.T) {
	if err := PostFinished(t.Context(), "  ", nil, Finished{JobID: "j"}); err != nil {
		t.Fatal(err)
	}
}

func TestPostFinishedPostsJSON(t *testing.T) {
	var got Finished
	var method, ctype string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method, ctype = r.Method, r.Header.Get("Content-Type")
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	err := PostFinished(t.Context(), srv.URL, srv.Client(), Finished{JobID: "j1", State: "succeeded", Grams: 3})
	if err != nil {
		t.Fatal(err)
	}
	if method != http.MethodPost || ctype != "application/json" || got.JobID != "j1" || got.State != "succeeded" || got.Grams != 3 {
		t.Fatalf("%s %s %+v", method, ctype, got)
	}
}
