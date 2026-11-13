package slice

import (
	"context"

	"github.com/jjheffernan/BambuFarmer/packages/events"
)

// Bus is the Redis Streams subset slicer needs for slice lifecycle events.
type Bus interface {
	Publish(ctx context.Context, stream string, env events.Envelope) (string, error)
}

// Emit publishes slice.queued then slice.succeeded or slice.failed.
// A nil bus is a no-op. Publish errors are ignored so HTTP still returns the Result.
func Emit(ctx context.Context, bus Bus, siteID string, req Request, res Result) {
	if bus == nil {
		return
	}
	if siteID == "" {
		siteID = "default"
	}
	payload := map[string]any{
		"name":        res.Name,
		"input_path":  req.InputPath,
		"output_path": res.OutputPath,
		"sliced":      res.Sliced,
	}
	if res.Reason != "" {
		payload["reason"] = res.Reason
	}
	if res.ExitCode != 0 {
		payload["exit_code"] = res.ExitCode
	}
	_, _ = bus.Publish(ctx, events.StreamSlice, events.NewEnvelope(events.TypeSliceQueued, "slicer", siteID, payload))
	typ := events.TypeSliceFailed
	if res.Sliced {
		typ = events.TypeSliceSucceeded
	}
	_, _ = bus.Publish(ctx, events.StreamSlice, events.NewEnvelope(typ, "slicer", siteID, payload))
}
