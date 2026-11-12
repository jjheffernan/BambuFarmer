package power

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Switch POSTs {"action":"on"|"off"} when endpoint is set.
// An empty URL does nothing.
func Switch(ctx context.Context, endpoint string, client *http.Client, action string) error {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return nil
	}
	action = strings.ToLower(strings.TrimSpace(action))
	if action != "on" && action != "off" {
		return fmt.Errorf("power: action must be on or off")
	}
	if client == nil {
		client = http.DefaultClient
	}
	raw, err := json.Marshal(map[string]string{"action": action})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return fmt.Errorf("power: %s", res.Status)
	}
	return nil
}

// Register mounts POST /api/v1/power. An empty power URL accepts the call and does not POST.
func Register(mux *http.ServeMux, endpoint string) {
	mux.HandleFunc("POST /api/v1/power", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Action string `json:"action"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
			return
		}
		action := strings.ToLower(strings.TrimSpace(body.Action))
		if strings.TrimSpace(endpoint) == "" {
			writeJSON(w, http.StatusOK, map[string]any{"switched": false, "reason": "POWER_URL unset"})
			return
		}
		if action != "on" && action != "off" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "action must be on or off"})
			return
		}
		if err := Switch(r.Context(), endpoint, nil, action); err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"switched": true, "action": action})
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
