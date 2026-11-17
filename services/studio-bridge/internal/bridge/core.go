package bridge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
)

// HTTPDo is injectable for tests.
type HTTPDo func(req *http.Request) (*http.Response, error)

// UploadArtifact POSTs multipart file to core. Empty coreURL returns ("", nil).
func UploadArtifact(ctx context.Context, coreURL, filename string, body []byte, do HTTPDo) (string, error) {
	coreURL = strings.TrimRight(strings.TrimSpace(coreURL), "/")
	if coreURL == "" {
		return "", nil
	}
	if len(body) == 0 {
		return "", fmt.Errorf("empty plate")
	}
	filename = filepath.Base(strings.TrimSpace(filename))
	if filename == "" || filename == "." {
		filename = "plate.gcode.3mf"
	}
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("file", filename)
	if err != nil {
		return "", err
	}
	if _, err := part.Write(body); err != nil {
		return "", err
	}
	_ = w.Close()
	if do == nil {
		do = http.DefaultClient.Do
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, coreURL+"/api/v1/artifacts", &buf)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("core artifact upload %s: %s", resp.Status, strings.TrimSpace(string(raw)))
	}
	var art struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(raw, &art); err != nil {
		return "", fmt.Errorf("decode artifact: %w", err)
	}
	if art.ID == "" {
		return "", fmt.Errorf("core returned empty artifact id")
	}
	return art.ID, nil
}

// CreateJob POSTs a queued job for an artifact. Empty coreURL returns ("", nil).
func CreateJob(ctx context.Context, coreURL, name, artifactID, targetModel string, do HTTPDo) (string, error) {
	coreURL = strings.TrimRight(strings.TrimSpace(coreURL), "/")
	if coreURL == "" {
		return "", nil
	}
	artifactID = strings.TrimSpace(artifactID)
	if artifactID == "" {
		return "", fmt.Errorf("artifact_id required")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		name = "studio-bridge"
	}
	payload := map[string]any{
		"name":        name,
		"artifact_id": artifactID,
	}
	if tm := strings.TrimSpace(targetModel); tm != "" {
		payload["target_model"] = tm
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	if do == nil {
		do = http.DefaultClient.Do
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, coreURL+"/api/v1/jobs", bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("core create job %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	var job struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &job); err != nil {
		return "", fmt.Errorf("decode job: %w", err)
	}
	return job.ID, nil
}
