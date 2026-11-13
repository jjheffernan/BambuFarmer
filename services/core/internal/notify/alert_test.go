package notify

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPostAlertEmptyURL(t *testing.T) {
	if err := PostAlert(t.Context(), "  ", nil, Alert{ExternalSN: "x"}); err != nil {
		t.Fatal(err)
	}
}

func TestPostAlertPostsJSON(t *testing.T) {
	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &got)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	err := PostAlert(t.Context(), srv.URL, srv.Client(), Alert{
		ExternalSN: "SN1", Name: "Bay", Reasons: []string{"PAUSE"}, GcodeState: "PAUSE", Online: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got["type"] != "alert.raised" || got["external_sn"] != "SN1" {
		t.Fatalf("%v", got)
	}
}
