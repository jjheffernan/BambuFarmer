// Package mqtt turns a farm report payload into a farm.DeviceReport.
// Subscribe dials only when FARM_MQTT_URL and cert files are both set. Polling stays on either way.
package mqtt

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jjheffernan/BambuFarmer/packages/farm"
)

// SerialFromTopic reads device/<sn>/report.
func SerialFromTopic(topic string) (string, bool) {
	parts := strings.Split(strings.Trim(topic, "/"), "/")
	if len(parts) != 3 || parts[0] != "device" || parts[2] != "report" || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}

// Apply folds one report payload onto base.
// msg 0 replaces telemetry. msg 1 writes only fields present in the payload.
func Apply(base farm.DeviceReport, topic string, payload []byte) (farm.DeviceReport, error) {
	sn, ok := SerialFromTopic(topic)
	if !ok {
		return base, fmt.Errorf("mqtt: topic %q", topic)
	}
	msg, fields, err := decode(payload)
	if err != nil {
		return base, err
	}
	out := base
	if msg == 0 {
		out = farm.DeviceReport{
			ExternalSN: sn,
			Name:       base.Name,
			Model:      base.Model,
			DevIP:      base.DevIP,
		}
	}
	out.ExternalSN = sn
	out.Online = true
	out.ObservedAt = time.Now().UTC()
	fill(&out, fields)
	return out, nil
}

func decode(payload []byte) (int, map[string]json.RawMessage, error) {
	var env struct {
		Print json.RawMessage `json:"print"`
	}
	if err := json.Unmarshal(payload, &env); err != nil {
		return 0, nil, err
	}
	if len(env.Print) == 0 {
		return 0, nil, fmt.Errorf("mqtt: missing print")
	}
	var head struct {
		Command string `json:"command"`
		Msg     *int   `json:"msg"`
	}
	if err := json.Unmarshal(env.Print, &head); err != nil {
		return 0, nil, err
	}
	if head.Command != "" && head.Command != "push_status" {
		return 0, nil, fmt.Errorf("mqtt: skip %s", head.Command)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(env.Print, &fields); err != nil {
		return 0, nil, err
	}
	msg := 0
	if head.Msg != nil {
		msg = *head.Msg
	}
	return msg, fields, nil
}

func fill(rep *farm.DeviceReport, fields map[string]json.RawMessage) {
	take(fields, "gcode_state", &rep.GcodeState)
	take(fields, "mc_percent", &rep.MCPercent)
	take(fields, "mc_remaining_time", &rep.MCRemainingTime)
	take(fields, "nozzle_temper", &rep.NozzleTemper)
	take(fields, "nozzle_target_temper", &rep.NozzleTargetTemper)
	take(fields, "bed_temper", &rep.BedTemper)
	take(fields, "bed_target_temper", &rep.BedTargetTemper)
	take(fields, "chamber_temper", &rep.ChamberTemper)
	take(fields, "layer_num", &rep.LayerNum)
	take(fields, "total_layer_num", &rep.TotalLayerNum)
	take(fields, "print_error", &rep.PrintError)
	take(fields, "wifi_signal", &rep.WifiSignal)
	take(fields, "nozzle_diameter", &rep.NozzleDiameter)
}

func take(fields map[string]json.RawMessage, key string, dest any) {
	raw, ok := fields[key]
	if !ok {
		return
	}
	_ = json.Unmarshal(raw, dest)
}
