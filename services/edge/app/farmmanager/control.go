package farmmanager

import (
	"context"
	"fmt"

	"github.com/jjheffernan/BambuFarmer/packages/api"
)

// Control posts POST /device/{sn}/opt with { "opt": "..." }.
func (c *Client) Control(ctx context.Context, sn string, opt api.ControlOpt) error {
	if sn == "" {
		return api.BadRequest("control: empty sn")
	}
	switch opt {
	case api.ControlPause, api.ControlResume, api.ControlStop, api.ControlBedClean:
	default:
		return api.BadRequest(fmt.Sprintf("control: unsupported opt %q", opt))
	}
	return c.doJSON(ctx, "POST", "/device/"+sn+"/opt", OptBody{Opt: string(opt)}, nil)
}
