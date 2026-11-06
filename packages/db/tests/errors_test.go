package tests

import (
	"errors"
	"net/http"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/db"
)

func TestMapQueryErrorNoRows(t *testing.T) {
	err := db.MapQueryError(pgx.ErrNoRows)
	ae, ok := api.AsError(err)
	if !ok || ae.HTTPStatus() != http.StatusNotFound || ae.Code != api.CodeNotFound {
		t.Fatalf("%v ok=%v", err, ok)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("expected wrap of pgx.ErrNoRows")
	}
}

func TestInvalidJobTransition(t *testing.T) {
	err := db.InvalidJobTransition("queued", "succeeded")
	ae, ok := api.AsError(err)
	if !ok || ae.Code != api.CodeInvalidTransition {
		t.Fatalf("%v", err)
	}
}

func TestValidation(t *testing.T) {
	err := db.Validation("missing id")
	ae, ok := api.AsError(err)
	if !ok || ae.HTTPStatus() != http.StatusBadRequest {
		t.Fatalf("%v", err)
	}
}
