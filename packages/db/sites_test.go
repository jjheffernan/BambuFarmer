package db

import (
	"testing"

	"github.com/jjheffernan/BambuFarmer/packages/api"
)

func TestSiteDTOFields(t *testing.T) {
	s := api.Site{ID: "default", Name: "Default Site", Timezone: "UTC"}
	if s.ID == "" || s.Name == "" {
		t.Fatal(s)
	}
}

func TestUpdateSiteValidation(t *testing.T) {
	_, err := UpdateSite(t.Context(), nil, "", "Lab", "UTC")
	if err == nil {
		t.Fatal("expected id required")
	}
	_, err = UpdateSite(t.Context(), nil, "default", "  ", "UTC")
	if err == nil {
		t.Fatal("expected name required")
	}
}
