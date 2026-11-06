package api

import (
	"errors"
	"fmt"
	"net/http"
)

// Stable machine-readable error codes for API clients.
const (
	CodeBadRequest   = "bad_request"
	CodeNotFound     = "not_found"
	CodeConflict     = "conflict"
	CodeUnauthorized = "unauthorized"
	CodeForbidden    = "forbidden"
	CodeUnavailable  = "unavailable"
	CodeBadGateway   = "bad_gateway"
	CodeNotConfigured = "not_configured"
	CodeInternal     = "internal"
	CodeInvalidTransition = "invalid_transition"
)

// ErrorBody is the JSON envelope returned for failed API calls.
type ErrorBody struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Details any    `json:"details,omitempty"`
}

// Error is a typed application error with an HTTP status and stable code.
type Error struct {
	Code    string
	Message string
	Status  int
	Details any
	Cause   error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *Error) Unwrap() error { return e.Cause }

// HTTPStatus returns the status to write, defaulting to 500.
func (e *Error) HTTPStatus() int {
	if e == nil || e.Status == 0 {
		return http.StatusInternalServerError
	}
	return e.Status
}

// Body builds the JSON response payload.
func (e *Error) Body() ErrorBody {
	msg := e.Message
	if msg == "" {
		msg = e.Error()
	}
	code := e.Code
	if code == "" {
		code = CodeInternal
	}
	return ErrorBody{Error: msg, Code: code, Details: e.Details}
}

// New builds an Error with status, code, and message.
func New(status int, code, message string) *Error {
	return &Error{Status: status, Code: code, Message: message}
}

// Wrap annotates cause with status/code/message.
func Wrap(status int, code, message string, cause error) *Error {
	return &Error{Status: status, Code: code, Message: message, Cause: cause}
}

// BadRequest is a 400 with code bad_request.
func BadRequest(message string) *Error {
	return New(http.StatusBadRequest, CodeBadRequest, message)
}

// NotFound is a 404 with code not_found.
func NotFound(message string) *Error {
	return New(http.StatusNotFound, CodeNotFound, message)
}

// Conflict is a 409 with code conflict.
func Conflict(message string) *Error {
	return New(http.StatusConflict, CodeConflict, message)
}

// Unauthorized is a 401.
func Unauthorized(message string) *Error {
	return New(http.StatusUnauthorized, CodeUnauthorized, message)
}

// Forbidden is a 403.
func Forbidden(message string) *Error {
	return New(http.StatusForbidden, CodeForbidden, message)
}

// Unavailable is a 503.
func Unavailable(message string) *Error {
	return New(http.StatusServiceUnavailable, CodeUnavailable, message)
}

// BadGateway is a 502 (upstream adapter failure).
func BadGateway(message string) *Error {
	return New(http.StatusBadGateway, CodeBadGateway, message)
}

// NotConfigured is a 503 when an optional integration/env is missing.
func NotConfigured(message string) *Error {
	return New(http.StatusServiceUnavailable, CodeNotConfigured, message)
}

// Internal is a 500.
func Internal(message string) *Error {
	return New(http.StatusInternalServerError, CodeInternal, message)
}

// InvalidTransition is a 400 for illegal state changes.
func InvalidTransition(message string) *Error {
	return New(http.StatusBadRequest, CodeInvalidTransition, message)
}

// AsError extracts *Error from err chain.
func AsError(err error) (*Error, bool) {
	var e *Error
	if errors.As(err, &e) {
		return e, true
	}
	return nil, false
}
