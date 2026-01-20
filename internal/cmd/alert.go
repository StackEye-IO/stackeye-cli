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

Available Commands:
  list          List all alerts with optional filtering
  get           Get detailed alert information by ID
  ack           Acknowledge one or more alerts

Examples:
  # List all active alerts
  stackeye alert list

  # List only critical alerts
  stackeye alert list --severity critical

  # Filter by status
  stackeye alert list --status active

  # Filter by probe
  stackeye alert list --probe <uuid>

For more information about a specific command:
  stackeye alert [command] --help`,
		Aliases: []string{"alerts", "alerting"},
	}

	// Register subcommands
	cmd.AddCommand(NewAlertListCmd())
	cmd.AddCommand(NewAlertGetCmd())
	cmd.AddCommand(NewAlertAckCmd())

	return cmd
}
