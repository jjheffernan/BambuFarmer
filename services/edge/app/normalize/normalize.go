package normalize

import (
	"time"

	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/farm"
	"github.com/jjheffernan/BambuFarmer/packages/printerstate"
)

// Report converts a farm.DeviceReport + state-machine observation into PrinterSnapshot.
func Report(siteID string, d farm.DeviceReport, from, to printerstate.GcodeState, forced bool, pending bool) api.PrinterSnapshot {
	_ = from // retained for event payloads / callers
	online := d.Online
	state := string(to)
	if to == "" {
		state = string(printerstate.Idle)
	}
	name := d.Name
	if name == "" {
		name = d.ExternalSN
	}
	sum := api.PrinterSummary{
		ID:         d.ExternalSN,
		SiteID:     siteID,
		ExternalSN: d.ExternalSN,
		Name:       name,
		Model:      d.Model,
		Online:     online,
		GcodeState: state,
		MCPercent:  d.MCPercent,
		RemainingS: d.MCRemainingTime,
		Bucket:     string(printerstate.BucketFor(online, to, pending)),
		Tags:       append([]string(nil), d.Tags...),
	}
	obs := d.ObservedAt
	if obs.IsZero() {
		obs = time.Now().UTC()
	}
	return api.PrinterSnapshot{
		PrinterSummary: sum,
		DevIP:          d.DevIP,
		NozzleTemper:   d.NozzleTemper,
		NozzleTarget:   d.NozzleTargetTemper,
		BedTemper:      d.BedTemper,
		BedTarget:      d.BedTargetTemper,
		ChamberTemper:  d.ChamberTemper,
		LayerNum:       d.LayerNum,
		TotalLayerNum:  d.TotalLayerNum,
		PrintError:     d.PrintError,
		HMS:            append([]string(nil), d.HMS...),
		WifiSignal:     d.WifiSignal,
		NozzleDiameter: d.NozzleDiameter,
		StateForced:    forced,
		ObservedAt:     obs,
	}
}
