package slice

import "strings"

// Result is the slice-job response. The Orca binary is not bundled.
type Result struct {
	Sliced bool   `json:"sliced"`
	Name   string `json:"name,omitempty"`
	Reason string `json:"reason,omitempty"`
}

// Job reports whether a slice can run. An empty ORCA_BIN does not invoke anything.
func Job(bin, name string) (Result, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Result{}, errName
	}
	if strings.TrimSpace(bin) == "" {
		return Result{Sliced: false, Name: name, Reason: "ORCA_BIN unset"}, nil
	}
	return Result{Sliced: false, Name: name, Reason: "orca binary is operator-supplied and not invoked"}, nil
}

type nameErr string

func (e nameErr) Error() string { return string(e) }

const errName nameErr = "name required"
