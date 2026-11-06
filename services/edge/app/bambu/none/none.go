package none

import (
	"context"
	"io"

	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/farm"
)

// Plane is a no-op control plane when neither FM URL nor fixtures are configured.
type Plane struct{}

func New() *Plane { return &Plane{} }

func (p *Plane) Kind() farm.Kind                  { return farm.KindNone }
func (p *Plane) Connect(context.Context) error    { return nil }
func (p *Plane) Close() error                     { return nil }
func (p *Plane) Ready(context.Context) error      { return farm.ErrNotConfigured }
func (p *Plane) Captain(context.Context) (api.CaptainInfo, error) {
	return api.CaptainInfo{}, farm.ErrNotConfigured
}
func (p *Plane) ListDevices(context.Context) ([]farm.DeviceReport, error) {
	return nil, farm.ErrNotConfigured
}
func (p *Plane) Control(context.Context, string, api.ControlOpt) error {
	return farm.ErrNotConfigured
}
func (p *Plane) Upload3MF(context.Context, string, io.Reader) (farm.FileRef, error) {
	return farm.FileRef{}, farm.ErrNotConfigured
}
func (p *Plane) StartPrint(context.Context, farm.PrintRequest) (farm.ExternalJobRef, error) {
	return farm.ExternalJobRef{}, farm.ErrNotConfigured
}
func (p *Plane) LiveviewJPEG(context.Context, string) ([]byte, error) {
	return nil, farm.ErrNotConfigured
}

var _ farm.ControlPlane = (*Plane)(nil)
