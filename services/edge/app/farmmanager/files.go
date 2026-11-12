package farmmanager

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path"
	"time"
)

// Upload3MFResponse is POST /file/upload3mf success body.
type Upload3MFResponse struct {
	Code     int    `json:"code"`
	Message  string `json:"message"`
	F3MFInfo F3MFInfo `json:"f3mf_info"`
}

// F3MFInfo is the staged file record returned by upload / list.
type F3MFInfo struct {
	ID       string `json:"id"`
	Name     string `json:"f3mf_name"`
	FolderID string `json:"folder_id"`
}

// Upload3MF posts multipart form to POST /file/upload3mf.
// folderID defaults to "1" when empty. Requires a Bambu-sliced .gcode.3mf.
func (c *Client) Upload3MF(ctx context.Context, name string, r io.Reader, folderID string) (F3MFInfo, error) {
	if name == "" {
		name = "plate.gcode.3mf"
	}
	if folderID == "" {
		folderID = "1"
	}

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile("file", path.Base(name))
	if err != nil {
		return F3MFInfo{}, err
	}
	if _, err := io.Copy(part, r); err != nil {
		return F3MFInfo{}, fmt.Errorf("read upload body: %w", err)
	}
	if err := w.WriteField("folder_id", folderID); err != nil {
		return F3MFInfo{}, err
	}
	if err := w.Close(); err != nil {
		return F3MFInfo{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/file/upload3mf", &body)
	if err != nil {
		return F3MFInfo{}, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-bbl-progress-id", newProgressID())
	c.applyBBLHeaders(req)

	cli := &http.Client{
		Timeout:   5 * time.Minute,
		Transport: c.http.Transport,
	}
	resp, err := cli.Do(req)
	if err != nil {
		return F3MFInfo{}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return F3MFInfo{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return F3MFInfo{}, fmt.Errorf("POST /file/upload3mf: HTTP %d: %s", resp.StatusCode, truncate(string(data), 256))
	}
	var out Upload3MFResponse
	if err := json.Unmarshal(data, &out); err != nil {
		return F3MFInfo{}, fmt.Errorf("decode upload3mf: %w", err)
	}
	if out.F3MFInfo.ID == "" {
		return F3MFInfo{}, fmt.Errorf("upload3mf: empty f3mf_info.id")
	}
	return out.F3MFInfo, nil
}

func newProgressID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
