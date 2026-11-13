package tests

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/httpserver"
)

func TestWriteErrorTyped(t *testing.T) {
	rec := httptest.NewRecorder()
	httpserver.WriteError(rec, api.NotFound("job missing"))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status %d", rec.Code)
	}
	var body api.ErrorBody
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Code != api.CodeNotFound || body.Error != "job missing" {
		t.Fatalf("body %+v", body)
	}
}

func TestWriteErrorPlain(t *testing.T) {
	rec := httptest.NewRecorder()
	httpserver.WriteError(rec, errors.New("boom"))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status %d", rec.Code)
	}
	var body api.ErrorBody
	_ = json.NewDecoder(rec.Body).Decode(&body)
	if body.Code != api.CodeInternal || body.Error != "boom" {
		t.Fatalf("body %+v", body)
	}
}

func TestMapErrorWrapped(t *testing.T) {
	err := api.Wrap(http.StatusBadGateway, api.CodeBadGateway, "spoolman failed", errors.New("timeout"))
	status, body := httpserver.MapError(err)
	if status != http.StatusBadGateway || body.Code != api.CodeBadGateway {
		t.Fatalf("%d %+v", status, body)
	}
}
