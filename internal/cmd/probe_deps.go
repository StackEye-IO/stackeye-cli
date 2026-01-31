// Package cmd implements the CLI commands for StackEye.
package cmd

import "github.com/spf13/cobra"

// NewProbeDepsCmd creates and returns the probe deps parent command.
// This command provides management of probe dependencies for hierarchical alerting.
func NewProbeDepsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deps",
		Short: "Manage probe dependencies",
		Long: `Manage parent/child probe dependencies for hierarchical alerting.

Dependencies help reduce alert noise by suppressing child alerts when a parent
probe fails. For example, if your database goes down, you don't want to receive
separate alerts for every web server that depends on it.

When a parent probe fails:
  - Child probes are marked as UNREACHABLE (not DOWN)
  - Alerts for child probes are suppressed
  - When the parent recovers, children are re-evaluated

Dependency Operations:
  list      List dependencies for a specific probe
  add       Add a parent dependency to a probe
  remove    Remove a parent dependency from a probe
  clear     Remove all dependencies from a probe
  tree      Display organization-wide dependency tree
  wizard    Interactive guided dependency setup

Examples:
  # List dependencies for a probe
  stackeye probe deps list <probe-id>

  # Add a parent dependency (web server depends on database)
  stackeye probe deps add <web-server-id> --parent <database-id>

  # Remove a parent dependency
  stackeye probe deps remove <web-server-id> --parent <database-id>

  # Clear all dependencies from a probe
  stackeye probe deps clear <probe-id>

  # View ASCII tree of all dependencies
  stackeye probe deps tree

  # Run interactive dependency wizard
  stackeye probe deps wizard

Common Dependency Patterns:
  Database -> Application Servers
    When database fails, suppress app server alerts

  Load Balancer -> Backend Servers
    When LB fails, suppress backend alerts

  Core Router -> All Downstream Devices
    When router fails, suppress all device alerts

For more information about a specific command:
  stackeye probe deps [command] --help`,
		Aliases: []string{"dependencies", "dependency", "dep"},
	}

	// Register implemented subcommands
	cmd.AddCommand(NewProbeDepsListCmd())
	cmd.AddCommand(NewProbeDepsAddCmd())

	// Subcommands to be registered as they are implemented:
	// - Task #8025: NewProbeDepsRemoveCmd()
	// - Task #8026: NewProbeDepsClearCmd()
	// - Task #8027: NewProbeDepsTreeCmd()
	// - Task #8028: NewProbeDepsWizardCmd()

	return cmd
}
