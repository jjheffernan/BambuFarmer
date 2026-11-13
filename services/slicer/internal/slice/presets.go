package slice

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// PresetFiles are absolute paths written for Orca --load-settings / --load-filaments.
type PresetFiles struct {
	PrinterJSON  string
	ProcessJSON  string
	FilamentJSON string
}

// HTTPDo is injectable for tests.
type HTTPDo func(req *http.Request) (*http.Response, error)

// MaterializePresets fetches core presets by id and writes JSON files under workDir.
// Empty CORE_URL with preset ids returns an error. Missing ids leave paths empty.
func MaterializePresets(ctx context.Context, coreURL, workDir, printPresetID, filamentPresetID string, do HTTPDo) (PresetFiles, error) {
	printPresetID = strings.TrimSpace(printPresetID)
	filamentPresetID = strings.TrimSpace(filamentPresetID)
	if printPresetID == "" && filamentPresetID == "" {
		return PresetFiles{}, nil
	}
	coreURL = strings.TrimRight(strings.TrimSpace(coreURL), "/")
	if coreURL == "" {
		return PresetFiles{}, fmt.Errorf("CORE_URL unset")
	}
	if strings.TrimSpace(workDir) == "" {
		return PresetFiles{}, fmt.Errorf("work_dir required for preset materialize")
	}
	if do == nil {
		do = http.DefaultClient.Do
	}
	dir := filepath.Join(workDir, ".slice-presets")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return PresetFiles{}, err
	}
	var out PresetFiles
	if printPresetID != "" {
		var list []struct {
			ID          string         `json:"id"`
			ProcessJSON map[string]any `json:"process_json"`
			PrinterJSON map[string]any `json:"printer_json"`
		}
		if err := getJSON(ctx, do, coreURL+"/api/v1/presets/print", &list); err != nil {
			return PresetFiles{}, err
		}
		found := false
		for _, p := range list {
			if p.ID != printPresetID {
				continue
			}
			found = true
			if len(p.PrinterJSON) > 0 {
				path := filepath.Join(dir, printPresetID+"-printer.json")
				if err := writeJSONFile(path, p.PrinterJSON); err != nil {
					return PresetFiles{}, err
				}
				out.PrinterJSON = path
			}
			if len(p.ProcessJSON) > 0 {
				path := filepath.Join(dir, printPresetID+"-process.json")
				if err := writeJSONFile(path, p.ProcessJSON); err != nil {
					return PresetFiles{}, err
				}
				out.ProcessJSON = path
			}
			break
		}
		if !found {
			return PresetFiles{}, fmt.Errorf("print preset %q not found", printPresetID)
		}
	}
	if filamentPresetID != "" {
		var list []map[string]any
		if err := getJSON(ctx, do, coreURL+"/api/v1/presets/filament", &list); err != nil {
			return PresetFiles{}, err
		}
		found := false
		for _, p := range list {
			id, _ := p["id"].(string)
			if id != filamentPresetID {
				continue
			}
			found = true
			path := filepath.Join(dir, filamentPresetID+"-filament.json")
			if err := writeJSONFile(path, p); err != nil {
				return PresetFiles{}, err
			}
			out.FilamentJSON = path
			break
		}
		if !found {
			return PresetFiles{}, fmt.Errorf("filament preset %q not found", filamentPresetID)
		}
	}
	return out, nil
}

func getJSON(ctx context.Context, do HTTPDo, url string, dest any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	res, err := do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 4<<20))
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("core %s: %s", res.Status, strings.TrimSpace(string(body)))
	}
	return json.Unmarshal(body, dest)
}

func writeJSONFile(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

// ApplyPresetFiles fills empty Request path fields from materialized presets.
func ApplyPresetFiles(req *Request, files PresetFiles) {
	if req == nil {
		return
	}
	if req.PrinterJSON == "" && files.PrinterJSON != "" {
		req.PrinterJSON = files.PrinterJSON
	}
	if req.ProcessJSON == "" && files.ProcessJSON != "" {
		req.ProcessJSON = files.ProcessJSON
	}
	if req.FilamentJSON == "" && files.FilamentJSON != "" {
		req.FilamentJSON = files.FilamentJSON
	}
}
