package main

import "testing"

func TestNormalizeAuthMode(t *testing.T) {
	if normalizeAuthMode("local_insecure") != "lan" {
		t.Fatal("alias")
	}
	if normalizeAuthMode("lan") != "lan" {
		t.Fatal("lan")
	}
	if normalizeAuthMode("oidc") != "oidc" {
		t.Fatal("oidc")
	}
}

func TestLanWarning(t *testing.T) {
	if lanWarning("lan") == "" {
		t.Fatal("expected warning for lan")
	}
	if lanWarning("oidc") != "" {
		t.Fatal("no warning for oidc")
	}
}
