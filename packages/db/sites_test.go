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
