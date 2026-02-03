// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	clioutput "github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/spf13/cobra"
)

// NewEnvCmd creates and returns the env command.
func NewEnvCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env",
		Short: "List environment variables that affect CLI behavior",
		Long: `Display all environment variables recognized by the StackEye CLI.

Shows each variable's current value (sensitive values are masked) and
whether it is set in the environment. This is useful for debugging
configuration issues or verifying CI/CD pipeline settings.

Environment variables override config file settings when set.

Examples:
  # List all environment variables
  stackeye env

  # Output as JSON for scripting
  stackeye env -o json

  # Show extended details
  stackeye env -o wide`,
		// Override PersistentPreRunE to skip config loading.
		// The env command should work even if configuration is broken,
		// since its purpose is to help debug configuration issues.
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEnv()
		},
	}

	return cmd
}

func runEnv() error {
	rows := clioutput.CollectEnvVars()
	// Use PrintEnvVarsWithFormat because this command skips config loading
	// (PersistentPreRunE override), so the global printer won't have the
	// output format preference set. Read the flag value directly instead.
	return clioutput.PrintEnvVarsWithFormat(rows, outputFormat)
}
