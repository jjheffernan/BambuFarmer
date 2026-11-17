package slice

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRegisterArtifactSkipsEmptyCore(t *testing.T) {
	id, err := RegisterArtifact(context.Background(), "", "/tmp/x", "n", nil)
	if err != nil || id != "" {
		t.Fatalf("got %q %v", id, err)
	}
}

func TestRegisterArtifactUpload(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.gcode.3mf")
	if err := os.WriteFile(path, []byte("3mf"), 0o644); err != nil {
		t.Fatal(err)
	}
	do := func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPost || !strings.HasSuffix(req.URL.Path, "/api/v1/artifacts") {
			t.Fatalf("unexpected %s %s", req.Method, req.URL)
		}
		ct := req.Header.Get("Content-Type")
		if !strings.Contains(ct, "multipart/") {
			t.Fatalf("content-type %s", ct)
		}
		_, _ = io.Copy(io.Discard, req.Body)
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(`{"id":"art1","filename":"out.gcode.3mf"}`)),
			Header:     make(http.Header),
		}, nil
	}
	id, err := RegisterArtifact(context.Background(), "http://core:8080", path, "Benchy", do)
	if err != nil || id != "art1" {
		t.Fatalf("got %q %v", id, err)
	}
}
