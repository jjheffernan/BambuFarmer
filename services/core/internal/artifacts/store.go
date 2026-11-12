package artifacts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Store is a local directory for staged .gcode.3mf bytes (MVP stand-in for MinIO).
type Store struct {
	Dir string
}

// ResolveDir picks ARTIFACTS_DIR or a default under the process cwd.
func ResolveDir(env string) string {
	if env != "" {
		return env
	}
	return "/data/artifacts"
}

// EnsureDir creates the store root.
func (s Store) EnsureDir() error {
	if s.Dir == "" {
		return fmt.Errorf("artifacts: empty dir")
	}
	return os.MkdirAll(s.Dir, 0o755)
}

// AbsPath joins object_key under Dir; rejects path escape.
func (s Store) AbsPath(objectKey string) (string, error) {
	key := strings.TrimSpace(objectKey)
	if key == "" || strings.Contains(key, "..") || strings.HasPrefix(key, "/") {
		return "", fmt.Errorf("artifacts: invalid object_key")
	}
	abs := filepath.Join(s.Dir, filepath.FromSlash(key))
	rel, err := filepath.Rel(s.Dir, abs)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("artifacts: path escape")
	}
	return abs, nil
}

// Write creates/replaces object_key with b.
func (s Store) Write(objectKey string, b []byte) error {
	if err := s.EnsureDir(); err != nil {
		return err
	}
	path, err := s.AbsPath(objectKey)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

// Open reads object_key.
func (s Store) Open(objectKey string) (*os.File, error) {
	path, err := s.AbsPath(objectKey)
	if err != nil {
		return nil, err
	}
	return os.Open(path)
}
