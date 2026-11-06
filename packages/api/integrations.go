package api

// SpoolSummary is a normalized Spoolman spool for the BambuFarmer API.
type SpoolSummary struct {
	ID           string  `json:"id"`
	Name         string  `json:"name,omitempty"`
	Material     string  `json:"material,omitempty"`
	ColorHex     string  `json:"color_hex,omitempty"`
	Vendor       string  `json:"vendor,omitempty"`
	RemainingG   float64 `json:"remaining_g,omitempty"`
	UsedG        float64 `json:"used_g,omitempty"`
	Location     string  `json:"location,omitempty"`
	Archived     bool    `json:"archived,omitempty"`
	SpoolmanID   int     `json:"spoolman_id"`
}

// ModelSummary is a normalized Manyfold model for the BambuFarmer API.
type ModelSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}
