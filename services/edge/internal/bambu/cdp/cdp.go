// Package cdp is the Farm Manager Client DevTools escape hatch (stub).
package cdp

import (
	"context"
	"fmt"
	"io"

	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/farm"
)

// Plane will drive FM Client via CDP. Not implemented yet.
type Plane struct {
	CDPURL string
}

func New(cdpURL string) *Plane {
	return &Plane{CDPURL: cdpURL}
}

func (p *Plane) Kind() farm.Kind { return farm.KindCDP }

func (p *Plane) Connect(context.Context) error {
	return fmt.Errorf("%w: CDP connect (url=%s)", farm.ErrNotImplemented, p.CDPURL)
}
func (p *Plane) Close() error                { return nil }
func (p *Plane) Ready(context.Context) error { return farm.ErrNotImplemented }

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
