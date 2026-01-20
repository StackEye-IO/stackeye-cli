// Package cmd implements the CLI commands for StackEye.
package cmd

import "github.com/spf13/cobra"

// NewAlertCmd creates and returns the alert parent command.
// This command provides management of monitoring alerts.
func NewAlertCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alert",
		Short: "Manage monitoring alerts",
		Long: `Manage monitoring alerts for your organization.

Alerts are triggered when a probe detects an issue with a monitored endpoint.
Each alert tracks the lifecycle of an incident from detection through resolution,
including acknowledgment status and notification history.

Alert States:
  active        Issue detected, not yet acknowledged
  acknowledged  Issue recognized, being investigated
  resolved      Issue fixed, alert closed

Alert Severities:
  critical      Service down or major outage
  warning       Degraded performance or minor issue
  info          Informational, no action required

Alert Operations:
  list          List all alerts with optional filtering
  get           Get detailed alert information
  ack           Acknowledge an alert (mark as being investigated)
  resolve       Resolve an alert (mark as fixed)
  history       View alert timeline and notification history

Examples:
  # List all active alerts
  stackeye alert list

  # List only critical alerts
  stackeye alert list --severity critical

  # Get details for a specific alert
  stackeye alert get abc123

  # Acknowledge an alert with a note
  stackeye alert ack abc123 --note "Investigating slow database queries"

  # Resolve an alert
  stackeye alert resolve abc123 --note "Fixed by restarting database connection pool"

  # View alert timeline
  stackeye alert history abc123

For more information about a specific command:
  stackeye alert [command] --help`,
		Aliases: []string{"alerts", "alerting"},
	}

	// Subcommands will be registered as they are implemented:
	// cmd.AddCommand(NewAlertListCmd())
	// cmd.AddCommand(NewAlertGetCmd())
	// cmd.AddCommand(NewAlertAckCmd())
	// cmd.AddCommand(NewAlertResolveCmd())
	// cmd.AddCommand(NewAlertHistoryCmd())

	return cmd
}
