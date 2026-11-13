package farmmanager

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// LiveviewPath builds the FM asset_path for a printer still.
func LiveviewPath(sn string) string {
	sn = strings.TrimSpace(sn)
	return "liveview\\" + sn + "\\liveview.jpeg"
}

// LiveviewJPEG fetches GET /file?asset_path=liveview\{sn}\liveview.jpeg.
func (c *Client) LiveviewJPEG(ctx context.Context, sn string) ([]byte, error) {
	sn = strings.TrimSpace(sn)
	if sn == "" {
		return nil, fmt.Errorf("fm: liveview sn required")
	}
	q := url.Values{}
	q.Set("asset_path", LiveviewPath(sn))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base+"/file?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "image/jpeg")
	c.applyBBLHeaders(req)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("GET /file liveview: HTTP %d: %s", resp.StatusCode, truncate(string(data), 128))
	}
	if len(data) < 2 || data[0] != 0xff || data[1] != 0xd8 {
		return nil, fmt.Errorf("GET /file liveview: not a JPEG")
	}
	return data, nil
}
