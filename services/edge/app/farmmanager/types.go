package farmmanager

import (
	"encoding/json"

	"github.com/jjheffernan/BambuFarmer/packages/api"
)

// Captain is the FM GET /captain payload (fields we care about).
type Captain struct {
	Name       string `json:"name"`
	ServerID   string `json:"server_id"`
	Release    string `json:"release"`
	Version    string `json:"version"`
	PowerLimit int    `json:"power_limit"`
	Alarming   int    `json:"alarming"`
}

func (c Captain) ToAPI() api.CaptainInfo {
	return api.CaptainInfo{
		Name:       c.Name,
		ServerID:   c.ServerID,
		Release:    c.Release,
		Version:    c.Version,
		PowerLimit: c.PowerLimit,
		Alarming:   c.Alarming,
	}
}

// DevicesResponse is GET /devices|/devices2.
type DevicesResponse struct {
	Devices []Device `json:"devices"`
}

// Device is one FM device entry.
type Device struct {
	DevID        string        `json:"dev_id"`
	DevModel     string        `json:"dev_model"`
	DevName      string        `json:"dev_name"`
	DevIP        string        `json:"dev_ip"`
	Name         string        `json:"name"`
	Tags         []string      `json:"tags,omitempty"`
	Online       *bool         `json:"online,omitempty"` // some builds omit; infer from report_status
	ReportStatus *ReportStatus `json:"report_status"`
}

// ReportStatus mirrors push_status / report_status telemetry subset.
type ReportStatus struct {
	GcodeState         string  `json:"gcode_state"`
	NozzleTemper       float64 `json:"nozzle_temper"`
	NozzleTargetTemper float64 `json:"nozzle_target_temper"`
	BedTemper          float64 `json:"bed_temper"`
	BedTargetTemper    float64 `json:"bed_target_temper"`
	ChamberTemper      float64 `json:"chamber_temper"`
	MCPercent          int     `json:"mc_percent"`
	MCRemainingTime    int     `json:"mc_remaining_time"`
	WifiSignal         string  `json:"wifi_signal"`
	PrintError         int             `json:"print_error"`
	HMS                json.RawMessage `json:"hms,omitempty"`
	AMS                json.RawMessage `json:"ams,omitempty"`
	TaskID             string          `json:"task_id,omitempty"`
	SubtaskName        string          `json:"subtask_name,omitempty"`
	LayerNum           int             `json:"layer_num"`
	TotalLayerNum      int             `json:"total_layer_num"`
	NozzleDiameter     string          `json:"nozzle_diameter"`
	NozzleType         string          `json:"nozzle_type"`
}

// TicketResponse is POST /login/local/tickets.
type TicketResponse struct {
	Message  string `json:"message"`
	Code     int    `json:"code"`
	TicketID int64  `json:"ticket_id"`
	Ticket   string `json:"ticket"`
}

// TicketPayload is the base64-decoded ticket JSON (challenge + salt).
type TicketPayload struct {
	Challenge string `json:"challenge"`
	Salt      string `json:"salt"`
}

// TokenResponse is POST /login/local/tokens (shape varies; accept common keys).
type TokenResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
	Token   string `json:"token"`
	Access  string `json:"access_token"`
}

func (t TokenResponse) JWT() string {
	if t.Token != "" {
		return t.Token
	}
	return t.Access
}

// OptBody is POST /device/{sn}/opt.
type OptBody struct {
	Opt string `json:"opt"`
}
