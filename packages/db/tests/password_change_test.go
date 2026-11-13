package tests

import (
	"testing"

	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/db"
)

func TestChangePasswordValidation(t *testing.T) {
	// Hash/verify path used by ChangePassword when DB is unavailable —
	// ensure new password rules still apply via HashPassword.
	if _, err := db.HashPassword("short"); err == nil {
		t.Fatal("expected short password rejected")
	}
	h, err := db.HashPassword("old-password-ok")
	if err != nil {
		t.Fatal(err)
	}
	ok, err := db.VerifyPassword(h, "old-password-ok")
	if err != nil || !ok {
		t.Fatal("verify")
	}
	// Conflict constructor used when passwords match is Validation — smoke.
	err = api.BadRequest("x")
	if err == nil {
		t.Fatal("expected error")
	}
}
