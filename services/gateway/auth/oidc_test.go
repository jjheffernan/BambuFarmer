package auth

import "testing"

func TestParseClaimsUsername(t *testing.T) {
	c := parseClaims(map[string]any{
		"sub":                "abc",
		"preferred_username": "ops",
		"email":              "ops@example.com",
		"name":               "Ops User",
		"groups":             []any{"farm-admins"},
	})
	if c.Username() != "ops" {
		t.Fatalf("username %q", c.Username())
	}
	if c.DisplayName() != "Ops User" {
		t.Fatalf("display %q", c.DisplayName())
	}
	if len(c.Groups) != 1 || c.Groups[0] != "farm-admins" {
		t.Fatalf("groups %#v", c.Groups)
	}
}

func TestDefaultRoles(t *testing.T) {
	c := &OIDCClient{cfg: OIDCConfig{DefaultRole: "operator", AdminGroups: []string{"farm-admins"}}}
	if roles := c.DefaultRoles(Claims{}, 0); len(roles) != 1 || roles[0] != "admin" {
		t.Fatalf("first user %#v", roles)
	}
	if roles := c.DefaultRoles(Claims{Groups: []string{"farm-admins"}}, 3); roles[0] != "admin" {
		t.Fatalf("admin group %#v", roles)
	}
	if roles := c.DefaultRoles(Claims{}, 3); roles[0] != "operator" {
		t.Fatalf("default %#v", roles)
	}
}

func TestRealmAccessRoles(t *testing.T) {
	c := parseClaims(map[string]any{
		"sub": "x",
		"realm_access": map[string]any{
			"roles": []any{"farm-admins", "offline_access"},
		},
	})
	if len(c.Groups) < 1 || c.Groups[0] != "farm-admins" {
		t.Fatalf("%#v", c.Groups)
	}
}
