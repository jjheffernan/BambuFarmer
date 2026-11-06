// Package farm defines the Bambu / Farm Manager integration abstraction.
// Edge is the only process that speaks a concrete backend (fm_native, CDP, fixture, later Fleet Hub).
package farm

import (
	"context"
	"errors"
	"io"

	"github.com/jjheffernan/BambuFarmer/packages/api"
)

// Kind identifies a ControlPlane backend.
type Kind string

const (
	KindNative  Kind = "fm_native" // Farm Manager Client–shaped REST
	KindCDP     Kind = "cdp"      // FM Client DevTools escape hatch
	KindFixture Kind = "fixture"  // redacted JSON for lab / CI
	KindHub     Kind = "fleet_hub" // future
	KindNone    Kind = "none"
)

// Common errors for unimplemented / unavailable operations.
var (
	ErrNotConfigured   = errors.New("farm: control plane not configured")
	ErrNotImplemented  = errors.New("farm: operation not implemented for this backend")
	ErrNotConnected    = errors.New("farm: not connected")
	ErrUnsupportedKind = errors.New("farm: unsupported backend kind")
)

// ControlPlane is the Bambu Lab farm control abstraction (DESIGN_SPEC §5.3).
// Implementations must not leak backend-specific HTTP types to callers.
type ControlPlane interface {
	Kind() Kind
	// Connect establishes a session (login / CDP attach). Idempotent.
	Connect(ctx context.Context) error
	// Close releases resources.
	Close() error
	// Ready reports whether the plane can serve ListDevices / Control.
	Ready(ctx context.Context) error
	// Captain returns server identity when available (FM /captain).
	Captain(ctx context.Context) (api.CaptainInfo, error)
	// ListDevices returns a full-ish fleet snapshot (REST poll or MQTT cache).
	ListDevices(ctx context.Context) ([]DeviceReport, error)
	// Control issues pause|resume|stop|bed_clean for one printer.
	Control(ctx context.Context, sn string, opt api.ControlOpt) error
	// Upload3MF stages a plate on the farm control plane.
	Upload3MF(ctx context.Context, name string, r io.Reader) (FileRef, error)
	// StartPrint launches a managed print (upload may already be done).
	StartPrint(ctx context.Context, req PrintRequest) (ExternalJobRef, error)
	// LiveviewJPEG fetches a camera still when supported.
	LiveviewJPEG(ctx context.Context, sn string) ([]byte, error)
}

// EventSource is an optional push telemetry source (MQTT preferred).
// When unavailable, the edge agent falls back to polling ControlPlane.ListDevices.
type EventSource interface {
	// Subscribe yields device reports until ctx cancel. Channel is closed on exit.
	Subscribe(ctx context.Context) (<-chan DeviceReport, error)
}
