// Package version provides build-time version information for the CLI.
// Variables are populated via ldflags during the build process.
//
// Example build command:
//
//	go build -ldflags "-X github.com/StackEye-IO/stackeye-cli/internal/version.Version=1.0.0 \
//	  -X github.com/StackEye-IO/stackeye-cli/internal/version.Commit=$(git rev-parse --short HEAD) \
//	  -X github.com/StackEye-IO/stackeye-cli/internal/version.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
//	  -X github.com/StackEye-IO/stackeye-cli/internal/version.BuiltBy=makefile"
package version

import (
	"fmt"
	"runtime"
	"strings"
)

// Build information variables, set via ldflags at compile time.
var (
	// Version is the semantic version of the CLI (e.g., "1.0.0", "1.0.0-dirty").
	// Set via: -X github.com/StackEye-IO/stackeye-cli/internal/version.Version=...
	Version = "dev"

	// Commit is the short git commit SHA at build time.
	// Set via: -X github.com/StackEye-IO/stackeye-cli/internal/version.Commit=...
	Commit = "none"

	// Date is the UTC timestamp when the binary was built.
	// Set via: -X github.com/StackEye-IO/stackeye-cli/internal/version.Date=...
	Date = "unknown"

	// BuiltBy indicates how the binary was built (e.g., "goreleaser", "makefile", "go install").
	// Set via: -X github.com/StackEye-IO/stackeye-cli/internal/version.BuiltBy=...
	BuiltBy = "unknown"
)

// Info holds structured version information.
type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Date      string `json:"date"`
	BuiltBy   string `json:"builtBy"`
	GoVersion string `json:"goVersion"`
	Platform  string `json:"platform"`
}

// GetInfo returns structured version information.
func GetInfo() Info {
	return Info{
		Version:   Version,
		Commit:    Commit,
		Date:      Date,
		BuiltBy:   BuiltBy,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// String returns a formatted multi-line version string.
func (i Info) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("stackeye version %s\n", i.Version))
	b.WriteString(fmt.Sprintf("  Commit:     %s\n", i.Commit))
	b.WriteString(fmt.Sprintf("  Built:      %s\n", i.Date))
	b.WriteString(fmt.Sprintf("  Built by:   %s\n", i.BuiltBy))
	b.WriteString(fmt.Sprintf("  Go version: %s\n", i.GoVersion))
	b.WriteString(fmt.Sprintf("  Platform:   %s", i.Platform))
	return b.String()
}

// Short returns just the version string.
func (i Info) Short() string {
	return i.Version
}

// IsDev returns true if this is a development build.
func IsDev() bool {
	return Version == "dev"
}

// IsDirty returns true if the build was from a dirty working directory.
func IsDirty() bool {
	return strings.HasSuffix(Version, "-dirty")
}
