package tests

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/jjheffernan/BambuFarmer/services/edge/internal/fm"
)

func TestHashPasswordChallenge(t *testing.T) {
	got := fm.HashPasswordChallenge("Secret1!", "salt", "challenge")
	if len(got) != 64 {
		t.Fatalf("len=%d want 64, got %s", len(got), got)
	}
	again := fm.HashPasswordChallenge("Secret1!", "salt", "challenge")
	if got != again {
		t.Fatal("not deterministic")
	}
	other := fm.HashPasswordChallenge("Secret1!", "salt", "other")
	if got == other {
		t.Fatal("challenge should change hash")
	}
}

func TestDecodeTicketPayload(t *testing.T) {
	payload := fm.TicketPayload{Challenge: "ch", Salt: "sa"}
	raw, _ := json.Marshal(payload)
	ticket := base64.StdEncoding.EncodeToString(raw)
	got, err := fm.DecodeTicketPayload(ticket)
	if err != nil {
		t.Fatal(err)
	}
	if got.Challenge != "ch" || got.Salt != "sa" {
		t.Fatalf("%+v", got)
	}
}
