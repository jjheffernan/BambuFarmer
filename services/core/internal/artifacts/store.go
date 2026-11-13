package artifacts

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jjheffernan/BambuFarmer/packages/objectstore"
)

// Store stages .gcode.3mf under Dir and optionally mirrors to OBJECT_STORE_*.
type Store struct {
	Dir    string
	Remote *objectstore.Client
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

// Write creates/replaces object_key locally and mirrors to object store when configured.
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
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return err
	}
	if s.Remote != nil {
		if err := s.Remote.Put(context.Background(), objectKey, b); err != nil {
			return fmt.Errorf("artifacts: object store put: %w", err)
		}
	}
	return nil
}

// Open reads object_key from local disk, falling back to object store.
func (s Store) Open(objectKey string) (io.ReadCloser, error) {
	path, err := s.AbsPath(objectKey)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err == nil {
		return f, nil
	}
	if s.Remote == nil || !os.IsNotExist(err) {
		return nil, err
	}
	return s.Remote.Get(context.Background(), objectKey)
}

// Remove deletes object_key locally and from object store when configured.
func (s Store) Remove(objectKey string) error {
	path, err := s.AbsPath(objectKey)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	if s.Remote != nil {
		if err := s.Remote.Delete(context.Background(), objectKey); err != nil {
			return fmt.Errorf("artifacts: object store delete: %w", err)
		}
	}
	return nil
}
