package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jjheffernan/BambuFarmer/services/integrations/app/httpapi"
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

func TestImportSkipsWhenUnset(t *testing.T) {
	mux := http.NewServeMux()
	httpapi.Register(mux, httpapi.Deps{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/library/import", strings.NewReader(`{"name":"Benchy"}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK || got["imported"] != false {
		t.Fatalf("%d %+v", rec.Code, got)
	}
}

func TestCreateModelPostsName(t *testing.T) {
	var got map[string]string
	var method string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		if r.URL.Path != "/models" {
			t.Fatalf("path %s", r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&got)
		_ = json.NewEncoder(w).Encode(map[string]string{"@id": "/models/mf9", "name": got["name"]})
	}))
	defer srv.Close()
	id, err := manyfold.New(srv.URL, "pat-test", "", "", time.Second).CreateModel(context.Background(), " Benchy ")
	if err != nil || id != "/models/mf9" || method != http.MethodPost || got["name"] != "Benchy" {
		t.Fatalf("%s %q %+v %v", method, id, got, err)
	}
	if _, err := manyfold.New("", "", "", "", time.Second).CreateModel(context.Background(), "Benchy"); err == nil {
		t.Fatal("empty url")
	}
}
