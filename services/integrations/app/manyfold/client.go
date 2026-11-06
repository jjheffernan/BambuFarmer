// Package manyfold talks to Manyfold REST API v0
// (https://github.com/manyfold3d/manyfold — Accept: application/vnd.manyfold.v0+json).
package manyfold

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/jjheffernan/BambuFarmer/packages/api"
)

const acceptV0 = "application/vnd.manyfold.v0+json"

// Client is a Manyfold HTTP adapter.
type Client struct {
	base         string
	token        string
	clientID     string
	clientSecret string
	http         *http.Client

	mu        sync.Mutex
	oauthTok  string
	oauthExp  time.Time
}

// New returns a client. Prefer MANYFOLD_TOKEN (personal access token); otherwise
// client_credentials via client id/secret when the instance supports it.
func New(baseURL, token, clientID, clientSecret string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	return &Client{
		base:         strings.TrimRight(baseURL, "/"),
		token:        token,
		clientID:     clientID,
		clientSecret: clientSecret,
		http:         &http.Client{Timeout: timeout},
	}
}

type collectionDTO struct {
	Type        string `json:"@type"`
	TotalItems  int    `json:"totalItems"`
	Member      []memberDTO `json:"member"`
	// Some serializers nest under hydra:member
	HydraMember []memberDTO `json:"hydra:member"`
}

type memberDTO struct {
	ID   string `json:"@id"`
	Name string `json:"name"`
	Type string `json:"@type"`
}

// ListModels calls GET /models?page=&order= (JSON-LD hydra:Collection).
func (c *Client) ListModels(ctx context.Context, page int, order string) ([]api.ModelSummary, error) {
	if c == nil || c.base == "" {
		return nil, fmt.Errorf("manyfold: not configured")
	}
	if page <= 0 {
		page = 1
	}
	if order == "" {
		order = "name"
	}
	u, err := url.Parse(c.base + "/models")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("page", fmt.Sprintf("%d", page))
	q.Set("order", order)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", acceptV0)
	if err := c.authorize(ctx, req); err != nil {
		return nil, err
	}
	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(res.Body, 4<<20))
	if res.StatusCode >= 300 {
		return nil, fmt.Errorf("manyfold: list models: %s: %s", res.Status, truncate(string(body), 200))
	}
	var col collectionDTO
	if err := json.Unmarshal(body, &col); err != nil {
		return nil, err
	}
	members := col.Member
	if len(members) == 0 {
		members = col.HydraMember
	}
	out := make([]api.ModelSummary, 0, len(members))
	for _, m := range members {
		id := m.ID
		if id == "" {
			continue
		}
		out = append(out, api.ModelSummary{
			ID:   id,
			Name: m.Name,
			URL:  absolutize(c.base, id),
		})
	}
	return out, nil
}

func (c *Client) authorize(ctx context.Context, req *http.Request) error {
	tok := c.token
	if tok == "" {
		var err error
		tok, err = c.oauthToken(ctx)
		if err != nil {
			return err
		}
	}
	if tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	return nil
}

func (c *Client) oauthToken(ctx context.Context) (string, error) {
	if c.clientID == "" || c.clientSecret == "" {
		return "", nil // public / unauthenticated read may still work
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.oauthTok != "" && time.Now().Before(c.oauthExp) {
		return c.oauthTok, nil
	}
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", c.clientID)
	form.Set("client_secret", c.clientSecret)
	form.Set("scope", "public read")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/oauth/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode >= 300 {
		return "", fmt.Errorf("manyfold: oauth token: %s: %s", res.Status, truncate(string(body), 200))
	}
	var tok struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tok); err != nil {
		return "", err
	}
	if tok.AccessToken == "" {
		return "", fmt.Errorf("manyfold: oauth token empty")
	}
	exp := 30 * time.Minute
	if tok.ExpiresIn > 0 {
		exp = time.Duration(tok.ExpiresIn) * time.Second
	}
	c.oauthTok = tok.AccessToken
	c.oauthExp = time.Now().Add(exp - time.Minute)
	return c.oauthTok, nil
}

func absolutize(base, id string) string {
	if strings.HasPrefix(id, "http://") || strings.HasPrefix(id, "https://") {
		return id
	}
	if strings.HasPrefix(id, "/") {
		return strings.TrimRight(base, "/") + id
	}
	return strings.TrimRight(base, "/") + "/" + id
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
