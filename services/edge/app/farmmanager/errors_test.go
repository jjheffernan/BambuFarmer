package farmmanager

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/jjheffernan/BambuFarmer/packages/api"
)

func TestHTTPErrorHint(t *testing.T) {
	e := NewHTTPError("GET", "/devices", http.StatusNotFound, `{"message":"nope"}`)
	if e.Hint == "" || !strings.Contains(e.Hint, "API drift") {
		t.Fatalf("hint %q", e.Hint)
	}
	if !strings.Contains(e.Error(), "HTTP 404") {
		t.Fatalf("err %q", e.Error())
	}
}

func TestMapToAPIHTTP(t *testing.T) {
	err := MapToAPI("control", NewHTTPError("POST", "/device/SN/opt", 401, "auth"))
	ae, ok := api.AsError(err)
	if !ok {
		t.Fatal(err)
	}
	if ae.Code != api.CodeBadGateway {
		t.Fatalf("code %s", ae.Code)
	}
	d, _ := ae.Details.(map[string]any)
	if d["hint"] == nil || d["upstream_status"] != 401 {
		t.Fatalf("%#v", d)
	}
}

func TestMapToAPIDecode(t *testing.T) {
	err := MapToAPI("devices", &DecodeError{
		Path: "/devices",
		Err:  errors.New("unexpected end of JSON"),
		Hint: "shape changed",
	})
	ae, ok := api.AsError(err)
	if !ok || !strings.Contains(ae.Message, "devices failed") {
		t.Fatalf("%v", err)
	}
	body, _ := json.Marshal(ae.Body())
	if !strings.Contains(string(body), "shape changed") {
		t.Fatalf("%s", body)
	}
}

func TestHintForStatusAuth(t *testing.T) {
	h := HintForStatus(http.StatusUnauthorized, "/login")
	if !strings.Contains(h, "auth") {
		t.Fatal(h)
	}
}
