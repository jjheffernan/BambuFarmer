package events

import "time"

// Envelope is the common Redis Streams / SSE payload wrapper.
type Envelope struct {
	ID            string         `json:"id"`
	Type          string         `json:"type"`
	TS            time.Time      `json:"ts"`
	SiteID        string         `json:"site_id,omitempty"`
	Source        string         `json:"source"`
	CorrelationID string         `json:"correlation_id,omitempty"`
	Payload       map[string]any `json:"payload"`
}

// Stream names (Redis Streams keys).
const (
	StreamPrinter  = "bf.events.printer"
	StreamJob      = "bf.events.job"
	StreamFilament = "bf.events.filament"
	StreamAlert    = "bf.events.alert"
	StreamSlice    = "bf.events.slice"
)

// CommandStream returns the per-site edge command stream key.
func CommandStream(siteID string) string {
	return "bf.commands.edge." + siteID
}

// Common event type constants.
const (
	TypePrinterOnline       = "printer.online"
	TypePrinterOffline      = "printer.offline"
	TypePrinterStateChanged = "printer.state_changed"
	TypePrinterAMSChanged   = "printer.ams_changed"
	TypeJobQueued           = "job.queued"
	TypeJobDispatched       = "job.dispatched"
	TypeJobProgress         = "job.progress"
	TypeJobFinished         = "job.finished"
	TypeJobFailed           = "job.failed"
	TypeJobCancelled        = "job.cancelled"
	TypeFilamentConsumed    = "filament.consumed"
	TypeAlertRaised         = "alert.raised"
)