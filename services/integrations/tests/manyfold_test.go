package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jjheffernan/BambuFarmer/services/integrations/app/manyfold"
)

func TestListModelsMapsHydraCollection(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/models" {
			t.Fatalf("path %s", r.URL.Path)
		}
		if r.Header.Get("Accept") != "application/vnd.manyfold.v0+json" {
			t.Fatalf("accept %q", r.Header.Get("Accept"))
		}
		if r.Header.Get("Authorization") != "Bearer pat-test" {
			t.Fatalf("auth %q", r.Header.Get("Authorization"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"@type":      "hydra:Collection",
			"totalItems": 1,
			"member": []map[string]any{
				{"@id": "/models/mf1", "name": "Benchy", "@type": "3DModel"},
			},
		})
	}))
	defer srv.Close()

	cli := manyfold.New(srv.URL, "pat-test", "", "", time.Second)
	models, err := cli.ListModels(context.Background(), 1, "name")
	if err != nil {
		t.Fatal(err)
	}
	if len(models) != 1 || models[0].Name != "Benchy" {
		t.Fatalf("%+v", models)
	}
	if models[0].URL != srv.URL+"/models/mf1" {
		t.Fatalf("url %s", models[0].URL)
	}
}
