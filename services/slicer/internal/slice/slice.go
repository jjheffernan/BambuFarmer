package slice

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Result is the slice-job response. The Orca binary is not bundled.
type Result struct {
	Sliced     bool   `json:"sliced"`
	Name       string `json:"name,omitempty"`
	Reason     string `json:"reason,omitempty"`
	OutputPath string `json:"output_path,omitempty"`
	ExitCode   int    `json:"exit_code,omitempty"`
}

// Request is one slice attempt.
type Request struct {
	Name           string
	InputPath      string
	OutputPath     string
	PrinterJSON    string // optional --load-settings path
	ProcessJSON    string // optional --load-settings path (joined with printer)
	FilamentJSON   string // optional --load-filaments path
	WorkDir        string // default dir for relative paths / output
	Timeout        time.Duration
	CommandContext func(ctx context.Context, name string, arg ...string) *exec.Cmd // tests
}

// Job runs Orca when ORCA_BIN is set. An empty bin does not invoke anything.
func Job(ctx context.Context, bin string, req Request) (Result, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return Result{}, errName
	}
	bin = strings.TrimSpace(bin)
	if bin == "" {
		return Result{Sliced: false, Name: name, Reason: "ORCA_BIN unset"}, nil
	}
	input := strings.TrimSpace(req.InputPath)
	if input == "" {
		return Result{Sliced: false, Name: name, Reason: "input_path required"}, nil
	}
	work := strings.TrimSpace(req.WorkDir)
	inAbs, err := resolvePath(work, input)
	if err != nil {
		return Result{}, err
	}
	out := strings.TrimSpace(req.OutputPath)
	if out == "" {
		base := strings.TrimSuffix(filepath.Base(inAbs), filepath.Ext(inAbs))
		if base == "" || base == "." {
			base = name
		}
		out = filepath.Join(filepath.Dir(inAbs), base+".gcode.3mf")
	}
	outAbs, err := resolvePath(work, out)
	if err != nil {
		return Result{}, err
	}
	args := []string{inAbs}
	settingsParts := make([]string, 0, 2)
	for _, p := range []string{req.PrinterJSON, req.ProcessJSON} {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		abs, err := resolvePath(work, p)
		if err != nil {
			return Result{}, err
		}
		settingsParts = append(settingsParts, abs)
	}
	if len(settingsParts) > 0 {
		args = append(args, "--load-settings", strings.Join(settingsParts, ";"))
	}
	if f := strings.TrimSpace(req.FilamentJSON); f != "" {
		fAbs, err := resolvePath(work, f)
		if err != nil {
			return Result{}, err
		}
		args = append(args, "--load-filaments", fAbs)
	}
	args = append(args, "--arrange", "1", "--slice", "0", "--export-3mf", outAbs)

	timeout := req.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Minute
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmdFn := req.CommandContext
	if cmdFn == nil {
		cmdFn = exec.CommandContext
	}
	cmd := cmdFn(runCtx, bin, args...)
	outBytes, runErr := cmd.CombinedOutput()
	exit := 0
	if runErr != nil {
		var ee *exec.ExitError
		if errors.As(runErr, &ee) {
			exit = ee.ExitCode()
		} else {
			return Result{Sliced: false, Name: name, Reason: runErr.Error(), ExitCode: -1}, nil
		}
	}
	if exit != 0 {
		reason := strings.TrimSpace(string(outBytes))
		if reason == "" {
			reason = fmt.Sprintf("orca exit %d", exit)
		}
		if len(reason) > 400 {
			reason = reason[:400]
		}
		return Result{Sliced: false, Name: name, Reason: reason, ExitCode: exit, OutputPath: outAbs}, nil
	}
	if st, err := os.Stat(outAbs); err != nil || st.IsDir() {
		return Result{Sliced: false, Name: name, Reason: "orca exited 0 but output missing", OutputPath: outAbs}, nil
	}
	return Result{Sliced: true, Name: name, OutputPath: outAbs}, nil
}

func resolvePath(work, p string) (string, error) {
	p = strings.TrimSpace(p)
	if p == "" {
		return "", fmt.Errorf("path required")
	}
	if strings.Contains(p, "..") {
		return "", fmt.Errorf("invalid path")
	}
	if filepath.IsAbs(p) {
		return filepath.Clean(p), nil
	}
	if work == "" {
		return "", fmt.Errorf("relative path needs work_dir")
	}
	abs := filepath.Join(work, filepath.FromSlash(p))
	rel, err := filepath.Rel(work, abs)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("path escape")
	}
	return abs, nil
}

type nameErr string

func (e nameErr) Error() string { return string(e) }

const errName nameErr = "name required"
