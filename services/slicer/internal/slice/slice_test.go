package slice

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestJobUnset(t *testing.T) {
	got, err := Job(context.Background(), "  ", Request{Name: " plate "})
	if err != nil || got.Sliced || got.Name != "plate" || got.Reason != "ORCA_BIN unset" {
		t.Fatalf("%+v %v", got, err)
	}
	if _, err := Job(context.Background(), "", Request{}); err == nil {
		t.Fatal("expected name")
	}
}

func TestJobBinNeedsInput(t *testing.T) {
	got, err := Job(context.Background(), "/usr/bin/true", Request{Name: "plate"})
	if err != nil || got.Sliced || got.Reason != "input_path required" {
		t.Fatalf("%+v %v", got, err)
	}
}

func TestJobInvokesAndWritesOutput(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "model.3mf")
	if err := os.WriteFile(in, []byte("model"), 0o644); err != nil {
		t.Fatal(err)
	}
	fake := filepath.Join(dir, "fake-orca")
	if runtime.GOOS == "windows" {
		t.Skip("shell fake")
	}
	script := "#!/bin/sh\n# find --export-3mf path and write it\nwhile [ $# -gt 0 ]; do\n  if [ \"$1\" = \"--export-3mf\" ]; then\n    shift\n    printf ok > \"$1\"\n    exit 0\n  fi\n  shift\ndone\nexit 2\n"
	if err := os.WriteFile(fake, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "plate.gcode.3mf")
	got, err := Job(context.Background(), fake, Request{
		Name:       "plate",
		InputPath:  in,
		OutputPath: out,
		Timeout:    5 * time.Second,
	})
	if err != nil || !got.Sliced || got.OutputPath != out {
		t.Fatalf("%+v %v", got, err)
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatal(err)
	}
}

func TestJobNonzeroExit(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "model.3mf")
	if err := os.WriteFile(in, []byte("model"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := Job(context.Background(), "/bin/false", Request{
		Name:       "plate",
		InputPath:  in,
		OutputPath: filepath.Join(dir, "out.gcode.3mf"),
		Timeout:    5 * time.Second,
		CommandContext: func(ctx context.Context, name string, arg ...string) *exec.Cmd {
			return exec.CommandContext(ctx, name, arg...)
		},
	})
	if err != nil || got.Sliced || got.ExitCode == 0 {
		t.Fatalf("%+v %v", got, err)
	}
}

func TestResolveRejectsDotDot(t *testing.T) {
	if _, err := resolvePath("/tmp", "../etc/passwd"); err == nil {
		t.Fatal("expected reject")
	}
}
