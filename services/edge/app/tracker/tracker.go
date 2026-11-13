package tracker

import (
	"sync"

	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/farm"
	"github.com/jjheffernan/BambuFarmer/packages/printerstate"
	"github.com/jjheffernan/BambuFarmer/services/edge/app/normalize"
)

// Tracker holds per-SN gcode_state machines and applies DeviceReport updates.
type Tracker struct {
	siteID string

	mu       sync.Mutex
	machines map[string]*printerstate.Machine
	pending  map[string]bool
	latest   map[string]api.PrinterSnapshot
}

// New creates a tracker for one site.
func New(siteID string) *Tracker {
	return &Tracker{
		siteID:   siteID,
		machines: make(map[string]*printerstate.Machine),
		pending:  make(map[string]bool),
		latest:   make(map[string]api.PrinterSnapshot),
	}
}

// SetPending marks whether SN has queued work (affects attention bucket).
func (t *Tracker) SetPending(sn string, pending bool) {
	t.mu.Lock()
	t.pending[sn] = pending
	t.mu.Unlock()
}

// Observe applies one device report through the state machine.
func (t *Tracker) Observe(d farm.DeviceReport) (api.PrinterSnapshot, printerstate.GcodeState, printerstate.GcodeState, bool, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	m, ok := t.machines[d.ExternalSN]
	if !ok {
		m = printerstate.NewMachine(d.ExternalSN)
		t.machines[d.ExternalSN] = m
	}
	from, to, forced, err := m.Observe(d.GcodeState)
	if err != nil {
		return api.PrinterSnapshot{}, from, to, forced, err
	}
	snap := normalize.Report(t.siteID, d, from, to, forced, t.pending[d.ExternalSN])
	t.latest[d.ExternalSN] = snap
	return snap, from, to, forced, nil
}

// ObserveAll processes a device list.
func (t *Tracker) ObserveAll(devices []farm.DeviceReport) ([]api.PrinterSnapshot, error) {
	out := make([]api.PrinterSnapshot, 0, len(devices))
	for _, d := range devices {
		snap, _, _, _, err := t.Observe(d)
		if err != nil {
			return out, err
		}
		out = append(out, snap)
	}
	return out, nil
}

// Summaries returns the latest fleet list (summary fields only).
func (t *Tracker) Summaries() []api.PrinterSummary {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]api.PrinterSummary, 0, len(t.latest))
	for _, s := range t.latest {
		out = append(out, s.PrinterSummary)
	}
	return out
}

// Snapshots returns the latest full printer snapshots (AMS, task, temps).
func (t *Tracker) Snapshots() []api.PrinterSnapshot {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]api.PrinterSnapshot, 0, len(t.latest))
	for _, s := range t.latest {
		out = append(out, s)
	}
	return out
}

// Count returns tracked printers.
func (t *Tracker) Count() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.latest)
}
