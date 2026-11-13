package main

import (
	"strings"
	"testing"
)

func TestLoginPageSupportsOIDC(t *testing.T) {
	if !strings.Contains(loginHTML, `id="oidc-panel"`) || !strings.Contains(loginHTML, "/api/v1/auth/oidc/login") {
		t.Fatal("expected OIDC sign-in control on login page")
	}
	if !strings.Contains(loginHTML, "auth_mode") || !strings.Contains(loginHTML, "/api/v1/auth/me") {
		t.Fatal("expected login page to probe auth mode")
	}
}

func TestShellPasswordChange(t *testing.T) {
	if !strings.Contains(layoutHTML, `id="shell-pw"`) || !strings.Contains(layoutHTML, "/api/v1/auth/password") {
		t.Fatal("expected shell password change for LAN users")
	}
}
