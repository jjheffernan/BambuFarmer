package db

import (
	"strings"

	"github.com/jjheffernan/BambuFarmer/packages/api"
)

// PrinterListFilter narrows ListPrinterSnapshots. Empty fields mean no constraint.
type PrinterListFilter struct {
	Tag    string // exact tag match (case-insensitive)
	State  string // gcode_state
	Model  string // substring, case-insensitive
	Query  string // name, SN, or model substring
	Online string // "true" | "false" | ""
}

// FilterPrinterSnapshots applies f in memory (projections are small).
func FilterPrinterSnapshots(items []api.PrinterSnapshot, f PrinterListFilter) []api.PrinterSnapshot {
	if f.Tag == "" && f.State == "" && f.Model == "" && f.Query == "" && f.Online == "" {
		return items
	}
	tagWant := strings.ToLower(strings.TrimSpace(f.Tag))
	stateWant := strings.ToUpper(strings.TrimSpace(f.State))
	modelWant := strings.ToLower(strings.TrimSpace(f.Model))
	qWant := strings.ToLower(strings.TrimSpace(f.Query))
	out := make([]api.PrinterSnapshot, 0, len(items))
	for _, p := range items {
		if tagWant != "" && !hasTag(p.Tags, tagWant) {
			continue
		}
		if stateWant != "" && strings.ToUpper(p.GcodeState) != stateWant {
			continue
		}
		if modelWant != "" && !strings.Contains(strings.ToLower(p.Model), modelWant) {
			continue
		}
		if qWant != "" {
			hay := strings.ToLower(p.Name + " " + p.ExternalSN + " " + p.Model)
			if !strings.Contains(hay, qWant) {
				continue
			}
		}
		switch strings.ToLower(strings.TrimSpace(f.Online)) {
		case "true", "1":
			if !p.Online {
				continue
			}
		case "false", "0":
			if p.Online {
				continue
			}
		}
		out = append(out, p)
	}
	return out
}

func hasTag(tags []string, want string) bool {
	for _, t := range tags {
		if strings.ToLower(strings.TrimSpace(t)) == want {
			return true
		}
	}
	return false
}
