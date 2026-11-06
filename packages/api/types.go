package api

import "time"

// HealthResponse is returned by /healthz and /readyz.
type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

// FilamentPreset is a per-install filament definition (see DESIGN_SPEC §8.4).
type FilamentPreset struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	Version            int      `json:"version"`
	Material           string   `json:"material"`
	Vendor             string   `json:"vendor,omitempty"`
	ProductLine        string   `json:"product_line,omitempty"`
	ColorHex           string   `json:"color_hex,omitempty"`
	DensityGCm3        float64  `json:"density_g_cm3,omitempty"`
	DiameterMM         float64  `json:"diameter_mm,omitempty"`
	NozzleTempMin      int      `json:"nozzle_temp_min,omitempty"`
	NozzleTempMax      int      `json:"nozzle_temp_max,omitempty"`
	BedTemp            int      `json:"bed_temp,omitempty"`
	TrayInfoIdx        string   `json:"tray_info_idx,omitempty"`
	PricePerKg         float64  `json:"price_per_kg,omitempty"`
	SpoolmanFilamentID *int64   `json:"spoolman_filament_id,omitempty"`
	Notes              string   `json:"notes,omitempty"`
	Active             bool     `json:"active"`
	UpdatedAt          time.Time `json:"updated_at,omitempty"`
}

// PrintPreset is a per-install print/process definition.
type PrintPreset struct {
	ID               string         `json:"id"`
	Name             string         `json:"name"`
	Version          int            `json:"version"`
	CompatibleModels []string       `json:"compatible_models,omitempty"`
	NozzleDiameterMM float64        `json:"nozzle_diameter_mm,omitempty"`
	LayerHeightMM    float64        `json:"layer_height_mm,omitempty"`
	InfillPct        float64        `json:"infill_pct,omitempty"`
	ProcessJSON      map[string]any `json:"process_json,omitempty"`
	PrinterJSON      map[string]any `json:"printer_json,omitempty"`
	DefaultForModel  string         `json:"default_for_model,omitempty"`
	EstSpeedFactor   float64        `json:"est_speed_factor,omitempty"`
	Active           bool           `json:"active"`
	UpdatedAt        time.Time      `json:"updated_at,omitempty"`
}

// PresetBundle is the import/export JSON shape.
type PresetBundle struct {
	FilamentPresets []FilamentPreset `json:"filament_presets"`
	PrintPresets    []PrintPreset    `json:"print_presets"`
}

// PrinterSummary is a normalized fleet list item.
type PrinterSummary struct {
	ID         string `json:"id"`
	SiteID     string `json:"site_id"`
	ExternalSN string `json:"external_sn"`
	Name       string `json:"name"`
	Model      string `json:"model"`
	Online     bool   `json:"online"`
	GcodeState string `json:"gcode_state"`
	MCPercent  int    `json:"mc_percent"`
	RemainingS int    `json:"mc_remaining_time"`
	Bucket     string `json:"bucket,omitempty"` // attention bucket (FM nav mirror)
}

// PrinterSnapshot is the edge-normalized device view (FM report_status subset).
type PrinterSnapshot struct {
	PrinterSummary
	DevIP           string  `json:"dev_ip,omitempty"`
	NozzleTemper    float64 `json:"nozzle_temper,omitempty"`
	NozzleTarget    float64 `json:"nozzle_target_temper,omitempty"`
	BedTemper       float64 `json:"bed_temper,omitempty"`
	BedTarget       float64 `json:"bed_target_temper,omitempty"`
	ChamberTemper   float64 `json:"chamber_temper,omitempty"`
	LayerNum        int     `json:"layer_num,omitempty"`
	TotalLayerNum   int     `json:"total_layer_num,omitempty"`
	PrintError      int     `json:"print_error,omitempty"`
	WifiSignal      string  `json:"wifi_signal,omitempty"`
	NozzleDiameter  string  `json:"nozzle_diameter,omitempty"`
	StateForced     bool    `json:"state_forced,omitempty"` // true if FM jumped an illegal soft transition
	ObservedAt      time.Time `json:"observed_at,omitempty"`
}

// ControlOpt is a Farm Manager /device/{sn}/opt value.
type ControlOpt string

const (
	ControlPause    ControlOpt = "pause"
	ControlResume   ControlOpt = "resume"
	ControlStop     ControlOpt = "stop"
	ControlBedClean ControlOpt = "bed_clean"
)

// CaptainInfo is a subset of GET /captain.
type CaptainInfo struct {
	Name      string `json:"name"`
	ServerID  string `json:"server_id"`
	Release   string `json:"release"`
	Version   string `json:"version,omitempty"`
	PowerLimit int   `json:"power_limit,omitempty"`
	Alarming  int    `json:"alarming,omitempty"`
}

