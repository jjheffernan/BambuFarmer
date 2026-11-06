// Package version holds build-time identity for releases and the install wizard.
package version

// Values are overridden via -ldflags at release build time.
var (
	Version = "0.1.0-dev"
	Commit  = "unknown"
	BuiltAt = "unknown"
)

// Info is returned by /api/v1/version and the wizard.
type Info struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	BuiltAt string `json:"built_at"`
}

// Get returns the current build info.
func Get() Info {
	return Info{Version: Version, Commit: Commit, BuiltAt: BuiltAt}
}
