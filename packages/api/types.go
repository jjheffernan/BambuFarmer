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