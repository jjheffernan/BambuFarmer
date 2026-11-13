// Package direct is an in-process printer adapter. It does not open a network connection.
package direct

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/farm"
)

// Plane implements farm.ControlPlane for printers listed in DIRECT_PRINTERS.
// Format: "sn=name,sn2=name2". An empty list is ready and has no devices.
type Plane struct {
	mu      sync.Mutex
	devices map[string]*farm.DeviceReport
	order   []string
	files   map[string]farm.FileRef
	next    int
}

// ParsePrinters reads DIRECT_PRINTERS. Blank pairs are skipped.
func ParsePrinters(raw string) []farm.DeviceReport {
	var out []farm.DeviceReport
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		sn, name, ok := strings.Cut(part, "=")
		sn, name = strings.TrimSpace(sn), strings.TrimSpace(name)
		if sn == "" {
			continue
		}
		if !ok || name == "" {
			name = sn
		}
		out = append(out, farm.DeviceReport{
			ExternalSN: sn,
			Name:       name,
			Online:     true,
			GcodeState: "IDLE",
		})
	}
	return out
}

// New builds a plane from a DIRECT_PRINTERS value.
func New(raw string) *Plane {
	p := &Plane{
		devices: map[string]*farm.DeviceReport{},
		files:   map[string]farm.FileRef{},
	}
	for _, d := range ParsePrinters(raw) {
		cp := d
		p.devices[d.ExternalSN] = &cp
		p.order = append(p.order, d.ExternalSN)
	}
	return p
}

func (p *Plane) Kind() farm.Kind               { return farm.KindDirect }
func (p *Plane) Connect(context.Context) error { return nil }
func (p *Plane) Close() error                  { return nil }
func (p *Plane) Ready(context.Context) error   { return nil }

func (p *Plane) Captain(context.Context) (api.CaptainInfo, error) {
	return api.CaptainInfo{ServerID: "direct", Release: "local"}, nil
}

func (p *Plane) ListDevices(context.Context) ([]farm.DeviceReport, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]farm.DeviceReport, 0, len(p.order))
	for _, sn := range p.order {
		out = append(out, *p.devices[sn])
	}
	return out, nil
}

func (p *Plane) Control(_ context.Context, sn string, opt api.ControlOpt) error {
	if !api.ValidControlOpt(opt) {
		return fmt.Errorf("direct: unknown action %q", opt)
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	d := p.devices[sn]
	if d == nil {
		return fmt.Errorf("direct: unknown printer %s", sn)
	}
	switch opt {
	case api.ControlPause:
		d.GcodeState = "PAUSE"
	case api.ControlResume:
		d.GcodeState = "RUNNING"
	case api.ControlStop, api.ControlBedClean:
		d.GcodeState = "IDLE"
		d.MCPercent = 0
	}
	return nil
}

func (p *Plane) Upload3MF(_ context.Context, name string, r io.Reader) (farm.FileRef, error) {
	if _, err := io.Copy(io.Discard, r); err != nil {
		return farm.FileRef{}, err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.next++
	ref := farm.FileRef{ID: fmt.Sprintf("direct-%d", p.next), Name: strings.TrimSpace(name)}
	p.files[ref.ID] = ref
	return ref, nil
}

func (p *Plane) StartPrint(_ context.Context, req farm.PrintRequest) (farm.ExternalJobRef, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	d := p.devices[req.ExternalSN]
	if d == nil {
		return farm.ExternalJobRef{}, fmt.Errorf("direct: unknown printer %s", req.ExternalSN)
	}
	if req.FileRef.ID != "" {
		if _, ok := p.files[req.FileRef.ID]; !ok {
			return farm.ExternalJobRef{}, fmt.Errorf("direct: unknown file %s", req.FileRef.ID)
		}
	}
	d.GcodeState = "RUNNING"
	d.MCPercent = 0
	return farm.ExternalJobRef{TaskID: "direct-" + req.JobID}, nil
}

func (p *Plane) LiveviewJPEG(context.Context, string) ([]byte, error) {
	return nil, farm.ErrNotImplemented
}

var _ farm.ControlPlane = (*Plane)(nil)
