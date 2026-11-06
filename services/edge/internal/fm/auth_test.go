package fm

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

func TestHashPasswordChallenge(t *testing.T) {
	// Deterministic fixture: known inputs → stable hex length 64.
	got := HashPasswordChallenge("Secret1!", "salt", "challenge")
	if len(got) != 64 {
		t.Fatalf("len=%d want 64, got %s", len(got), got)
	}
	again := HashPasswordChallenge("Secret1!", "salt", "challenge")
	if got != again {
		t.Fatal("not deterministic")
	}
	other := HashPasswordChallenge("Secret1!", "salt", "other")
	if got == other {
		t.Fatal("challenge should change hash")
	}
}

func TestDecodeTicketPayload(t *testing.T) {
	payload := TicketPayload{Challenge: "ch", Salt: "sa"}
	raw, _ := json.Marshal(payload)
	ticket := base64.StdEncoding.EncodeToString(raw)
	got, err := DecodeTicketPayload(ticket)
	if err != nil {
		t.Fatal(err)
	}
	if got.Challenge != "ch" || got.Salt != "sa" {
		t.Fatalf("%+v", got)
	}
}
