package db

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jjheffernan/BambuFarmer/packages/api"
)

// Sentinel / constructors for data-layer failures callers can errors.Is / api.AsError.

// ErrNotFound is returned when a row is missing (wraps pgx.ErrNoRows).
var ErrNotFound = api.NotFound("not found")

// MapQueryError converts driver errors into typed API errors.
func MapQueryError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return api.Wrap(ErrNotFound.Status, api.CodeNotFound, "not found", err)
	}
	return err
}

// InvalidJobTransition returns a typed error for illegal job state changes.
func InvalidJobTransition(from, to string) error {
	e := api.InvalidTransition(fmt.Sprintf("invalid job transition %s → %s", from, to))
	e.Details = map[string]string{"from": from, "to": to}
	return e
}

// Validation returns a 400 bad_request for caller-supplied input problems.
func Validation(msg string) error {
	return api.BadRequest(msg)
}
