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
	mu    sync.Mutex
	last  map[string]string // sn → hash
}

// New returns an empty delta filter.
func New() *Filter {
	return &Filter{last: make(map[string]string)}
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

// Reset clears remembered fingerprints (e.g. after reconnect).
func (f *Filter) Reset() {
	f.mu.Lock()
	f.last = make(map[string]string)
	f.mu.Unlock()
}

func fingerprint(s api.PrinterSnapshot) string {
	// Stable subset: state + progress + online + temps rounded lightly via JSON.
	type fp struct {
		SN     string   `json:"sn"`
		Online bool     `json:"online"`
		State  string   `json:"state"`
		Pct    int      `json:"pct"`
		Remain int      `json:"remain"`
		Err    int      `json:"err"`
		HMS    []string `json:"hms"`
		Nozzle float64  `json:"nz"`
		Bed    float64  `json:"bed"`
		Layer  int      `json:"layer"`
		Forced bool     `json:"forced"`
	}
	b, _ := json.Marshal(fp{
		SN: s.ExternalSN, Online: s.Online, State: s.GcodeState,
		Pct: s.MCPercent, Remain: s.RemainingS, Err: s.PrintError, HMS: s.HMS,
		Nozzle: s.NozzleTemper, Bed: s.BedTemper, Layer: s.LayerNum, Forced: s.StateForced,
	})
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
