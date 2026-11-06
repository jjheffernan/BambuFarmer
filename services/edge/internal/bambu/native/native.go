// Package native implements farm.ControlPlane against Farm Manager Server (Client-shaped REST).
package native

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/farm"
	"github.com/jjheffernan/BambuFarmer/services/edge/internal/fm"
)

// Config for the FM-native backend.
type Config struct {
	BaseURL     string
	User        string
	Password    string
	TLSInsecure bool
	CertDir     string
}

// Plane is a farm.ControlPlane backed by fm.Client.
type Plane struct {
	cfg    Config
	client *fm.Client
}

// New constructs an unconnected native plane.
func New(cfg Config) (*Plane, error) {
	if cfg.BaseURL == "" {
		return nil, farm.ErrNotConfigured
	}
	cli, err := fm.NewClient(fm.ClientOptions{
		BaseURL:     cfg.BaseURL,
		TLSInsecure: cfg.TLSInsecure,
		CertDir:     cfg.CertDir,
	})
	if err != nil {
		return nil, err
	}
	return &Plane{cfg: cfg, client: cli}, nil
}

func (p *Plane) Kind() farm.Kind { return farm.KindNative }

func (p *Plane) Connect(ctx context.Context) error {
	if err := p.client.EnsureSession(ctx, p.cfg.User, p.cfg.Password); err != nil {
		return err
	}
	_, err := p.client.Captain(ctx)
	return err
}

func (p *Plane) Close() error { return nil }

func (p *Plane) Ready(ctx context.Context) error {
	if !p.client.Connected() {
		return farm.ErrNotConnected
	}
	return nil
}

func (p *Plane) Captain(ctx context.Context) (api.CaptainInfo, error) {
	if err := p.client.EnsureSession(ctx, p.cfg.User, p.cfg.Password); err != nil {
		return api.CaptainInfo{}, err
	}
	c, err := p.client.Captain(ctx)
	if err != nil {
		return api.CaptainInfo{}, err
	}
	return c.ToAPI(), nil
}

func (p *Plane) ListDevices(ctx context.Context) ([]farm.DeviceReport, error) {
	if err := p.client.EnsureSession(ctx, p.cfg.User, p.cfg.Password); err != nil {
		return nil, err
	}
	devices, err := p.client.ListDevices(ctx)
	if err != nil {
		// one re-login retry
		_ = p.client.Login(ctx, p.cfg.User, p.cfg.Password)
		devices, err = p.client.ListDevices(ctx)
		if err != nil {
			return nil, err
		}
	}
	out := make([]farm.DeviceReport, 0, len(devices))
	for _, d := range devices {
		out = append(out, mapDevice(d))
	}
	return out, nil
}

func (p *Plane) Control(ctx context.Context, sn string, opt api.ControlOpt) error {
	if err := p.client.EnsureSession(ctx, p.cfg.User, p.cfg.Password); err != nil {
		return err
	}
	return p.client.Control(ctx, sn, opt)
}

func (p *Plane) Upload3MF(ctx context.Context, name string, r io.Reader) (farm.FileRef, error) {
	_ = ctx
	_ = name
	_ = r
	return farm.FileRef{}, fmt.Errorf("%w: Upload3MF", farm.ErrNotImplemented)
}

func (p *Plane) StartPrint(ctx context.Context, req farm.PrintRequest) (farm.ExternalJobRef, error) {
	_ = ctx
	_ = req
	return farm.ExternalJobRef{}, fmt.Errorf("%w: StartPrint", farm.ErrNotImplemented)
}

func (p *Plane) LiveviewJPEG(ctx context.Context, sn string) ([]byte, error) {
	_ = ctx
	_ = sn
	return nil, fmt.Errorf("%w: LiveviewJPEG", farm.ErrNotImplemented)
}

func mapDevice(d fm.Device) farm.DeviceReport {
	online := d.Online != nil && *d.Online
	rep := farm.DeviceReport{
		ExternalSN: d.DevID,
		Name:       d.Name,
		Model:      d.DevModel,
		DevIP:      d.DevIP,
		Online:     online,
		ObservedAt: time.Now().UTC(),
	}
	if rep.Name == "" {
		rep.Name = d.DevName
	}
	if rep.Name == "" {
		rep.Name = d.DevID
	}
	if d.ReportStatus != nil {
		rs := d.ReportStatus
		if !online && rs.GcodeState != "" {
			rep.Online = true
		}
		rep.GcodeState = rs.GcodeState
		rep.MCPercent = rs.MCPercent
		rep.MCRemainingTime = rs.MCRemainingTime
		rep.NozzleTemper = rs.NozzleTemper
		rep.NozzleTargetTemper = rs.NozzleTargetTemper
		rep.BedTemper = rs.BedTemper
		rep.BedTargetTemper = rs.BedTargetTemper
		rep.ChamberTemper = rs.ChamberTemper
		rep.LayerNum = rs.LayerNum
		rep.TotalLayerNum = rs.TotalLayerNum
		rep.PrintError = rs.PrintError
		rep.WifiSignal = rs.WifiSignal
		rep.NozzleDiameter = rs.NozzleDiameter
	}
	return rep
}

// Ensure interface compliance.
var _ farm.ControlPlane = (*Plane)(nil)
