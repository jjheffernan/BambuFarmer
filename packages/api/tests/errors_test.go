package tests

import (
	"errors"
	"net/http"
	"testing"

	"github.com/jjheffernan/BambuFarmer/packages/api"
)

func TestErrorUnwrapAndAs(t *testing.T) {
	cause := errors.New("row missing")
	err := api.Wrap(http.StatusNotFound, api.CodeNotFound, "job not found", cause)
	if !errors.Is(err, cause) {
		t.Fatalf("expected unwrap to cause")
	}
	got, ok := api.AsError(err)
	if !ok || got.Code != api.CodeNotFound || got.HTTPStatus() != http.StatusNotFound {
		t.Fatalf("AsError: %+v ok=%v", got, ok)
	}
}

func TestConstructors(t *testing.T) {
	cases := []struct {
		err    *api.Error
		status int
		code   string
	}{
		{api.BadRequest("bad json"), http.StatusBadRequest, api.CodeBadRequest},
		{api.NotFound("gone"), http.StatusNotFound, api.CodeNotFound},
		{api.Conflict("exists"), http.StatusConflict, api.CodeConflict},
		{api.Unavailable("db down"), http.StatusServiceUnavailable, api.CodeUnavailable},
		{api.BadGateway("upstream"), http.StatusBadGateway, api.CodeBadGateway},
		{api.NotConfigured("no spoolman"), http.StatusServiceUnavailable, api.CodeNotConfigured},
		{api.InvalidTransition("a → b"), http.StatusBadRequest, api.CodeInvalidTransition},
	}
	for _, tc := range cases {
		if tc.err.HTTPStatus() != tc.status || tc.err.Code != tc.code {
			t.Fatalf("%v: status=%d code=%s", tc.err, tc.err.HTTPStatus(), tc.err.Code)
		}
		body := tc.err.Body()
		if body.Error == "" || body.Code != tc.code {
			t.Fatalf("body: %+v", body)
		}
	}
}
