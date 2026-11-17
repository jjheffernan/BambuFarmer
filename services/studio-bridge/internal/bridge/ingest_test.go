package bridge

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestParsePolicy(t *testing.T) {
	if ParsePolicy("archive") != PolicyArchive {
		t.Fatal("archive")
	}
	if ParsePolicy("") != PolicyQueue {
		t.Fatal("default queue")
	}
}

func TestIngestArchiveAndQueue(t *testing.T) {
	var calls []string
	do := func(req *http.Request) (*http.Response, error) {
		calls = append(calls, req.Method+" "+req.URL.Path)
		if strings.HasSuffix(req.URL.Path, "/artifacts") {
			_, _ = io.Copy(io.Discard, req.Body)
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(`{"id":"art1"}`)),
				Header:     make(http.Header),
			}, nil
		}
		if strings.HasSuffix(req.URL.Path, "/jobs") {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(`{"id":"job1"}`)),
				Header:     make(http.Header),
			}, nil
		}
		t.Fatalf("unexpected %s", req.URL)
		return nil, nil
	}
	res, err := Ingest(context.Background(), "http://core:8080", PolicyArchive, Request{
		Filename: "benchy.gcode.3mf",
		Body:     []byte("3mf"),
	}, do)
	if err != nil || res.ArtifactID != "art1" || res.JobID != "" {
		t.Fatalf("%+v %v", res, err)
	}
	q := true
	res, err = Ingest(context.Background(), "http://core:8080", PolicyArchive, Request{
		Filename: "benchy.gcode.3mf",
		Body:     []byte("3mf"),
		Queue:    &q,
	}, do)
	if err != nil || res.JobID != "job1" {
		t.Fatalf("%+v %v", res, err)
	}
	if len(calls) < 3 {
		t.Fatalf("calls %v", calls)
	}
}

func TestIngestNoCore(t *testing.T) {
	res, err := Ingest(context.Background(), "", PolicyQueue, Request{
		Filename: "a.3mf",
		Body:     []byte("x"),
	}, nil)
	if err != nil || !res.Received || res.ArtifactID != "" {
		t.Fatalf("%+v %v", res, err)
	}
}
