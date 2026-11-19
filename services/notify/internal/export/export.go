package export

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/jjheffernan/BambuFarmer/packages/events"
	"github.com/jjheffernan/BambuFarmer/services/notify/internal/store"
)

// Delivery is one webhook POST attempt result.
type Delivery struct {
	ChannelID string `json:"channel_id"`
	Name      string `json:"name"`
	OK        bool   `json:"ok"`
	Error     string `json:"error,omitempty"`
}

// Client POSTs envelopes to webhook channels.
type Client struct {
	HTTP *http.Client
	Log  *slog.Logger
}

func (c *Client) http() *http.Client {
	if c != nil && c.HTTP != nil {
		return c.HTTP
	}
	return &http.Client{Timeout: 10 * time.Second}
}

func (c *Client) log() *slog.Logger {
	if c != nil && c.Log != nil {
		return c.Log
	}
	return slog.Default()
}

// Body is the JSON posted to webhook destinations.
type Body struct {
	Type    string         `json:"type"`
	TS      time.Time      `json:"ts"`
	SiteID  string         `json:"site_id,omitempty"`
	Source  string         `json:"source,omitempty"`
	Payload map[string]any `json:"payload"`
}

// PostChannel sends one body to one channel.
func (c *Client) PostChannel(ctx context.Context, ch store.Channel, body Body) error {
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ch.URL, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if ah := strings.TrimSpace(ch.AuthHeader); ah != "" {
		req.Header.Set("Authorization", ah)
	}
	res, err := c.http().Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return fmt.Errorf("webhook %s", res.Status)
	}
	return nil
}

// FanOut posts env to all matching channels (best-effort; never returns stream error).
func (c *Client) FanOut(ctx context.Context, st *store.Store, env events.Envelope) []Delivery {
	sn, _ := env.Payload["external_sn"].(string)
	site := env.SiteID
	channels, err := st.ChannelsForEvent(ctx, site, env.Type, sn)
	if err != nil {
		c.log().Warn("notify channel lookup", "err", err, "type", env.Type)
		return nil
	}
	body := Body{
		Type:    env.Type,
		TS:      env.TS,
		SiteID:  env.SiteID,
		Source:  env.Source,
		Payload: env.Payload,
	}
	if body.TS.IsZero() {
		body.TS = time.Now().UTC()
	}
	if body.Payload == nil {
		body.Payload = map[string]any{}
	}
	var out []Delivery
	for _, ch := range channels {
		d := Delivery{ChannelID: ch.ID, Name: ch.Name}
		if err := c.PostChannel(ctx, ch, body); err != nil {
			d.Error = err.Error()
			c.log().Warn("notify webhook", "channel", ch.ID, "err", err, "type", env.Type)
		} else {
			d.OK = true
		}
		out = append(out, d)
	}
	return out
}

// TestBody builds a sample job.finished payload.
func TestBody(siteID string) Body {
	return Body{
		Type:   events.TypeJobFinished,
		TS:     time.Now().UTC(),
		SiteID: siteID,
		Source: "notify",
		Payload: map[string]any{
			"job_id": "notify-test",
			"state":  "test",
		},
	}
}
