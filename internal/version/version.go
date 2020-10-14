package version

import (
	"fmt"
	"runtime"
)

// Version information.
var (
	Version   = "None"
	BuildTS   = "None"
	GitCommit = "None"
	GitBranch = "None"
)

// BuildInfo describes the compile time information.
type BuildInfo struct {
	// Version is the current semver.
	Version string `json:"version,omitempty"`
	// BuildTS is the timestamp
	BuildTS string `json:"build_timestamp,omitempty"`
	// GitBranch is the git rev-parse --abbrev-ref HEAD.
	GitBranch string `json:"git_branch,omitempty"`
	// GitCommit is the git sha1.
	GitCommit string `json:"git_commit,omitempty"`
	// GoVersion is the version of the Go compiler used.
	GoVersion string `json:"go_version,omitempty"`
}

func Info() string {
	info := BuildInfo{
		Version:   Version,
		BuildTS:   BuildTS,
		GitBranch: GitBranch,
		GitCommit: GitCommit,
		GoVersion: runtime.Version(),
	}
	return fmt.Sprintf("%#v", info)
}
