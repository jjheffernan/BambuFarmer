package farmmanager

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLiveviewJPEG(t *testing.T) {
	jpeg := []byte{0xff, 0xd8, 0xff, 0xd9}
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Query().Get("asset_path")
		w.Header().Set("Content-Type", "image/jpeg")
		_, _ = w.Write(jpeg)
	}))
	defer srv.Close()
	cli, err := NewClient(ClientOptions{BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	cli.setToken("tok")
	got, err := cli.LiveviewJPEG(t.Context(), "SN1")
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != `liveview\SN1\liveview.jpeg` || string(got) != string(jpeg) {
		t.Fatalf("%q %+v", gotPath, got)
	}
	if _, err := cli.LiveviewJPEG(t.Context(), " "); err == nil {
		t.Fatal("empty sn")
	}
}
