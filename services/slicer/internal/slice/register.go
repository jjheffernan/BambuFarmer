package slice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// RegisterArtifact uploads a sliced .gcode.3mf to core POST /api/v1/artifacts.
// Returns artifact id. Empty coreURL skips (returns "", nil).
func RegisterArtifact(ctx context.Context, coreURL, outputPath, name string, do HTTPDo) (string, error) {
	coreURL = strings.TrimRight(strings.TrimSpace(coreURL), "/")
	if coreURL == "" {
		return "", nil
	}
	outputPath = strings.TrimSpace(outputPath)
	if outputPath == "" {
		return "", fmt.Errorf("output_path required to register artifact")
	}
	f, err := os.Open(outputPath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	filename := filepath.Base(outputPath)
	if filename == "" || filename == "." {
		filename = strings.TrimSpace(name)
		if filename == "" {
			filename = "plate.gcode.3mf"
		}
		if !strings.HasSuffix(strings.ToLower(filename), ".3mf") {
			filename += ".gcode.3mf"
		}
	}
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("file", filename)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(part, f); err != nil {
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
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("core artifact upload %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	var art struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &art); err != nil {
		return "", fmt.Errorf("decode artifact: %w", err)
	}
	if art.ID == "" {
		return "", fmt.Errorf("core returned empty artifact id")
	}
	return art.ID, nil
}
