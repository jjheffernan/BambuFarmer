package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	SessionCookie = "bf_session"
	sessionTTL    = 12 * time.Hour
)

// SessionClaims are signed into the session cookie.
type SessionClaims struct {
	UserID string `json:"uid"`
	Exp    int64  `json:"exp"`
}

// IssueSession writes an HttpOnly session cookie.
func IssueSession(w http.ResponseWriter, secret, userID string, secure bool) error {
	if secret == "" {
		return fmt.Errorf("SESSION_SECRET is empty")
	}
	claims := SessionClaims{
		UserID: userID,
		Exp:    time.Now().Add(sessionTTL).Unix(),
	}
	raw, err := json.Marshal(claims)
	if err != nil {
		return err
	}
	payload := base64.RawURLEncoding.EncodeToString(raw)
	sig := sign(secret, payload)
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookie,
		Value:    payload + "." + sig,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
		MaxAge:   int(sessionTTL.Seconds()),
	})
	return nil
}

// ClearSession expires the session cookie.
func ClearSession(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
		MaxAge:   -1,
	})
}

// ParseSession reads and validates the session cookie. Returns user id.
func ParseSession(r *http.Request, secret string) (string, error) {
	if secret == "" {
		return "", fmt.Errorf("SESSION_SECRET is empty")
	}
	c, err := r.Cookie(SessionCookie)
	if err != nil || c.Value == "" {
		return "", fmt.Errorf("no session")
	}
	parts := strings.Split(c.Value, ".")
	if len(parts) != 2 {
		return "", fmt.Errorf("malformed session")
	}
	payload, sig := parts[0], parts[1]
	if !hmac.Equal([]byte(sign(secret, payload)), []byte(sig)) {
		return "", fmt.Errorf("bad session signature")
	}
	raw, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return "", err
	}
	var claims SessionClaims
	if err := json.Unmarshal(raw, &claims); err != nil {
		return "", err
	}
	if claims.UserID == "" || time.Now().Unix() > claims.Exp {
		return "", fmt.Errorf("session expired")
	}
	return claims.UserID, nil
}

func sign(secret, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
