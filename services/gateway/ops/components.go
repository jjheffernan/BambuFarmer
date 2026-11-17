package ops

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Component is one updatable compose service / plugin.
type Component struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Kind     string `json:"kind"` // core | plugin
	Profile  string `json:"profile,omitempty"`
	URLEnv   string `json:"-"`
	Health   string `json:"health"` // ok | down | unknown | offline
	Detail   string `json:"detail,omitempty"`
	Version  string `json:"version,omitempty"`
	Required bool   `json:"required"`
}

// Catalog is the fixed list of stack parts Admin can refresh independently.
func Catalog() []Component {
	return []Component{
		{ID: "gateway", Label: "Gateway", Kind: "core", URLEnv: "", Required: true},
		{ID: "core", Label: "Core", Kind: "core", URLEnv: "CORE_URL", Required: true},
		{ID: "edge", Label: "Edge", Kind: "core", URLEnv: "EDGE_URL", Required: true},
		{ID: "web", Label: "Web", Kind: "core", URLEnv: "WEB_URL", Required: true},
		{ID: "analytics", Label: "Analytics", Kind: "core", URLEnv: "ANALYTICS_URL", Required: true},
		{ID: "integrations", Label: "Integrations", Kind: "core", URLEnv: "INTEGRATIONS_URL", Required: true},
		{ID: "commerce", Label: "Commerce", Kind: "plugin", Profile: "commerce", URLEnv: "COMMERCE_URL"},
		{ID: "slicer", Label: "Slicer", Kind: "plugin", Profile: "slicer", URLEnv: "SLICER_URL"},
		{ID: "studio-bridge", Label: "Studio bridge", Kind: "plugin", Profile: "studio-bridge", URLEnv: "STUDIO_BRIDGE_URL"},
	}
}

// ValidComponentIDs returns the allowlist.
func ValidComponentIDs() map[string]bool {
	m := map[string]bool{}
	for _, c := range Catalog() {
		m[c.ID] = true
	}
	return m
}

// NormalizeComponents filters to known IDs; empty input means all required+configured plugins.
func NormalizeComponents(ids []string) []string {
	allow := ValidComponentIDs()
	if len(ids) == 0 {
		var all []string
		for _, c := range Catalog() {
			if c.Required || strings.TrimSpace(os.Getenv(c.URLEnv)) != "" || c.ID == "gateway" {
				all = append(all, c.ID)
			}
		}
		return all
	}
	seen := map[string]bool{}
	var out []string
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if !allow[id] || seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	return out
}

// ProbeComponents hits each service healthz (best-effort).
func ProbeComponents(ctx context.Context) []Component {
	client := &http.Client{Timeout: 2 * time.Second}
	out := Catalog()
	for i := range out {
		c := &out[i]
		base := strings.TrimSpace(os.Getenv(c.URLEnv))
		if c.ID == "gateway" {
			// Gateway is this process — always ok when answering.
			c.Health = "ok"
			c.Detail = "this process"
			continue
		}
		if base == "" {
			if c.Kind == "plugin" {
				c.Health = "offline"
				c.Detail = c.URLEnv + " unset (profile off)"
			} else {
				c.Health = "unknown"
				c.Detail = c.URLEnv + " unset"
			}
			continue
		}
		url := strings.TrimRight(base, "/") + "/healthz"
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			c.Health = "down"
			c.Detail = err.Error()
			continue
		}
		if c.ID == "edge" {
			if tok := strings.TrimSpace(os.Getenv("EDGE_TOKEN")); tok != "" {
				req.Header.Set("X-Edge-Token", tok)
			}
		}
		res, err := client.Do(req)
		if err != nil {
			c.Health = "down"
			c.Detail = err.Error()
			continue
		}
		body, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
		_ = res.Body.Close()
		if res.StatusCode < 200 || res.StatusCode >= 300 {
			c.Health = "down"
			c.Detail = "HTTP " + res.Status
			continue
		}
		c.Health = "ok"
		var payload map[string]any
		if json.Unmarshal(body, &payload) == nil {
			if v, ok := payload["service"].(string); ok && v != "" {
				c.Detail = v
			}
			if v, ok := payload["version"].(string); ok {
				c.Version = v
			}
		}
	}
	return out
}
