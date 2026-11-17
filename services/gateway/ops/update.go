// Package ops handles host-side update requests from the Admin UI.
// Containers cannot rebuild themselves; the UI drops a request file that
// `./scripts/bf update --watch` (or a cron) applies on the host.
package ops

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	requestFile = "update.requested"
	statusFile  = "update.status.json"
)

// Status is returned by GET /api/v1/admin/system.
type Status struct {
	Pending          bool        `json:"pending"`
	RequestedAt      string      `json:"requested_at,omitempty"`
	RequestedBy      string      `json:"requested_by,omitempty"`
	RequestedTargets []string    `json:"requested_targets,omitempty"`
	LastState        string      `json:"last_state,omitempty"` // idle|running|ok|error
	LastMessage      string      `json:"last_message,omitempty"`
	LastAt           string      `json:"last_at,omitempty"`
	LastTargets      []string    `json:"last_targets,omitempty"`
	UpdateCommand    string      `json:"update_command"`
	WatchCommand     string      `json:"watch_command"`
	OpsDir           string      `json:"ops_dir"`
	Note             string      `json:"note"`
	Components       []Component `json:"components"`
}

// Result is written by the host updater after a run.
type Result struct {
	State      string   `json:"state"`
	Message    string   `json:"message,omitempty"`
	At         string   `json:"at"`
	Version    string   `json:"version,omitempty"`
	Components []string `json:"components,omitempty"`
}

// Request is the JSON body of update.requested / POST /admin/system/update.
type Request struct {
	At         string   `json:"at"`
	By         string   `json:"by"`
	Components []string `json:"components,omitempty"` // empty / ["all"] = full stack
	Pull       bool     `json:"pull,omitempty"`
	NoGit      bool     `json:"no_git,omitempty"`
}

// Dir returns OPS_DIR (default /ops).
func Dir() string {
	if d := os.Getenv("OPS_DIR"); d != "" {
		return d
	}
	return "/ops"
}

// GetStatus reads pending request + last result + live component probes.
func GetStatus(ctx context.Context) Status {
	dir := Dir()
	st := Status{
		OpsDir:        dir,
		UpdateCommand: "./scripts/bf update",
		WatchCommand:  "./scripts/bf update --watch",
		Note:          "Containers cannot rebuild themselves. Admin requests write data/ops/update.requested; run ./scripts/bf update --watch on the host. Target individual services or the full stack.",
		LastState:     "idle",
		Components:    ProbeComponents(ctx),
	}
	reqPath := filepath.Join(dir, requestFile)
	if b, err := os.ReadFile(reqPath); err == nil {
		st.Pending = true
		var req Request
		if json.Unmarshal(b, &req) == nil {
			st.RequestedAt = req.At
			st.RequestedBy = req.By
			st.RequestedTargets = req.Components
			if len(st.RequestedTargets) == 0 {
				st.RequestedTargets = []string{"all"}
			}
		}
	}
	if b, err := os.ReadFile(filepath.Join(dir, statusFile)); err == nil {
		var r Result
		if json.Unmarshal(b, &r) == nil {
			st.LastState = r.State
			st.LastMessage = r.Message
			st.LastAt = r.At
			st.LastTargets = r.Components
		}
	}
	return st
}

// RequestUpdate writes the host-side request file (admin-triggered).
func RequestUpdate(by string, components []string, pull, noGit bool) (Status, error) {
	dir := Dir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return Status{}, err
	}
	if len(components) == 1 && strings.EqualFold(strings.TrimSpace(components[0]), "all") {
		components = nil
	}
	norm := NormalizeComponents(components)
	if len(components) > 0 && len(norm) == 0 {
		return Status{}, fmt.Errorf("no valid components in %v (allowed: %s)", components, strings.Join(componentIDList(), ", "))
	}
	req := Request{
		At:         time.Now().UTC().Format(time.RFC3339),
		By:         by,
		Components: norm, // empty = full stack; watcher applies these compose services
		Pull:       pull,
		NoGit:      noGit,
	}
	meta, err := json.Marshal(req)
	if err != nil {
		return Status{}, err
	}
	path := filepath.Join(dir, requestFile)
	if err := os.WriteFile(path, meta, 0o644); err != nil {
		return Status{}, err
	}
	return GetStatus(context.Background()), nil
}

func componentIDList() []string {
	var ids []string
	for _, c := range Catalog() {
		ids = append(ids, c.ID)
	}
	return ids
}
