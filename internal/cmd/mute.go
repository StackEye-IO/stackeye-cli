// Package cmd implements the CLI commands for StackEye.
package cmd

import "github.com/spf13/cobra"

// NewMuteCmd creates and returns the mute parent command.
// This command provides management of alert mute periods.
func NewMuteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mute",
		Short: "Manage alert mute periods",
		Long: `Manage alert mute periods for your organization.

Mutes temporarily silence alert notifications for specific scopes during
maintenance windows or known outages. Mutes can target:

Scope Types:
  organization  Silence all alerts for the entire organization
  probe         Silence alerts for a specific probe
  channel       Silence a specific notification channel
  alert_type    Silence alerts of a specific type (down, degraded, ssl_expiry, etc.)

Mutes are time-limited and automatically expire after their duration.
You can also mark mutes as maintenance windows with custom names.

Available Commands:
  create        Create a new alert mute period

Examples:
  # Mute all alerts organization-wide for 1 hour
  stackeye mute create --scope organization --duration 60

  # Mute alerts for a specific probe during maintenance
  stackeye mute create --scope probe --probe-id <uuid> --duration 120 \
    --maintenance --maintenance-name "Server upgrade"

  # Mute a notification channel for 30 minutes
  stackeye mute create --scope channel --channel-id <uuid> --duration 30 \
    --reason "Testing channel configuration"

  # List all active mutes
  stackeye mute list

For more information about a specific command:
  stackeye mute [command] --help`,
		Aliases: []string{"mutes", "silence"},
	}

	// Register subcommands
	cmd.AddCommand(NewMuteCreateCmd())
	cmd.AddCommand(NewMuteListCmd())

	return cmd
}
