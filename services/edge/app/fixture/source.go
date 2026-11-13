package fixture

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/jjheffernan/BambuFarmer/services/edge/app/farmmanager"
)

// Source cycles redacted FM device fixtures for local/dev without a live server.
type Source struct {
	dir   string
	files []string
	idx   atomic.Uint64
}

// New loads *.json device lists from dir (expects devices_idle.json / devices_running.json).
func New(dir string) (*Source, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if filepath.Ext(name) != ".json" {
			continue
		}
		if name == "captain.json" || name == "tags.json" {
			continue
		}
		files = append(files, filepath.Join(dir, name))
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no device fixtures in %s", dir)
	}
	return &Source{dir: dir, files: files}, nil
}

// LoadTags reads tags.json from the fixture directory. Missing file yields nil.
func LoadTags(dir string) ([]farmmanager.Tag, error) {
	path := filepath.Join(dir, "tags.json")
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var resp farmmanager.TagsResponse
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, err
	}
	return resp.Tags, nil
}

// NextDevices returns the next fixture device list (round-robin).
func (s *Source) NextDevices() ([]farmmanager.Device, error) {
	i := s.idx.Add(1) - 1
	path := s.files[int(i)%len(s.files)]
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var resp farmmanager.DevicesResponse
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, err
	}
	return resp.Devices, nil
}
