// Package fixtureplane implements farm.ControlPlane from redacted FM JSON fixtures.
package fixtureplane

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/farm"
	"github.com/jjheffernan/BambuFarmer/packages/printerstate"
	"github.com/jjheffernan/BambuFarmer/services/edge/app/farmmanager"
	"github.com/jjheffernan/BambuFarmer/services/edge/app/fixture"
)

// snState is a mutable overlay applied after StartPrint / Control (no LAN).
type snState struct {
	GcodeState string
	MCPercent  int
	Online     bool
}

// Plane cycles fixtures/fm device JSON and applies local control overlays.
type Plane struct {
	src *fixture.Source
	dir string
	log *slog.Logger

	mu        sync.Mutex
	overrides map[string]snState
	taskSeq   int
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
	return &Plane{
		src:       src,
		dir:       dir,
		log:       log,
		overrides: map[string]snState{},
	}, nil
}

func (p *Plane) Kind() farm.Kind { return farm.KindFixture }

func (p *Plane) Connect(context.Context) error { return nil }
func (p *Plane) Close() error                  { return nil }
func (p *Plane) Ready(context.Context) error   { return nil }

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
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, d := range devices {
		rep := mapDevice(d)
		if ov, ok := p.overrides[rep.ExternalSN]; ok {
			rep.GcodeState = ov.GcodeState
			rep.Online = ov.Online
			if ov.MCPercent > 0 {
				rep.MCPercent = ov.MCPercent
			}
		}
		out = append(out, rep)
	}
	if tags, err := fixture.LoadTags(p.dir); err == nil {
		farmmanager.ApplyDeviceTags(out, tags)
	}
	return out, nil
}

// ListTags returns fixture tags.json.
func (p *Plane) ListTags(context.Context) ([]farmmanager.Tag, error) {
	return fixture.LoadTags(p.dir)
}

func (p *Plane) Control(_ context.Context, sn string, opt api.ControlOpt) error {
	if sn == "" {
		return fmt.Errorf("Control: empty ExternalSN")
	}
	if !api.ValidControlOpt(opt) {
		return fmt.Errorf("Control: unsupported opt %q", opt)
	}
	p.log.Info("fixture control", "sn", sn, "opt", opt)
	p.mu.Lock()
	defer p.mu.Unlock()
	st := p.overrides[sn]
	st.Online = true
	switch opt {
	case api.ControlPause:
		st.GcodeState = string(printerstate.Pause)
	case api.ControlResume:
		st.GcodeState = string(printerstate.Running)
		if st.MCPercent == 0 {
			st.MCPercent = 10
		}
	case api.ControlStop, api.ControlBedClean:
		st.GcodeState = string(printerstate.Idle)
		st.MCPercent = 0
	}
	p.overrides[sn] = st
	return nil
}

func (p *Plane) Upload3MF(_ context.Context, name string, _ io.Reader) (farm.FileRef, error) {
	if name == "" {
		name = "fixture.gcode.3mf"
	}
	return farm.FileRef{ID: "fixture-file", Name: name}, nil
}

func (p *Plane) StartPrint(_ context.Context, req farm.PrintRequest) (farm.ExternalJobRef, error) {
	if req.ExternalSN == "" {
		return farm.ExternalJobRef{}, fmt.Errorf("StartPrint: empty ExternalSN")
	}
	if req.FileRef.ID == "" {
		return farm.ExternalJobRef{}, fmt.Errorf("StartPrint: FileRef.ID (f3mf_id) required")
	}
	p.log.Info("fixture start_print", "sn", req.ExternalSN, "job_id", req.JobID, "file_id", req.FileRef.ID)
	p.mu.Lock()
	p.taskSeq++
	taskID := fmt.Sprintf("fixture-task-%d", p.taskSeq)
	p.overrides[req.ExternalSN] = snState{
		GcodeState: string(printerstate.Running),
		MCPercent:  5,
		Online:     true,
	}
	p.mu.Unlock()
	return farm.ExternalJobRef{TaskID: taskID, Raw: req.JobID}, nil
}

func (p *Plane) LiveviewJPEG(_ context.Context, sn string) ([]byte, error) {
	if strings.TrimSpace(sn) == "" {
		return nil, fmt.Errorf("LiveviewJPEG: sn required")
	}
	// Minimal SOI/EOI JPEG so fleet UI and tests work offline.
	return []byte{0xff, 0xd8, 0xff, 0xd9}, nil
}

func mapDevice(d farmmanager.Device) farm.DeviceReport {
	rep := farm.DeviceReport{
		ExternalSN: d.DevID,
		Name:       d.Name,
		Model:      d.DevModel,
		DevIP:      d.DevIP,
		Tags:       append([]string(nil), d.Tags...),
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
