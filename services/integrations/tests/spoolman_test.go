package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/events"
	"github.com/jjheffernan/BambuFarmer/services/integrations/app/spoolman"
)

func TestLowStock(t *testing.T) {
	if spoolman.LowGrams("") != 100 || spoolman.LowGrams("nope") != 100 || spoolman.LowGrams("50") != 50 {
		t.Fatal(spoolman.LowGrams(""))
	}
	got := spoolman.LowStock([]api.SpoolSummary{
		{ID: "a", RemainingG: 40},
		{ID: "b", RemainingG: 100},
		{ID: "c", RemainingG: 10, Archived: true},
	}, 100)
	if len(got) != 1 || got[0].ID != "a" {
		t.Fatalf("%+v", got)
	}
	if len(spoolman.LowStock(nil, 100)) != 0 {
		t.Fatal("nil spools")
	}
}

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

func TestUseWeightNoopWhenUnset(t *testing.T) {
	if err := spoolman.Consume(context.Background(), nil, spoolmanEnvelope(4, "9")); err != nil {
		t.Fatal(err)
	}
	cli := spoolman.New("", "", time.Second)
	if err := cli.UseWeight(context.Background(), "9", 4); err != nil {
		t.Fatal(err)
	}
}

func TestUseWeightPostsGrams(t *testing.T) {
	var gotMethod, gotPath string
	var got map[string]float64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	cli := spoolman.New(srv.URL, "", time.Second)
	if err := spoolman.Consume(context.Background(), cli, spoolmanEnvelope(12.5, "42")); err != nil {
		t.Fatal(err)
	}
	if gotMethod != http.MethodPut || gotPath != "/api/v1/spool/42/use" || got["use_weight"] != 12.5 {
		t.Fatalf("%s %s %#v", gotMethod, gotPath, got)
	}
}

func spoolmanEnvelope(grams float64, spoolID string) events.Envelope {
	return events.NewEnvelope(events.TypeFilamentConsumed, "core", "default", map[string]any{
		"grams": grams, "spool_id": spoolID,
	})
}
