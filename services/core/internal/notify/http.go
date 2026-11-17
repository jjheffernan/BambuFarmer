package notify

import (
	"encoding/json"
	"net/http"
	"strings"
)

// Register mounts GET /api/v1/notify and POST /api/v1/notify/test.
// An empty notify URL accepts the test POST and does not forward.
func Register(mux *http.ServeMux, endpoint string) {
	configured := strings.TrimSpace(endpoint) != ""
	mux.HandleFunc("GET /api/v1/notify", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"configured": configured,
			"events":     []string{TypeJobFinished, TypeAlertRaised},
			"note":       "NOTIFY_URL receives typed JSON (type=job.finished|alert.raised). POST /api/v1/notify/test. Prometheus: scrape service /metrics if exposed; job/fleet signals also via Redis Streams + analytics export.",
		})
	})
	mux.HandleFunc("POST /api/v1/notify/test", func(w http.ResponseWriter, r *http.Request) {
		if !configured {
			writeJSON(w, http.StatusOK, map[string]any{"sent": false, "reason": "NOTIFY_URL unset"})
			return
		}
		body := Finished{Type: TypeJobFinished, JobID: "notify-test", State: "test"}
		if err := PostFinished(r.Context(), endpoint, nil, body); err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"sent": true, "type": body.Type, "job_id": body.JobID, "state": body.State})
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
