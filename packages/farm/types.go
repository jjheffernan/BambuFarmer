package farm

import "time"

// DeviceReport is a backend-neutral printer telemetry snapshot before edge state-machine enrichment.
type DeviceReport struct {
	ExternalSN         string
	Name               string
	Model              string
	DevIP              string
	Online             bool
	GcodeState         string
	MCPercent          int
	MCRemainingTime    int
	NozzleTemper       float64
	NozzleTargetTemper float64
	BedTemper          float64
	BedTargetTemper    float64
	ChamberTemper      float64
	LayerNum           int
	TotalLayerNum      int
	PrintError         int
	HMS                []string // formatted XXXX_XXXX_XXXX_XXXX codes from push_status hms[]
	AMS                []AMSUnit
	TaskID             string
	SubtaskName        string
	WifiSignal         string
	NozzleDiameter     string
	Tags               []string
	ObservedAt         time.Time
}

// PrintRequest starts a print on one or more printers via the farm control plane.
type PrintRequest struct {
	JobID      string  // BambuFarmer job id (correlation)
	ExternalSN string  // target printer
	FileRef    FileRef // previously uploaded plate, or empty if Name+Body used
	Name       string  // filename when uploading inline
	PlateIndex int     // 0-based when relevant
	Copies     int
}

// FileRef is an opaque farm-side file id / path.
type FileRef struct {
	ID   string `json:"id,omitempty"`
	Path string `json:"path,omitempty"`
	Name string `json:"name,omitempty"`
}

// ExternalJobRef is the farm/task id returned after start.
type ExternalJobRef struct {
	TaskID    string `json:"task_id,omitempty"`
	SubtaskID string `json:"subtask_id,omitempty"`
	Raw       string `json:"raw,omitempty"`
}
