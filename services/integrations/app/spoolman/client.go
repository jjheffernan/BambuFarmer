// Package spoolman talks to Spoolman REST API v1
// (https://github.com/Donkie/Spoolman — GET /api/v1/spool).
package spoolman

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jjheffernan/BambuFarmer/packages/api"
)

// Client is a Spoolman HTTP adapter.
type Client struct {
	base   string
	token  string
	http   *http.Client
}

// New returns a client for baseURL (e.g. http://spoolman:7912). Token is optional
// (Spoolman itself has no auth; use when a reverse proxy expects X-API-Key / Bearer).
func New(baseURL, token string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	return &Client{
		base:  strings.TrimRight(baseURL, "/"),
		token: token,
		http:  &http.Client{Timeout: timeout},
	}
}

type filamentDTO struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Material string `json:"material"`
	ColorHex string `json:"color_hex"`
	Vendor   *struct {
		Name string `json:"name"`
	} `json:"vendor"`
}

type spoolDTO struct {
	ID              int         `json:"id"`
	RemainingWeight *float64    `json:"remaining_weight"`
	UsedWeight      float64     `json:"used_weight"`
	Location        string      `json:"location"`
	Archived        bool        `json:"archived"`
	Filament        filamentDTO `json:"filament"`
}

// ListSpools calls GET /api/v1/spool?limit=… (Spoolman paginates; high limit for MVP).
func (c *Client) ListSpools(ctx context.Context, limit int) ([]api.SpoolSummary, error) {
	if c == nil || c.base == "" {
		return nil, fmt.Errorf("spoolman: not configured")
	}
	if limit <= 0 {
		limit = 1000
	}
	u, err := url.Parse(c.base + "/api/v1/spool")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("limit", strconv.Itoa(limit))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	if c.token != "" {
		req.Header.Set("X-API-Key", c.token)
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return nil, fmt.Errorf("spoolman: list spools: %s", res.Status)
	}
	var raw []spoolDTO
	if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
		return nil, err
	}
	out := make([]api.SpoolSummary, 0, len(raw))
	for _, s := range raw {
		name := s.Filament.Name
		vendor := ""
		if s.Filament.Vendor != nil {
			vendor = s.Filament.Vendor.Name
		}
		rem := 0.0
		if s.RemainingWeight != nil {
			rem = *s.RemainingWeight
		}
		out = append(out, api.SpoolSummary{
			ID:         strconv.Itoa(s.ID),
			SpoolmanID: s.ID,
			Name:       name,
			Material:   s.Filament.Material,
			ColorHex:   s.Filament.ColorHex,
			Vendor:     vendor,
			RemainingG: rem,
			UsedG:      s.UsedWeight,
			Location:   s.Location,
			Archived:   s.Archived,
		})
	}
	return out, nil
}
