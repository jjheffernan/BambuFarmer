package vp

import (
	"archive/zip"
	"bytes"
	"io"
	"regexp"
	"strings"
)

var (
	rePrinterModel = regexp.MustCompile(`(?i)"printer_model"\s*:\s*"([^"]+)"`)
	rePrinterModel2 = regexp.MustCompile(`(?i)printer_model["\s:=]+([A-Za-z0-9._-]+)`)
	reFilamentType = regexp.MustCompile(`(?i)"filament_type"\s*:\s*"([^"]+)"`)
	reFilamentType2 = regexp.MustCompile(`(?i)"type"\s*:\s*"(PLA|PETG|ABS|ASA|TPU|PA|PC|PVA)[^"]*"`)
)

// PlateHints are extracted from a .3mf (zip) when possible.
type PlateHints struct {
	TargetModel string   `json:"target_model,omitempty"`
	Filaments   []string `json:"filaments,omitempty"`
}

// HintsFrom3MF scans zip contents for printer_model / filament_type strings.
func HintsFrom3MF(body []byte) PlateHints {
	if len(body) < 4 || body[0] != 'P' || body[1] != 'K' {
		// Not a zip — try plain text scan.
		return scanHints(string(body))
	}
	zr, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return scanHints(string(body))
	}
	var h PlateHints
	seenFil := map[string]bool{}
	for _, f := range zr.File {
		name := strings.ToLower(f.Name)
		if !strings.Contains(name, "config") && !strings.Contains(name, "json") &&
			!strings.Contains(name, "slice") && !strings.Contains(name, "project") &&
			!strings.Contains(name, "filament") && !strings.Contains(name, "model_settings") {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			continue
		}
		raw, _ := io.ReadAll(io.LimitReader(rc, 1<<20))
		_ = rc.Close()
		part := scanHints(string(raw))
		if h.TargetModel == "" && part.TargetModel != "" {
			h.TargetModel = part.TargetModel
		}
		for _, fil := range part.Filaments {
			if !seenFil[fil] {
				seenFil[fil] = true
				h.Filaments = append(h.Filaments, fil)
			}
		}
	}
	return h
}

func scanHints(s string) PlateHints {
	var h PlateHints
	if m := rePrinterModel.FindStringSubmatch(s); len(m) > 1 {
		h.TargetModel = normalizeModel(m[1])
	} else if m := rePrinterModel2.FindStringSubmatch(s); len(m) > 1 {
		h.TargetModel = normalizeModel(m[1])
	}
	seen := map[string]bool{}
	for _, re := range []*regexp.Regexp{reFilamentType, reFilamentType2} {
		for _, m := range re.FindAllStringSubmatch(s, 8) {
			if len(m) < 2 {
				continue
			}
			fil := strings.ToUpper(strings.TrimSpace(m[1]))
			if fil == "" || seen[fil] {
				continue
			}
			seen[fil] = true
			h.Filaments = append(h.Filaments, fil)
		}
	}
	return h
}

func normalizeModel(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "Bambu Lab ", "")
	s = strings.ReplaceAll(s, "BambuLab ", "")
	return s
}

// MergeHints prefers explicit over extracted, then defaultModel.
func MergeHints(explicit, extracted, defaultModel string) string {
	if strings.TrimSpace(explicit) != "" {
		return strings.TrimSpace(explicit)
	}
	if strings.TrimSpace(extracted) != "" {
		return strings.TrimSpace(extracted)
	}
	return strings.TrimSpace(defaultModel)
}
