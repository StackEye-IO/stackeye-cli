// Package cmd implements the CLI commands for StackEye.
package cmd

import "github.com/spf13/cobra"

// NewMaintenanceCmd creates and returns the maintenance parent command.
// This command provides management of scheduled maintenance windows.
func NewMaintenanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "maintenance",
		Short: "Manage scheduled maintenance windows",
		Long: `Manage scheduled maintenance windows for your organization.

Maintenance windows temporarily suppress alert notifications during planned
downtime, such as deployments, upgrades, or infrastructure changes. Unlike
ad-hoc mutes, maintenance windows are named and scheduled for better tracking.

Key Features:
  - Named windows for audit trails and team coordination
  - Scheduled start times for planned maintenance
  - Scope to specific probes or organization-wide
  - Automatic expiration after specified duration

Examples:
  # Schedule a 2-hour maintenance window for a specific probe
  stackeye maintenance create --name "Server Upgrade" \
    --probe-id <uuid> --duration 120

  # Schedule organization-wide maintenance starting at a specific time
  stackeye maintenance create --name "Network Migration" \
    --organization-wide --duration 60 \
    --starts-at 2024-01-15T02:00:00Z

For more information about a specific command:
  stackeye maintenance [command] --help`,
		Aliases: []string{"maint", "mw"},
	}

	// Register subcommands
	cmd.AddCommand(NewMaintenanceCreateCmd())
	cmd.AddCommand(NewMaintenanceListCmd())
	cmd.AddCommand(NewMaintenanceCalendarCmd())
	cmd.AddCommand(NewMaintenanceDeleteCmd())

	return cmd
}
