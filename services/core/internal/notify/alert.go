package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Alert is the alert.raised webhook body.
type Alert struct {
	Type       string   `json:"type"`
	ExternalSN string   `json:"external_sn"`
	Name       string   `json:"name,omitempty"`
	Model      string   `json:"model,omitempty"`
	Online     bool     `json:"online"`
	GcodeState string   `json:"gcode_state,omitempty"`
	Reasons    []string `json:"reasons,omitempty"`
	HMS        []string `json:"hms,omitempty"`
	SiteID     string   `json:"site_id,omitempty"`
}

// PostAlert POSTs a JSON body when endpoint is set.
// An empty URL does nothing. A nil client uses http.DefaultClient.
func PostAlert(ctx context.Context, endpoint string, client *http.Client, body Alert) error {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return nil
	}
	if client == nil {
		client = http.DefaultClient
	}
	if body.Type == "" {
		body.Type = "alert.raised"
	}
	raw, err := json.Marshal(body)
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
		return fmt.Errorf("notify webhook: %s", res.Status)
	}
	return nil
}