// EdgeStatus is returned by GET /api/v1/edge/status.
type EdgeStatus struct {
	SiteID         string            `json:"site_id"`
	FarmMode       string            `json:"farm_mode"`
	FarmManagerURL string            `json:"farm_manager_url"`
	Transport      string            `json:"transport"`
	Connected      bool              `json:"connected"`
	ServerID       string            `json:"server_id,omitempty"`
	Release        string            `json:"release,omitempty"`
	DeviceCount    int               `json:"device_count"`
	LastPollAt     *time.Time        `json:"last_poll_at,omitempty"`
	LastError      string            `json:"last_error,omitempty"`
	Note           string            `json:"note,omitempty"`
	Devices        []PrinterSummary  `json:"devices,omitempty"`
}

// RateCard holds org/site costing defaults ($/hr, $/kg, kg/hr).
type RateCard struct {
	Currency            string  `json:"currency"`
	MachineRatePerHour  float64 `json:"machine_rate_per_hour"`  // $/hr
	FilamentPricePerKg  float64 `json:"filament_price_per_kg"`  // $/kg
	FilamentKgPerHour   float64 `json:"filament_kg_per_hour"`   // kg/hr
	LaborFlat           float64 `json:"labor_flat,omitempty"`
	FailureSurchargePct float64 `json:"failure_surcharge_pct,omitempty"`
}

// CreateJobRequest is POST /api/v1/jobs body.
type CreateJobRequest struct {
	ID         string `json:"id,omitempty"`
	Name       string `json:"name,omitempty"`
	ExternalSN string `json:"external_sn,omitempty"`
	ArtifactID string `json:"artifact_id,omitempty"`
	Copies     int    `json:"copies,omitempty"`
	EstSeconds int    `json:"est_seconds,omitempty"`
	EstGrams   float64 `json:"est_grams,omitempty"`
}

// InstallState is the first-run wizard singleton.
type InstallState struct {
	Completed             bool       `json:"completed"`
	CompletedAt           *time.Time `json:"completed_at,omitempty"`
	Version               string     `json:"version,omitempty"`
	Step                  string     `json:"step"`
	SiteName              string     `json:"site_name"`
	Timezone              string     `json:"timezone"`
	AuthMode              string     `json:"auth_mode"`
	FarmManagerURL        string     `json:"farm_manager_url,omitempty"`
	FarmManagerTransport  string     `json:"farm_manager_transport,omitempty"`
	SpoolmanURL           string     `json:"spoolman_url,omitempty"`
	ManyfoldURL           string     `json:"manyfold_url,omitempty"`
	AdminUsername         string     `json:"admin_username,omitempty"`
	Notes                 string     `json:"notes,omitempty"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

// SetupRequest is POST /api/v1/setup (draft or complete).
type SetupRequest struct {
	Step                  string     `json:"step,omitempty"`
	SiteName              string     `json:"site_name,omitempty"`
	Timezone              string     `json:"timezone,omitempty"`
	AuthMode              string     `json:"auth_mode,omitempty"`
	FarmManagerURL        string     `json:"farm_manager_url,omitempty"`
	FarmManagerTransport  string     `json:"farm_manager_transport,omitempty"`
	SpoolmanURL           string     `json:"spoolman_url,omitempty"`
	ManyfoldURL           string     `json:"manyfold_url,omitempty"`
	AdminUsername         string     `json:"admin_username,omitempty"`
	Notes                 string     `json:"notes,omitempty"`
	RateCard              *RateCard  `json:"rate_card,omitempty"`
	ImportExamplePresets  bool       `json:"import_example_presets,omitempty"`
	Complete              bool       `json:"complete,omitempty"`
}

// Job is a print job record.
type Job struct {
	ID           string     `json:"id"`
	SiteID       string     `json:"site_id"`
	Name         string     `json:"name"`
	State        string     `json:"state"`
	ExternalSN   string     `json:"external_sn,omitempty"`
	ArtifactID   string     `json:"artifact_id,omitempty"`
	Copies       int        `json:"copies"`
	EstSeconds   int        `json:"est_seconds,omitempty"`
	EstGrams     float64    `json:"est_grams,omitempty"`
	CostEstimate float64    `json:"cost_estimate"`
	CostActual   float64    `json:"cost_actual"`
	Currency     string     `json:"currency"`
	Error        string     `json:"error,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	FinishedAt   *time.Time `json:"finished_at,omitempty"`
}