// Package cmd implements the CLI commands for StackEye.
package cmd

import "github.com/spf13/cobra"

// NewAdminCmd creates and returns the admin parent command.
// This command provides access to platform administration features.
// Admin commands require platform administrator privileges.
func NewAdminCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "admin",
		Short: "Platform administration commands",
		Long: `Platform administration commands for StackEye.

These commands are restricted to platform administrators and provide
access to system-wide management features including:

  - Worker key management (creating keys for regional probes)
  - Machine-to-machine (M2M) key management
  - User impersonation for support
  - Region configuration

Commands:
  worker-key  Manage worker authentication keys

Examples:
  # List all worker keys
  stackeye admin worker-key list

  # Create a new worker key for a region
  stackeye admin worker-key create --region nyc3

  # Check worker health status
  stackeye admin worker-key health`,
		Aliases: []string{"adm"},
	}

	// Register subcommands
	cmd.AddCommand(NewAdminWorkerKeyCmd())

	return cmd
}
