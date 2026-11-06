package farmmanager

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

// HashPasswordChallenge implements the Bambu local challenge-response used by
// Fleet Hub (and expected for FM Client when challenge/salt are present):
//
//	hex(sha256(hex(sha256(hex(sha256(password_plaintext))) + salt) + challenge))
func HashPasswordChallenge(password, salt, challenge string) string {
	h1 := sha256Hex([]byte(password))
	h2 := sha256Hex([]byte(h1 + salt))
	h3 := sha256Hex([]byte(h2 + challenge))
	return h3
}

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// DecodeTicketPayload base64-decodes a ticket string into challenge/salt.
// Ticket may be raw base64 JSON or a JWT-looking blob; we try std and raw encodings.
func DecodeTicketPayload(ticket string) (TicketPayload, error) {
	var out TicketPayload
	raw, err := base64.StdEncoding.DecodeString(ticket)
	if err != nil {
		raw, err = base64.RawStdEncoding.DecodeString(ticket)
		if err != nil {
			// Some captures show URL-safe base64.
			raw, err = base64.URLEncoding.DecodeString(ticket)
			if err != nil {
				raw, err = base64.RawURLEncoding.DecodeString(ticket)
				if err != nil {
					return out, fmt.Errorf("decode ticket: %w", err)
				}
			}
		}
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return out, fmt.Errorf("parse ticket json: %w", err)
	}
	if out.Challenge == "" || out.Salt == "" {
		return out, fmt.Errorf("ticket missing challenge/salt")
	}
	return out, nil
}

// Login performs the local ticket → token exchange as observed for FM Client.
//
// Primary path (GeraldCox Client capture): POST tickets with user_name+password,
// then POST tokens with {ticket}.
//
// If the ticket decodes to challenge/salt and tokens reject the bare ticket,
// callers can re-login after lab capture; HashPasswordChallenge is exported for
// that adjustment. For now we follow the Client capture shape.
func (c *Client) Login(ctx context.Context, user, password string) error {
	if user == "" || password == "" {
		return fmt.Errorf("fm login: user and password required")
	}

	var ticketResp TicketResponse
	if err := c.doJSON(ctx, "POST", "/login/local/tickets", map[string]string{
		"user_name": user,
		"password":  password,
	}, &ticketResp); err != nil {
		return fmt.Errorf("tickets: %w", err)
	}
	if ticketResp.Ticket == "" {
		return fmt.Errorf("tickets: empty ticket (code=%d message=%q)", ticketResp.Code, ticketResp.Message)
	}

	// Best-effort decode for logging / future hashed re-auth; not required for tokens.
	if payload, err := DecodeTicketPayload(ticketResp.Ticket); err == nil {
		_ = payload
	}

	var tokenResp TokenResponse
	if err := c.doJSON(ctx, "POST", "/login/local/tokens", map[string]string{
		"ticket": ticketResp.Ticket,
	}, &tokenResp); err != nil {
		return fmt.Errorf("tokens: %w", err)
	}
	jwt := strings.TrimSpace(tokenResp.JWT())
	if jwt == "" {
		return fmt.Errorf("tokens: empty jwt (code=%d message=%q)", tokenResp.Code, tokenResp.Message)
	}
	c.setToken(jwt)
	return nil
}

// EnsureSession logs in when no JWT is held.
func (c *Client) EnsureSession(ctx context.Context, user, password string) error {
	if c.Token() != "" {
		return nil
	}
	return c.Login(ctx, user, password)
}
