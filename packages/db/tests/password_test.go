package tests

import (
	"testing"

	"github.com/jjheffernan/BambuFarmer/packages/db"
)

func TestHashVerifyPassword(t *testing.T) {
	h, err := db.HashPassword("correct-horse-battery")
	if err != nil {
		t.Fatal(err)
	}
	ok, err := db.VerifyPassword(h, "correct-horse-battery")
	if err != nil || !ok {
		t.Fatalf("verify ok=%v err=%v", ok, err)
	}
	ok, err = db.VerifyPassword(h, "wrong")
	if err != nil || ok {
		t.Fatalf("expected reject ok=%v err=%v", ok, err)
	}
}

func TestHashPasswordTooShort(t *testing.T) {
	if _, err := db.HashPassword("short"); err == nil {
		t.Fatal("expected validation error")
	}
}
