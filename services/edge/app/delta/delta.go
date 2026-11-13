package delta

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"

	"github.com/jjheffernan/BambuFarmer/packages/api"
)

// Filter emits only when a printer snapshot fingerprint changes.
type Filter struct {
	mu        sync.Mutex
	last      map[string]string // sn → full snapshot hash
	lastAMS   map[string]string // sn → AMS-only hash
	lastAlert map[string]string // sn → attention hash
}

// New returns an empty delta filter.
func New() *Filter {
	return &Filter{
		last:      make(map[string]string),
		lastAMS:   make(map[string]string),
		lastAlert: make(map[string]string),
	}
}

// Changed returns true if snap differs from the last seen fingerprint for its SN.
func (f *Filter) Changed(snap api.PrinterSnapshot) bool {
	h := fingerprint(snap)
	f.mu.Lock()
	defer f.mu.Unlock()
	prev, ok := f.last[snap.ExternalSN]
	if ok && prev == h {
		return false
	}
	f.last[snap.ExternalSN] = h
	return true
}

// AMSChanged returns true when AMS trays for SN change.
// The first empty AMS sighting is ignored so idle printers do not spam ams_changed.
func (f *Filter) AMSChanged(snap api.PrinterSnapshot) bool {
	h := amsFingerprint(snap.AMS)
	empty := amsFingerprint(nil)
	f.mu.Lock()
	defer f.mu.Unlock()
	prev, ok := f.lastAMS[snap.ExternalSN]
	if ok && prev == h {
		return false
	}
	f.lastAMS[snap.ExternalSN] = h
	if !ok && h == empty {
		return false
	}
	return true
}

// AlertRaised returns true when the printer newly needs attention
// (offline, PAUSE/FAILED, or non-empty HMS) or the attention fingerprint changes.
func (f *Filter) AlertRaised(snap api.PrinterSnapshot) bool {
	if !NeedsAttention(snap) {
		f.mu.Lock()
		delete(f.lastAlert, snap.ExternalSN)
		f.mu.Unlock()
		return false
	}
	h := attentionFingerprint(snap)
	f.mu.Lock()
	defer f.mu.Unlock()
	prev, ok := f.lastAlert[snap.ExternalSN]
	if ok && prev == h {
		return false
	}
	f.lastAlert[snap.ExternalSN] = h
	return true
}

// NeedsAttention reports offline, PAUSE/FAILED, or HMS codes.
func NeedsAttention(snap api.PrinterSnapshot) bool {
	if !snap.Online {
		return true
	}
	switch snap.GcodeState {
	case "PAUSE", "FAILED":
		return true
	}
	return len(snap.HMS) > 0
}

// Reset clears remembered fingerprints (e.g. after reconnect).
func (f *Filter) Reset() {
	f.mu.Lock()
	f.last = make(map[string]string)
	f.lastAMS = make(map[string]string)
	f.lastAlert = make(map[string]string)
	f.mu.Unlock()
}

func fingerprint(s api.PrinterSnapshot) string {
	// Stable subset: state + progress + online + temps rounded lightly via JSON.
	type fp struct {
		SN     string           `json:"sn"`
		Online bool             `json:"online"`
		State  string           `json:"state"`
		Pct    int              `json:"pct"`
		Remain int              `json:"remain"`
		Err    int              `json:"err"`
		HMS    []string         `json:"hms"`
		AMS    []api.PrinterAMS `json:"ams"`
		Task   *api.PrinterTask `json:"task"`
		Nozzle float64          `json:"nz"`
		Bed    float64          `json:"bed"`
		Layer  int              `json:"layer"`
		Forced bool             `json:"forced"`
	}
	b, _ := json.Marshal(fp{
		SN: s.ExternalSN, Online: s.Online, State: s.GcodeState,
		Pct: s.MCPercent, Remain: s.RemainingS, Err: s.PrintError, HMS: s.HMS, AMS: s.AMS, Task: s.Task,
		Nozzle: s.NozzleTemper, Bed: s.BedTemper, Layer: s.LayerNum, Forced: s.StateForced,
	})
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func amsFingerprint(ams []api.PrinterAMS) string {
	b, _ := json.Marshal(ams)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func attentionFingerprint(s api.PrinterSnapshot) string {
	type fp struct {
		Online bool     `json:"online"`
		State  string   `json:"state"`
		HMS    []string `json:"hms"`
	}
	b, _ := json.Marshal(fp{Online: s.Online, State: s.GcodeState, HMS: s.HMS})
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
