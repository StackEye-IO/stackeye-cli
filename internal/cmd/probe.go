// Package cmd implements the CLI commands for StackEye.
package cmd

import "github.com/spf13/cobra"

// NewProbeCmd creates and returns the probe parent command.
// This command provides management of monitoring probes.
func NewProbeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "probe",
		Short: "Manage monitoring probes",
		Long: `Manage monitoring probes for your organization.

Probes are the core building blocks of StackEye monitoring. Each probe
monitors a specific endpoint (URL, IP address, or service) at regular
intervals from multiple global regions.

Probe Operations:
  list          List all probes with status summary
  get           Get detailed probe information
  create        Create a new monitoring probe
  wizard        Interactive wizard for creating a probe
  update        Update probe configuration
  delete        Delete a probe
  pause         Temporarily pause monitoring
  resume        Resume a paused probe
  test          Run an immediate test check
  history       View probe check history
  stats         View uptime and response time statistics
  watch         Watch probe status with live updates

Channel Management:
  link-channel     Link a notification channel to a probe
  unlink-channel   Unlink a notification channel from a probe

Label Management:
  label            Add labels to a probe
  unlabel          Remove labels from a probe

Dependency Management:
  deps             Manage probe dependencies for hierarchical alerting

Examples:
  # List all probes
  stackeye probe list

  # Get details for a specific probe
  stackeye probe get api-health

  # Create an HTTP probe
  stackeye probe create --name "API Health" --url https://api.example.com/health

  # Pause monitoring for maintenance
  stackeye probe pause api-health

  # Resume after maintenance
  stackeye probe resume api-health

  # Run an immediate test check
  stackeye probe test api-health

  # View uptime statistics
  stackeye probe stats api-health

For more information about a specific command:
  stackeye probe [command] --help`,
		Aliases: []string{"probes", "monitor", "monitors"},
	}

	// Register subcommands
	cmd.AddCommand(NewProbeListCmd())
	cmd.AddCommand(NewProbeGetCmd())
	cmd.AddCommand(NewProbeCreateCmd())
	cmd.AddCommand(NewProbeWizardCmd())
	cmd.AddCommand(NewProbeUpdateCmd())
	cmd.AddCommand(NewProbeDeleteCmd())
	cmd.AddCommand(NewProbePauseCmd())
	cmd.AddCommand(NewProbeResumeCmd())
	cmd.AddCommand(NewProbeTestCmd())
	cmd.AddCommand(NewProbeHistoryCmd())
	cmd.AddCommand(NewProbeStatsCmd())
	cmd.AddCommand(NewProbeLinkChannelCmd())
	cmd.AddCommand(NewProbeUnlinkChannelCmd()) // Task #7074
	cmd.AddCommand(NewProbeDepsCmd())
	cmd.AddCommand(NewProbeLabelCmd())   // Task #8068
	cmd.AddCommand(NewProbeUnlabelCmd()) // Task #8069
	cmd.AddCommand(NewProbeWatchCmd())   // Task #7109

	return cmd
}
