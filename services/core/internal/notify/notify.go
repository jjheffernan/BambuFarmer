package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Finished is the job-finished webhook body.
type Finished struct {
	JobID   string  `json:"job_id"`
	State   string  `json:"state"`
	Grams   float64 `json:"grams,omitempty"`
	SpoolID string  `json:"spool_id,omitempty"`
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
