package fm

import (
	"context"
	"fmt"
)

// ListDevices calls GET /devices (falls back to /devices2 on failure).
func (c *Client) ListDevices(ctx context.Context) ([]Device, error) {
	var resp DevicesResponse
	if err := c.doJSON(ctx, "GET", "/devices", nil, &resp); err != nil {
		var resp2 DevicesResponse
		if err2 := c.doJSON(ctx, "GET", "/devices2", nil, &resp2); err2 != nil {
			return nil, fmt.Errorf("devices: %v; devices2: %v", err, err2)
		}
		return resp2.Devices, nil
	}
	return resp.Devices, nil
}

// GetDevice calls GET /device/{dev_id}.
func (c *Client) GetDevice(ctx context.Context, sn string) (Device, error) {
	var wrap struct {
		DevInfo Device `json:"dev_info"`
		Device  Device `json:"device"`
	}
	if err := c.doJSON(ctx, "GET", "/device/"+sn, nil, &wrap); err != nil {
		return Device{}, err
	}
	if wrap.DevInfo.DevID != "" {
		return wrap.DevInfo, nil
	}
	return wrap.Device, nil
}
