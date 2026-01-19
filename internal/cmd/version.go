// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Build information variables, set via ldflags at compile time.
// Example build command:
//
//	go build -ldflags "-X github.com/StackEye-IO/stackeye-cli/internal/cmd.Version=1.0.0 \
//	  -X github.com/StackEye-IO/stackeye-cli/internal/cmd.GitCommit=$(git rev-parse --short HEAD) \
//	  -X github.com/StackEye-IO/stackeye-cli/internal/cmd.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
var (
	// Version is the semantic version of the CLI (e.g., "1.0.0").
	Version = "dev"

	// GitCommit is the short git commit SHA at build time.
	GitCommit = "unknown"

	// BuildTime is the UTC timestamp when the binary was built.
	BuildTime = "unknown"
)

// NewVersionCmd creates and returns the version command.
func NewVersionCmd() *cobra.Command {
	var shortFlag bool

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Long: `Display version, git commit SHA, build date, and runtime information.

Use --short to display only the version number.`,
		// Override PersistentPreRunE to skip config loading.
		// The version command should work without a valid configuration.
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			if shortFlag {
				fmt.Println(Version)
				return
			}

			fmt.Printf("stackeye version %s\n", Version)
			fmt.Printf("  Git commit:  %s\n", GitCommit)
			fmt.Printf("  Built:       %s\n", BuildTime)
			fmt.Printf("  Go version:  %s\n", runtime.Version())
			fmt.Printf("  OS/Arch:     %s/%s\n", runtime.GOOS, runtime.GOARCH)
		},
	}

	cmd.Flags().BoolVar(&shortFlag, "short", false, "print only the version number")

	return cmd
}
