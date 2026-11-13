// Package native implements farm.ControlPlane against Farm Manager Server (Client-shaped REST).
package native

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/farm"
	"github.com/jjheffernan/BambuFarmer/services/edge/app/farmmanager"
)

// Config for the FM-native backend.
type Config struct {
	BaseURL     string
	User        string
	Password    string
	TLSInsecure bool
	CertDir     string
}

// Plane is a farm.ControlPlane backed by farmmanager.Client.
type Plane struct {
	cfg    Config
	client *farmmanager.Client
}

// New constructs an unconnected native plane.
func New(cfg Config) (*Plane, error) {
	if cfg.BaseURL == "" {
		return nil, farm.ErrNotConfigured
	}
	cli, err := farmmanager.NewClient(farmmanager.ClientOptions{
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
		rep := mapDevice(d)
		rep.Tags = append([]string(nil), d.Tags...)
		out = append(out, rep)
	}
	if tags, err := p.client.ListTags(ctx); err == nil {
		farmmanager.ApplyDeviceTags(out, tags)
	}
	return out, nil
}

// ListTags returns farm tags for circuit and group membership.
func (p *Plane) ListTags(ctx context.Context) ([]farmmanager.Tag, error) {
	if err := p.client.EnsureSession(ctx, p.cfg.User, p.cfg.Password); err != nil {
		return nil, err
	}
	return p.client.ListTags(ctx)
}

func (p *Plane) Control(ctx context.Context, sn string, opt api.ControlOpt) error {
	if err := p.client.EnsureSession(ctx, p.cfg.User, p.cfg.Password); err != nil {
		return err
	}
	return p.client.Control(ctx, sn, opt)
}

func (p *Plane) Upload3MF(ctx context.Context, name string, r io.Reader) (farm.FileRef, error) {
	if err := p.client.EnsureSession(ctx, p.cfg.User, p.cfg.Password); err != nil {
		return farm.FileRef{}, err
	}
	info, err := p.client.Upload3MF(ctx, name, r, "")
	if err != nil {
		return farm.FileRef{}, err
	}
	return farm.FileRef{ID: info.ID, Name: firstNonEmpty(info.Name, name)}, nil
}

func (p *Plane) StartPrint(ctx context.Context, req farm.PrintRequest) (farm.ExternalJobRef, error) {
	if err := p.client.EnsureSession(ctx, p.cfg.User, p.cfg.Password); err != nil {
		return farm.ExternalJobRef{}, err
	}
	f3mfID := req.FileRef.ID
	if f3mfID == "" {
		return farm.ExternalJobRef{}, fmt.Errorf("StartPrint: FileRef.ID (f3mf_id) required")
	}
	if req.ExternalSN == "" {
		return farm.ExternalJobRef{}, fmt.Errorf("StartPrint: empty ExternalSN")
	}
	taskName := req.Name
	if taskName == "" {
		taskName = req.JobID
	}
	if taskName == "" {
		taskName = req.FileRef.Name
	}
	ref, err := p.client.CreateTask(ctx, farmmanager.CreateDirectPrint(f3mfID, req.ExternalSN, taskName))
	if err != nil {
		return farm.ExternalJobRef{}, err
	}
	return farm.ExternalJobRef{
		TaskID:    ref.TaskID,
		SubtaskID: ref.SubtaskID,
		Raw:       req.JobID,
	}, nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func (p *Plane) LiveviewJPEG(ctx context.Context, sn string) ([]byte, error) {
	if err := p.client.EnsureSession(ctx, p.cfg.User, p.cfg.Password); err != nil {
		return nil, err
	}
	return p.client.LiveviewJPEG(ctx, sn)
}

func mapDevice(d farmmanager.Device) farm.DeviceReport {
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
		rep.HMS = farm.ParseHMSJSON(rs.HMS)
		rep.AMS = farm.ParseAMSJSON(rs.AMS)
		rep.TaskID = rs.TaskID
		rep.SubtaskName = rs.SubtaskName
		rep.WifiSignal = rs.WifiSignal
		rep.NozzleDiameter = rs.NozzleDiameter
	}
	return rep
}

// Ensure interface compliance.
var _ farm.ControlPlane = (*Plane)(nil)
