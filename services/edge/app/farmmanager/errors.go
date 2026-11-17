package farmmanager

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/jjheffernan/BambuFarmer/packages/api"
)

// HTTPError is a non-2xx Farm Manager response with a diagnostic hint.
type HTTPError struct {
	Method string
	Path   string
	Status int
	Body   string
	Hint   string
}

func (e *HTTPError) Error() string {
	if e == nil {
		return ""
	}
	msg := fmt.Sprintf("%s %s: HTTP %d", e.Method, e.Path, e.Status)
	if e.Body != "" {
		msg += ": " + e.Body
	}
	if e.Hint != "" {
		msg += " (" + e.Hint + ")"
	}
	return msg
}

// Details is JSON-friendly for api.Error.Details.
func (e *HTTPError) Details() map[string]any {
	if e == nil {
		return nil
	}
	return map[string]any{
		"upstream":         "farm_manager",
		"method":           e.Method,
		"path":             e.Path,
		"upstream_status":  e.Status,
		"body":             e.Body,
		"hint":             e.Hint,
	}
}

// DecodeError is a JSON parse failure against an FM response (often API drift).
type DecodeError struct {
	Path string
	Err  error
	Hint string
}

func (e *DecodeError) Error() string {
	if e == nil {
		return ""
	}
	msg := fmt.Sprintf("decode %s: %v", e.Path, e.Err)
	if e.Hint != "" {
		msg += " (" + e.Hint + ")"
	}
	return msg
}

func (e *DecodeError) Unwrap() error { return e.Err }

func (e *DecodeError) Details() map[string]any {
	if e == nil {
		return nil
	}
	return map[string]any{
		"upstream": "farm_manager",
		"path":     e.Path,
		"hint":     e.Hint,
	}
}

// HintForStatus returns an operator-facing hint for common FM HTTP statuses.
func HintForStatus(status int, path string) string {
	switch status {
	case http.StatusUnauthorized, http.StatusForbidden:
		return "FM auth failed — check FARM_MANAGER_USER/PASSWORD, client certs, or re-login after FM upgrade"
	case http.StatusNotFound:
		return "FM endpoint missing (" + path + ") — possible Farm Manager API drift after an update; capture /captain + path and adjust farmmanager client"
	case http.StatusMethodNotAllowed:
		return "FM rejected HTTP method for " + path + " — API contract may have changed in this FM release"
	case http.StatusBadRequest:
		return "FM rejected request body/params — field types or required keys may have changed with this FM release"
	case http.StatusConflict:
		return "FM conflict — printer busy, file already present, or task state mismatch"
	case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return "Farm Manager unreachable or restarting — verify FARM_MANAGER_URL and LAN reachability"
	default:
		if status >= 500 {
			return "Farm Manager internal error — check FM Server logs; may be transient after upgrade"
		}
		return "unexpected Farm Manager response — compare against fixtures/fm and SMOKE_FM32 checklist"
	}
}

// NewHTTPError builds an HTTPError with an automatic hint.
func NewHTTPError(method, path string, status int, body string) *HTTPError {
	return &HTTPError{
		Method: method,
		Path:   path,
		Status: status,
		Body:   truncate(body, 256),
		Hint:   HintForStatus(status, path),
	}
}

// MapToAPI converts FM client errors into typed api.Error for HTTP handlers.
func MapToAPI(op string, err error) error {
	if err == nil {
		return nil
	}
	if ae, ok := api.AsError(err); ok {
		return ae
	}
	var he *HTTPError
	if errors.As(err, &he) {
		return &api.Error{
			Status:  http.StatusBadGateway,
			Code:    api.CodeBadGateway,
			Message: op + " failed: " + he.Error(),
			Details: he.Details(),
			Cause:   he,
		}
	}
	var de *DecodeError
	if errors.As(err, &de) {
		return &api.Error{
			Status:  http.StatusBadGateway,
			Code:    api.CodeBadGateway,
			Message: op + " failed: " + de.Error(),
			Details: de.Details(),
			Cause:   de,
		}
	}
	msg := err.Error()
	hint := "Farm Manager / edge adapter error — see last_error on GET /api/v1/edge/status"
	if strings.Contains(msg, "decode") || strings.Contains(msg, "unmarshal") {
		hint = "response JSON shape unexpected — possible FM API drift after an update"
	}
	return &api.Error{
		Status:  http.StatusBadGateway,
		Code:    api.CodeBadGateway,
		Message: op + " failed: " + msg,
		Details: map[string]any{"hint": hint, "upstream": "farm_manager"},
		Cause:   err,
	}
}
