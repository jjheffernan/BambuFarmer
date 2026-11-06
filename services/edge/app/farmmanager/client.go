package farmmanager

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ClientOptions configure the FM-native HTTP client.
type ClientOptions struct {
	BaseURL     string
	TLSInsecure bool
	CertDir     string // optional: client.crt + client.key (+ ca.crt)
	Timeout     time.Duration
	ServerID    string // x-bbl-sec-sid; filled after Captain
}

// Client is an fm_native Farm Manager REST client (Client-shaped headers).
type Client struct {
	base    string
	http    *http.Client
	serverID string

	mu    sync.RWMutex
	token string
}

// NewClient builds an HTTP client. Prefer http://LAN:8888 for Client-shaped traffic;
// https with TLSInsecure is supported for lab.
func NewClient(opts ClientOptions) (*Client, error) {
	if opts.BaseURL == "" {
		return nil, fmt.Errorf("fm: BaseURL required")
	}
	if opts.Timeout == 0 {
		opts.Timeout = 15 * time.Second
	}

	tlsCfg := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: opts.TLSInsecure, //nolint:gosec // lab opt-in
	}
	if opts.CertDir != "" {
		if err := loadClientCerts(opts.CertDir, tlsCfg); err != nil {
			return nil, err
		}
	}

	tr := &http.Transport{
		TLSClientConfig: tlsCfg,
		Proxy:           http.ProxyFromEnvironment,
	}
	return &Client{
		base:     opts.BaseURL,
		serverID: opts.ServerID,
		http: &http.Client{
			Timeout:   opts.Timeout,
			Transport: tr,
		},
	}, nil
}

func loadClientCerts(dir string, cfg *tls.Config) error {
	certFile := firstExisting(dir, "client.crt", "tls.crt", "cert.pem")
	keyFile := firstExisting(dir, "client.key", "tls.key", "key.pem")
	if certFile != "" && keyFile != "" {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return fmt.Errorf("load client cert: %w", err)
		}
		cfg.Certificates = []tls.Certificate{cert}
	}
	caFile := firstExisting(dir, "ca.crt", "ca.pem")
	if caFile != "" {
		pem, err := os.ReadFile(caFile)
		if err != nil {
			return err
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(pem) {
			return fmt.Errorf("append ca from %s", caFile)
		}
		cfg.RootCAs = pool
	}
	return nil
}

func firstExisting(dir string, names ...string) string {
	for _, n := range names {
		p := filepath.Join(dir, n)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func (c *Client) setToken(tok string) {
	c.mu.Lock()
	c.token = tok
	c.mu.Unlock()
}

// Token returns the current JWT (empty if logged out).
func (c *Client) Token() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.token
}

// SetServerID sets x-bbl-sec-sid (from /captain server_id).
func (c *Client) SetServerID(id string) {
	c.mu.Lock()
	c.serverID = id
	c.mu.Unlock()
}

func (c *Client) serverIDLocked() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.serverID
}

// Connected is true when a JWT is present.
func (c *Client) Connected() bool {
	return c.Token() != ""
}

func (c *Client) doJSON(ctx context.Context, method, path string, body any, out any) error {
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.base+path, rdr)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	c.applyBBLHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%s %s: HTTP %d: %s", method, path, resp.StatusCode, truncate(string(data), 256))
	}
	if out == nil || len(data) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	return nil
}

func (c *Client) applyBBLHeaders(req *http.Request) {
	req.Header.Set("x-bbl-sec-ver", "1")
	req.Header.Set("x-bbl-sec-dev", "1")
	req.Header.Set("X-BBL-Client-Name", "BambuFarmManager")
	req.Header.Set("X-BBL-Client-Type", "farm-web")
	if sid := c.serverIDLocked(); sid != "" {
		req.Header.Set("x-bbl-sec-sid", sid)
	}
	if tok := c.Token(); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// Captain calls GET /captain and stores server_id for subsequent headers.
func (c *Client) Captain(ctx context.Context) (Captain, error) {
	var cap Captain
	if err := c.doJSON(ctx, "GET", "/captain", nil, &cap); err != nil {
		return Captain{}, err
	}
	if cap.ServerID != "" {
		c.SetServerID(cap.ServerID)
	}
	return cap, nil
}
