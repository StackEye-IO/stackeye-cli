// Package cli provides the public API for extending the StackEye CLI.
//
// This package exports the root command and utility functions that allow
// other modules to extend the CLI with additional commands. This is used
// by the private stackeye-cli-admin module to add admin-only commands.
//
// Example usage in an extending CLI:
//
//	package main
//
//	import (
//	    "os"
//	    "github.com/StackEye-IO/stackeye-cli/pkg/cli"
//	)
//
//	func main() {
//	    rootCmd := cli.RootCmd()
//	    rootCmd.AddCommand(myCustomCommand())
//	    os.Exit(cli.ExecuteWithExitCode(rootCmd))
//	}
package cli

import (
	"github.com/StackEye-IO/stackeye-cli/internal/cmd"
	clierrors "github.com/StackEye-IO/stackeye-cli/internal/errors"
	"github.com/spf13/cobra"
)

// RootCmd returns the root command for the StackEye CLI.
// This allows other packages to add subcommands to extend the CLI.
//
// The returned command has all standard commands (probe, alert, channel, etc.)
// already registered and is ready for execution.
func RootCmd() *cobra.Command {
	return cmd.RootCmd()
}

// ExecuteWithExitCode runs the given command and returns an appropriate exit code.
// This maps errors to exit codes for proper CLI behavior:
//   - 0: Success
//   - 1: General error
//   - 2: Command misuse (invalid arguments)
//   - 3: Authentication required
//   - 4: Permission denied
//   - 5: Resource not found
//   - 6: Rate limited
//   - 7: Server error
//   - 8: Network error
//   - 9: Timeout
//   - 10: Plan limit exceeded
func ExecuteWithExitCode(cmd *cobra.Command) int {
	err := cmd.Execute()
	return clierrors.HandleError(err)
}

// GetConfig returns the loaded configuration after Execute() has been called.
// Returns nil if called before Execute() or if config loading failed.
//
// This is useful for commands that need access to authentication tokens,
// current context, or user preferences.
func GetConfig() interface{} {
	return cmd.GetConfig()
}

// GetVerbosity returns the current verbosity level from CLI flags.
// Returns 0 if no verbosity flags were set, 1 for -v, 2 for -vv, etc.
//
// This allows extended CLIs to pass verbosity settings to SDK clients
// for debug logging.
func GetVerbosity() int {
	return cmd.GetVerbosity()
}
