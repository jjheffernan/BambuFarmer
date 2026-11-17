// Package cdp is the Farm Manager Client DevTools escape hatch (probe + stub control).
package cdp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/farm"
)

// Plane will drive FM Client via CDP. Control methods remain stubbed;
// Ready/Connect probe CDP_URL/json/version so edge status can show reachability.
type Plane struct {
	CDPURL string
	client *http.Client
}

func New(cdpURL string) *Plane {
	return &Plane{
		CDPURL: strings.TrimRight(strings.TrimSpace(cdpURL), "/"),
		client: &http.Client{Timeout: 3 * time.Second},
	}
}

func (p *Plane) Kind() farm.Kind { return farm.KindCDP }

func (p *Plane) Connect(ctx context.Context) error {
	return p.Ready(ctx)
}

func (p *Plane) Close() error { return nil }

func (p *Plane) Ready(ctx context.Context) error {
	if p.CDPURL == "" {
		return fmt.Errorf("%w: CDP_URL unset", farm.ErrNotConfigured)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.CDPURL+"/json/version", nil)
	if err != nil {
		return err
	}
	res, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("CDP probe: %w", err)
	}
	defer res.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, 1<<16))
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("CDP probe: HTTP %d", res.StatusCode)
	}
	return nil
}

func (p *Plane) Captain(context.Context) (api.CaptainInfo, error) {
	return api.CaptainInfo{}, farm.ErrNotImplemented
}
func (p *Plane) ListDevices(context.Context) ([]farm.DeviceReport, error) {
	return nil, farm.ErrNotImplemented
}
func (p *Plane) Control(context.Context, string, api.ControlOpt) error {
	return farm.ErrNotImplemented
}
func (p *Plane) Upload3MF(context.Context, string, io.Reader) (farm.FileRef, error) {
	return farm.FileRef{}, farm.ErrNotImplemented
}
func (p *Plane) StartPrint(context.Context, farm.PrintRequest) (farm.ExternalJobRef, error) {
	return farm.ExternalJobRef{}, farm.ErrNotImplemented
}
func (p *Plane) LiveviewJPEG(context.Context, string) ([]byte, error) {
	return nil, farm.ErrNotImplemented
}

var _ farm.ControlPlane = (*Plane)(nil)
