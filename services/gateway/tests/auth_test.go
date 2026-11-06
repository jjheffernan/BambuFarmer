package tests

import (
	"testing"

	"github.com/jjheffernan/BambuFarmer/services/gateway/auth"
)

func TestNormalizeMode(t *testing.T) {
	if auth.NormalizeMode("local_insecure") != "lan" {
		t.Fatal("alias")
	}
	if auth.NormalizeMode("lan") != "lan" {
		t.Fatal("lan")
	}
	if auth.NormalizeMode("oidc") != "oidc" {
		t.Fatal("oidc")
	}
}

func TestLanWarning(t *testing.T) {
	if auth.LanWarning("lan") == "" {
		t.Fatal("expected warning for lan")
	}
	if auth.LanWarning("oidc") != "" {
		t.Fatal("no warning for oidc")
	}
}
