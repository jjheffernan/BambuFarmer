package artifacts

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStoreWriteOpenRejectEscape(t *testing.T) {
	dir := t.TempDir()
	s := Store{Dir: dir}
	if err := s.Write("plate.gcode.3mf", []byte("abc")); err != nil {
		t.Fatal(err)
	}
	f, err := s.Open("plate.gcode.3mf")
	if err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	if _, err := s.AbsPath("../escape"); err == nil {
		t.Fatal("expected escape reject")
	}
	if _, err := s.AbsPath("/abs"); err == nil {
		t.Fatal("expected abs reject")
	}
	got, err := os.ReadFile(filepath.Join(dir, "plate.gcode.3mf"))
	if err != nil || string(got) != "abc" {
		t.Fatalf("%q %v", got, err)
	}
}

func TestExtFor(t *testing.T) {
	if extFor("a.gcode.3mf") != ".gcode.3mf" {
		t.Fatal(extFor("a.gcode.3mf"))
	}
	if extFor("x.3mf") != ".3mf" {
		t.Fatal(extFor("x.3mf"))
	}
}
