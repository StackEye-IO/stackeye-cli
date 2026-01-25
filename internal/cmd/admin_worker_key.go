// Package cmd implements the CLI commands for StackEye.
package cmd

import "github.com/spf13/cobra"

// NewAdminWorkerKeyCmd creates and returns the admin worker-key parent command.
// This command provides access to worker key management features.
func NewAdminWorkerKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "worker-key",
		Short: "Manage worker authentication keys",
		Long: `Manage worker authentication keys for regional probe workers.

Worker keys authenticate the probe workers that run in each monitoring region.
Each regional probe controller uses a worker key to authenticate with the
central API server.

Commands:
  create  Create a new worker key for a region
  delete  Delete a worker key permanently
  list    List all worker keys (future)
  health  Check worker health status (future)

Examples:
  # Create a new worker key for NYC region
  stackeye admin worker-key create --region nyc3

  # Create a worker key with a custom name
  stackeye admin worker-key create --region lon1 --name "London Probe Primary"`,
		Aliases: []string{"wk", "workerkey"},
	}

	// Register subcommands
	cmd.AddCommand(NewAdminWorkerKeyCreateCmd())
	cmd.AddCommand(NewAdminWorkerKeyDeleteCmd())

	return cmd
}
