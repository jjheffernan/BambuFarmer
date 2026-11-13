package farmmanager

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListTagsAndIndex(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tags" {
			t.Fatalf("path %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"tags": []map[string]any{
				{"id": "1", "name": "circuit-a", "devices": []string{"SN1", "SN2"}},
				{"id": "2", "dev_ids": []string{"SN1"}},
			},
		})
	}))
	defer srv.Close()
	cli, err := NewClient(ClientOptions{BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	cli.setToken("tok")
	tags, err := cli.ListTags(t.Context())
	if err != nil || len(tags) != 2 {
		t.Fatalf("%+v %v", tags, err)
	}
	idx := IndexByDevice(tags)
	if len(idx["SN1"]) != 2 || idx["SN2"][0] != "circuit-a" {
		t.Fatalf("%+v", idx)
	}
}
