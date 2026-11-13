package db

import (
	"testing"

	"github.com/jjheffernan/BambuFarmer/packages/api"
)

func TestRestoreRequiresVersion(t *testing.T) {
	_, err := RestoreFilamentPresetRevision(t.Context(), nil, "default", "x", 0)
	if ae, ok := api.AsError(err); !ok || ae.Code != api.CodeBadRequest {
		t.Fatalf("%v", err)
	}
	_, err = RestorePrintPresetRevision(t.Context(), nil, "default", "x", -1)
	if ae, ok := api.AsError(err); !ok || ae.Code != api.CodeBadRequest {
		t.Fatalf("%v", err)
	}
}
