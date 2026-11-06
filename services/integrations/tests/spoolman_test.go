package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jjheffernan/BambuFarmer/services/integrations/app/spoolman"
)

func TestListSpoolsMapsResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/spool" {
			t.Fatalf("path %s", r.URL.Path)
		}
		if r.URL.Query().Get("limit") == "" {
			t.Fatal("expected limit")
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{
				"id":               7,
				"used_weight":      12.5,
				"remaining_weight": 987.5,
				"location":         "Shelf A",
				"archived":         false,
				"filament": map[string]any{
					"id":        1,
					"name":      "PLA Black",
					"material":  "PLA",
					"color_hex": "111111",
					"vendor":    map[string]any{"name": "Polymaker"},
				},
			},
		})
	}))
	defer srv.Close()

	cli := spoolman.New(srv.URL, "", time.Second)
	spools, err := cli.ListSpools(context.Background(), 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(spools) != 1 || spools[0].SpoolmanID != 7 || spools[0].Material != "PLA" || spools[0].Vendor != "Polymaker" {
		t.Fatalf("%+v", spools)
	}
	if spools[0].RemainingG != 987.5 {
		t.Fatalf("remaining %v", spools[0].RemainingG)
	}
}
