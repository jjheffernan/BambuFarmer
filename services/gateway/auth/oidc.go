package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

const (
	oidcStateCookie = "bf_oidc_state"
	oidcStateTTL    = 10 * time.Minute
)

// OIDCConfig is loaded from env when AUTH_MODE=oidc.
type OIDCConfig struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	DefaultRole  string   // viewer|operator|admin — used on first insert
	AdminGroups  []string // claim groups that grant admin on create (and when sync roles)
}

// OIDCClient wraps discovery + oauth2 for authorization-code flow.
type OIDCClient struct {
	cfg      OIDCConfig
	provider *oidc.Provider
	verifier *oidc.IDTokenVerifier
	oauth    oauth2.Config
}

// NewOIDCClient discovers the issuer. Call once at boot.
func NewOIDCClient(ctx context.Context, cfg OIDCConfig) (*OIDCClient, error) {
	cfg.Issuer = strings.TrimRight(strings.TrimSpace(cfg.Issuer), "/")
	cfg.ClientID = strings.TrimSpace(cfg.ClientID)
	cfg.ClientSecret = strings.TrimSpace(cfg.ClientSecret)
	cfg.RedirectURL = strings.TrimSpace(cfg.RedirectURL)
	if cfg.Issuer == "" || cfg.ClientID == "" || cfg.RedirectURL == "" {
		return nil, fmt.Errorf("OIDC_ISSUER, OIDC_CLIENT_ID, and OIDC_REDIRECT_URL required")
	}
	if cfg.DefaultRole == "" {
		cfg.DefaultRole = "viewer"
	}
	provider, err := oidc.NewProvider(ctx, cfg.Issuer)
	if err != nil {
		return nil, fmt.Errorf("oidc discovery: %w", err)
	}
	oauth := oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}
	return &OIDCClient{
		cfg:      cfg,
		provider: provider,
		verifier: provider.Verifier(&oidc.Config{ClientID: cfg.ClientID}),
		oauth:    oauth,
	}, nil
}

// Ready reports whether the client can start a login.
func (c *OIDCClient) Ready() bool { return c != nil }

// BeginLogin sets a state cookie and redirects to the IdP authorize URL.
func (c *OIDCClient) BeginLogin(w http.ResponseWriter, r *http.Request, secure bool) error {
	state, err := randomState()
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     oidcStateCookie,
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
		MaxAge:   int(oidcStateTTL.Seconds()),
	})
	http.Redirect(w, r, c.oauth.AuthCodeURL(state), http.StatusFound)
	return nil
}

// Claims are verified ID-token fields used for local user upsert.
type Claims struct {
	Subject           string
	Email             string
	PreferredUsername string
	Name              string
	Groups            []string
}

// Username picks a stable local username.
func (c Claims) Username() string {
	for _, s := range []string{c.PreferredUsername, c.Email, c.Subject} {
		s = strings.TrimSpace(s)
		if s != "" {
			return s
		}
	}
	return ""
}

// DisplayName prefers name, then username.
func (c Claims) DisplayName() string {
	if n := strings.TrimSpace(c.Name); n != "" {
		return n
	}
	return c.Username()
}

// FinishLogin validates state, exchanges the code, and returns claims.
func (c *OIDCClient) FinishLogin(ctx context.Context, r *http.Request) (Claims, error) {
	var empty Claims
	stateCookie, err := r.Cookie(oidcStateCookie)
	if err != nil || stateCookie.Value == "" {
		return empty, fmt.Errorf("missing OIDC state cookie")
	}
	qState := r.URL.Query().Get("state")
	if qState == "" || qState != stateCookie.Value {
		return empty, fmt.Errorf("OIDC state mismatch")
	}
	if errMsg := r.URL.Query().Get("error"); errMsg != "" {
		desc := r.URL.Query().Get("error_description")
		if desc != "" {
			return empty, fmt.Errorf("IdP error: %s (%s)", errMsg, desc)
		}
		return empty, fmt.Errorf("IdP error: %s", errMsg)
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		return empty, fmt.Errorf("missing authorization code")
	}
	tok, err := c.oauth.Exchange(ctx, code)
	if err != nil {
		return empty, fmt.Errorf("token exchange: %w", err)
	}
	rawID, ok := tok.Extra("id_token").(string)
	if !ok || rawID == "" {
		return empty, fmt.Errorf("no id_token in token response")
	}
	idTok, err := c.verifier.Verify(ctx, rawID)
	if err != nil {
		return empty, fmt.Errorf("id_token verify: %w", err)
	}
	var payload map[string]any
	if err := idTok.Claims(&payload); err != nil {
		return empty, err
	}
	return parseClaims(payload), nil
}

// ClearState expires the state cookie after callback.
func ClearOIDCState(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     oidcStateCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
		MaxAge:   -1,
	})
}

// DefaultRoles returns roles for a new OIDC user.
func (c *OIDCClient) DefaultRoles(claims Claims, userCount int) []string {
	if userCount == 0 {
		return []string{"admin"}
	}
	for _, g := range claims.Groups {
		g = strings.TrimSpace(g)
		for _, want := range c.cfg.AdminGroups {
			if strings.EqualFold(g, want) {
				return []string{"admin"}
			}
		}
	}
	role := strings.ToLower(strings.TrimSpace(c.cfg.DefaultRole))
	switch role {
	case "admin", "operator", "viewer":
		return []string{role}
	default:
		return []string{"viewer"}
	}
}

func parseClaims(payload map[string]any) Claims {
	c := Claims{
		Subject:           strClaim(payload, "sub"),
		Email:             strClaim(payload, "email"),
		PreferredUsername: strClaim(payload, "preferred_username"),
		Name:              strClaim(payload, "name"),
	}
	if c.Name == "" {
		c.Name = strings.TrimSpace(strClaim(payload, "given_name") + " " + strClaim(payload, "family_name"))
	}
	c.Groups = groupClaim(payload)
	return c
}

func strClaim(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	default:
		b, _ := json.Marshal(t)
		return strings.Trim(string(b), `"`)
	}
}

func groupClaim(m map[string]any) []string {
	for _, key := range []string{"groups", "roles", "realm_access"} {
		v, ok := m[key]
		if !ok || v == nil {
			continue
		}
		switch t := v.(type) {
		case []any:
			var out []string
			for _, x := range t {
				if s, ok := x.(string); ok && s != "" {
					out = append(out, s)
				}
			}
			if len(out) > 0 {
				return out
			}
		case []string:
			return t
		case map[string]any:
			// Keycloak realm_access.roles
			if roles, ok := t["roles"].([]any); ok {
				var out []string
				for _, x := range roles {
					if s, ok := x.(string); ok && s != "" {
						out = append(out, s)
					}
				}
				if len(out) > 0 {
					return out
				}
			}
		}
	}
	return nil
}

func randomState() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
