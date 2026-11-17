package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Typed event names for NOTIFY_URL webhooks.
const (
	TypeJobFinished = "job.finished"
	TypeAlertRaised = "alert.raised"
)

// Finished is the job-finished webhook body.
type Finished struct {
	Type       string  `json:"type"`
	JobID      string  `json:"job_id"`
	State      string  `json:"state"`
	ExternalSN string  `json:"external_sn,omitempty"`
	Grams      float64 `json:"grams,omitempty"`
	SpoolID    string  `json:"spool_id,omitempty"`
	SiteID     string  `json:"site_id,omitempty"`
}

// PostFinished POSTs a JSON body when endpoint is set.
// An empty URL does nothing. A nil client uses http.DefaultClient.
func PostFinished(ctx context.Context, endpoint string, client *http.Client, body Finished) error {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return nil
	}
	if client == nil {
		client = http.DefaultClient
	}
	if body.Type == "" {
		body.Type = TypeJobFinished
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
