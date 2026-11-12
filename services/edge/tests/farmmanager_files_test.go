package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jjheffernan/BambuFarmer/services/edge/app/farmmanager"
)

func TestUpload3MF(t *testing.T) {
	var gotProgress string
	var gotCT string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/file/upload3mf" || r.Method != http.MethodPost {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		gotProgress = r.Header.Get("x-bbl-progress-id")
		gotCT = r.Header.Get("Content-Type")
		if !strings.HasPrefix(gotCT, "multipart/form-data") {
			t.Fatalf("content-type %q", gotCT)
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatal(err)
		}
		if r.FormValue("folder_id") != "1" {
			t.Fatalf("folder_id=%q", r.FormValue("folder_id"))
		}
		f, hdr, err := r.FormFile("file")
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		if hdr.Filename != "plate.gcode.3mf" {
			t.Fatalf("filename=%q", hdr.Filename)
		}
		b, _ := io.ReadAll(f)
		if string(b) != "3mf-bytes" {
			t.Fatalf("body=%q", b)
		}
		_ = json.NewEncoder(w).Encode(farmmanager.Upload3MFResponse{
			Code: 0,
			F3MFInfo: farmmanager.F3MFInfo{
				ID:   "503",
				Name: "plate.gcode.3mf",
			},
		})
	}))
	defer srv.Close()

	cli, err := farmmanager.NewClient(farmmanager.ClientOptions{BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	cli.SetServerID("sid-1")
	info, err := cli.Upload3MF(context.Background(), "plate.gcode.3mf", bytes.NewReader([]byte("3mf-bytes")), "")
	if err != nil {
		t.Fatal(err)
	}
	if info.ID != "503" {
		t.Fatalf("%+v", info)
	}
	if gotProgress == "" || !strings.Contains(gotProgress, "-") {
		t.Fatalf("progress id %q", gotProgress)
	}
}

func TestCreateTask(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/task" || r.Method != http.MethodPost {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		var body farmmanager.CreateTaskRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.F3MFID != "503" || body.TaskPrintModel != 1 {
			t.Fatalf("%+v", body)
		}
		if len(body.DevicePool) != 1 || body.DevicePool[0] != "SN1" {
			t.Fatalf("pool=%v", body.DevicePool)
		}
		w.Write([]byte(`{"code":0,"message":"","data":{"task_id":15,"f3mf_id":"503"}}`))
	}))
	defer srv.Close()

	cli, err := farmmanager.NewClient(farmmanager.ClientOptions{BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	ref, err := cli.CreateTask(context.Background(), farmmanager.CreateDirectPrint("503", "SN1", "job-1"))
	if err != nil {
		t.Fatal(err)
	}
	if ref.TaskID != "15" || ref.F3MFID != "503" {
		t.Fatalf("%+v", ref)
	}
}
