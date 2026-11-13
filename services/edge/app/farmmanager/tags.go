package farmmanager

import (
	"context"
	"strings"

	"github.com/jjheffernan/BambuFarmer/packages/farm"
)

// Tag is a farm tag used for groups and circuit membership.
type Tag struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Devices []string `json:"devices,omitempty"`
	DevIDs  []string `json:"dev_ids,omitempty"`
}

// TagsResponse is GET /tags.
type TagsResponse struct {
	Tags []Tag `json:"tags"`
}

// Members returns printer serials for a tag.
func (t Tag) Members() []string {
	if len(t.Devices) > 0 {
		return t.Devices
	}
	return t.DevIDs
}

// Label is the display name, falling back to id.
func (t Tag) Label() string {
	if n := strings.TrimSpace(t.Name); n != "" {
		return n
	}
	return strings.TrimSpace(t.ID)
}

// ListTags calls GET /tags.
func (c *Client) ListTags(ctx context.Context) ([]Tag, error) {
	var resp TagsResponse
	if err := c.doJSON(ctx, "GET", "/tags", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Tags, nil
}

// ApplyDeviceTags sets Tags on each report from membership or device-local tags.
func ApplyDeviceTags(reports []farm.DeviceReport, tags []Tag) {
	idx := IndexByDevice(tags)
	for i := range reports {
		sn := reports[i].ExternalSN
		if len(reports[i].Tags) == 0 {
			reports[i].Tags = append([]string(nil), idx[sn]...)
			continue
		}
		seen := map[string]bool{}
		for _, t := range reports[i].Tags {
			seen[t] = true
		}
		for _, t := range idx[sn] {
			if !seen[t] {
				reports[i].Tags = append(reports[i].Tags, t)
			}
		}
	}
}

// IndexByDevice maps each SN to tag labels that include it.
func IndexByDevice(tags []Tag) map[string][]string {
	out := map[string][]string{}
	for _, t := range tags {
		label := t.Label()
		if label == "" {
			continue
		}
		for _, sn := range t.Members() {
			sn = strings.TrimSpace(sn)
			if sn == "" {
				continue
			}
			out[sn] = append(out[sn], label)
		}
	}
	return out
}
