package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jjheffernan/BambuFarmer/services/gateway/auth"
)

func TestSessionRoundTrip(t *testing.T) {
	secret := "test-session-secret-not-for-prod"
	rec := httptest.NewRecorder()
	if err := auth.IssueSession(rec, secret, "usr_abc", false); err != nil {
		t.Fatal(err)
	}
	res := rec.Result()
	cookies := res.Cookies()
	if len(cookies) == 0 {
		t.Fatal("no cookie")
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(cookies[0])
	uid, err := auth.ParseSession(req, secret)
	if err != nil || uid != "usr_abc" {
		t.Fatalf("uid=%q err=%v", uid, err)
	}
}

func TestSessionBadSecret(t *testing.T) {
	rec := httptest.NewRecorder()
	_ = auth.IssueSession(rec, "secret-a", "usr_1", false)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(rec.Result().Cookies()[0])
	if _, err := auth.ParseSession(req, "secret-b"); err == nil {
		t.Fatal("expected signature failure")
	}
}
