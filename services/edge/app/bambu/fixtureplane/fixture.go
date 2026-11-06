// Package fixtureplane implements farm.ControlPlane from redacted FM JSON fixtures.
package fixtureplane

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/farm"
	"github.com/jjheffernan/BambuFarmer/services/edge/app/fixture"
	"github.com/jjheffernan/BambuFarmer/services/edge/app/farmmanager"
)

// Plane cycles fixtures/fm device JSON.
type Plane struct {
	src *fixture.Source
	dir string
	log *slog.Logger
}

// New opens a fixture directory.
func New(dir string, log *slog.Logger) (*Plane, error) {
	if log == nil {
		log = slog.Default()
	}
	src, err := fixture.New(dir)
	if err != nil {
		return nil, err
	}
	return &Plane{src: src, dir: dir, log: log}, nil
}

func (p *Plane) Kind() farm.Kind { return farm.KindFixture }

func (p *Plane) Connect(context.Context) error { return nil }
func (p *Plane) Close() error                   { return nil }
func (p *Plane) Ready(context.Context) error    { return nil }

func (p *Plane) Captain(context.Context) (api.CaptainInfo, error) {
	return api.CaptainInfo{
		Name:     "fixture",
		ServerID: "fixture-server",
		Release:  "fixture",
	}, nil
}

func (p *Plane) ListDevices(context.Context) ([]farm.DeviceReport, error) {
	devices, err := p.src.NextDevices()
	if err != nil {
		return nil, err
	}
	out := make([]farm.DeviceReport, 0, len(devices))
	for _, d := range devices {
		out = append(out, mapDevice(d))
	}
	return out, nil
}

func (p *Plane) Control(_ context.Context, sn string, opt api.ControlOpt) error {
	p.log.Info("fixture control", "sn", sn, "opt", opt)
	return nil
}

func (p *Plane) Upload3MF(context.Context, string, io.Reader) (farm.FileRef, error) {
	return farm.FileRef{ID: "fixture-file", Name: "fixture.gcode.3mf"}, nil
}

func (p *Plane) StartPrint(_ context.Context, req farm.PrintRequest) (farm.ExternalJobRef, error) {
	p.log.Info("fixture start_print", "sn", req.ExternalSN, "job_id", req.JobID)
	return farm.ExternalJobRef{TaskID: "fixture-task", Raw: req.JobID}, nil
}

func (p *Plane) LiveviewJPEG(context.Context, string) ([]byte, error) {
	return nil, fmt.Errorf("%w: LiveviewJPEG", farm.ErrNotImplemented)
}

func mapDevice(d farmmanager.Device) farm.DeviceReport {
	rep := farm.DeviceReport{
		ExternalSN: d.DevID,
		Name:       d.Name,
		Model:      d.DevModel,
		DevIP:      d.DevIP,
		ObservedAt: time.Now().UTC(),
	}
	if rep.Name == "" {
		rep.Name = d.DevName
	}
	if rep.Name == "" {
		rep.Name = d.DevID
	}
	if d.Online != nil {
		rep.Online = *d.Online
	}
	if d.ReportStatus != nil {
		rs := d.ReportStatus
		if rs.GcodeState != "" {
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

var _ farm.ControlPlane = (*Plane)(nil)
