package jobs

import (
	"errors"
	"testing"

	"github.com/jjheffernan/BambuFarmer/packages/api"
)

func TestReprintRequestCopiesPlate(t *testing.T) {
	req, err := ReprintRequest(api.Job{
		ID: "job-1", Name: "plate-a", ExternalSN: "SN1",
		ArtifactID: "art-9", Copies: 2, EstSeconds: 90, EstGrams: 12.5,
	})
	if err != nil {
		t.Fatal(err)
	}
	if req.Name != "plate-a (reprint)" || req.ArtifactID != "art-9" || req.FileID != "" {
		t.Fatalf("%+v", req)
	}
	if req.ExternalSN != "SN1" || req.Copies != 2 || req.EstSeconds != 90 || req.EstGrams != 12.5 {
		t.Fatalf("%+v", req)
	}
	if req.ID != "" {
		t.Fatal("reprint must allocate a new id")
	}
}

func TestReprintRequestKeepsFMFileID(t *testing.T) {
	req, err := ReprintRequest(api.Job{Name: "plate-a (reprint)", ExternalFileID: "503"})
	if err != nil {
		t.Fatal(err)
	}
	if req.Name != "plate-a (reprint)" || req.FileID != "503" {
		t.Fatalf("%+v", req)
	}
}

func TestReprintRequestRejectsEmptyPlate(t *testing.T) {
	_, err := ReprintRequest(api.Job{Name: "empty"})
	var apiErr *api.Error
	if !errors.As(err, &apiErr) || apiErr.Code != api.CodeBadRequest {
		t.Fatalf("got %v", err)
	}
}
