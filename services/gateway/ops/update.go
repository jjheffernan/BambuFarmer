// Package ops handles host-side update requests from the Admin UI.
// Containers cannot rebuild themselves; the UI drops a request file that
// `./scripts/bf update --watch` (or a cron) applies on the host.
package ops

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const (
	requestFile = "update.requested"
	statusFile  = "update.status.json"
)

// Status is returned by GET /api/v1/admin/system.
type Status struct {
	Pending       bool   `json:"pending"`
	RequestedAt   string `json:"requested_at,omitempty"`
	RequestedBy   string `json:"requested_by,omitempty"`
	LastState     string `json:"last_state,omitempty"` // idle|running|ok|error
	LastMessage   string `json:"last_message,omitempty"`
	LastAt        string `json:"last_at,omitempty"`
	UpdateCommand string `json:"update_command"`
	WatchCommand  string `json:"watch_command"`
	OpsDir        string `json:"ops_dir"`
	Note          string `json:"note"`
}

// Result is written by the host updater after a run.
type Result struct {
	State     string `json:"state"`
	Message   string `json:"message,omitempty"`
	At        string `json:"at"`
	Version   string `json:"version,omitempty"`
}

// Dir returns OPS_DIR (default /ops).
func Dir() string {
	if d := os.Getenv("OPS_DIR"); d != "" {
		return d
	}
	return "/ops"
}

// GetStatus reads pending request + last result.
func GetStatus() Status {
	dir := Dir()
	st := Status{
		OpsDir:        dir,
		UpdateCommand: "./scripts/bf update",
		WatchCommand:  "./scripts/bf update --watch",
		Note:          "After a Farm Manager upgrade, pull a matching BambuFarmer release and run update (or use Admin → Request update with a host watcher).",
		LastState:     "idle",
	}
	reqPath := filepath.Join(dir, requestFile)
	if b, err := os.ReadFile(reqPath); err == nil {
		st.Pending = true
		var meta map[string]string
		if json.Unmarshal(b, &meta) == nil {
			st.RequestedAt = meta["at"]
			st.RequestedBy = meta["by"]
		}
	}
	if b, err := os.ReadFile(filepath.Join(dir, statusFile)); err == nil {
		var r Result
		if json.Unmarshal(b, &r) == nil {
			st.LastState = r.State
			st.LastMessage = r.Message
			st.LastAt = r.At
		}
	}
	return st
}

// RequestUpdate writes the host-side request file (admin-triggered).
func RequestUpdate(by string) (Status, error) {
	dir := Dir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return Status{}, err
	}
	meta, _ := json.Marshal(map[string]string{
		"at": time.Now().UTC().Format(time.RFC3339),
		"by": by,
	})
	path := filepath.Join(dir, requestFile)
	if err := os.WriteFile(path, meta, 0o644); err != nil {
		return Status{}, err
	}
	return GetStatus(), nil
}
