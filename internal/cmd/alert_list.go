// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	clierrors "github.com/StackEye-IO/stackeye-cli/internal/errors"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// alertListTimeout is the maximum time to wait for the API response.
const alertListTimeout = 30 * time.Second

// alertListFlags holds the flag values for the alert list command.
type alertListFlags struct {
	status   string
	severity string
	probeID  string
	page     int
	limit    int
}

// NewAlertListCmd creates and returns the alert list subcommand.
func NewAlertListCmd() *cobra.Command {
	flags := &alertListFlags{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all monitoring alerts",
		Long: `List all monitoring alerts in your organization.

Displays alert status, severity, probe name, triggered time, and duration.
Results are paginated and can be filtered by status, severity, or probe.

Status Values:
  active        Alert is open and needs attention
  acknowledged  Alert has been seen, investigation in progress
  resolved      Alert has been resolved

Severity Levels:
  critical      Service down or major outage
  warning       Degraded performance or minor issue
  info          Informational, no action required

Examples:
  # List all alerts
  stackeye alert list

  # List only active alerts
  stackeye alert list -s active

  # List critical alerts only
  stackeye alert list --severity critical

  # List alerts for a specific probe
  stackeye alert list --probe abc123-def456-...

  # Output as JSON for scripting
  stackeye alert list -o json

  # Paginate through results
  stackeye alert list --page 2 --limit 50`,
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAlertList(cmd.Context(), flags)
		},
	}

	// Define command-specific flags
	cmd.Flags().StringVarP(&flags.status, "status", "s", "", "filter by status: active, acknowledged, resolved")
	cmd.Flags().StringVar(&flags.severity, "severity", "", "filter by severity: critical, warning, info")
	cmd.Flags().StringVar(&flags.probeID, "probe", "", "filter by probe ID")
	cmd.Flags().IntVar(&flags.page, "page", 1, "page number for pagination")
	cmd.Flags().IntVar(&flags.limit, "limit", 20, "results per page (max: 100)")

	return cmd
}

// runAlertList executes the alert list command logic.
func runAlertList(ctx context.Context, flags *alertListFlags) error {
	// Validate all flags before making any API calls
	if flags.limit < 1 || flags.limit > 100 {
		return fmt.Errorf("invalid limit %d: must be between 1 and 100", flags.limit)
	}

	if flags.page < 1 {
		return fmt.Errorf("invalid page %d: must be at least 1", flags.page)
	}

	var alertStatus client.AlertStatus
	if flags.status != "" {
		switch flags.status {
		case "active":
			alertStatus = client.AlertStatusActive
		case "acknowledged":
			alertStatus = client.AlertStatusAcknowledged
		case "resolved":
			alertStatus = client.AlertStatusResolved
		default:
			return clierrors.InvalidValueError("--status", flags.status, clierrors.ValidAlertStatuses)
		}
	}

	var alertSeverity client.AlertSeverity
	if flags.severity != "" {
		switch flags.severity {
		case "critical":
			alertSeverity = client.AlertSeverityCritical
		case "warning":
			alertSeverity = client.AlertSeverityWarning
		case "info":
			alertSeverity = client.AlertSeverityInfo
		default:
			return clierrors.InvalidValueError("--severity", flags.severity, clierrors.ValidSeverities)
		}
	}

	var probeID *uuid.UUID
	if flags.probeID != "" {
		parsed, err := uuid.Parse(flags.probeID)
		if err != nil {
			return fmt.Errorf("invalid probe ID %q: %w", flags.probeID, err)
		}
		probeID = &parsed
	}

	// Get authenticated API client (after validation passes)
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Build list options from validated flags
	// SDK uses offset-based pagination, convert page to offset
	offset := (flags.page - 1) * flags.limit
	opts := &client.ListAlertsOptions{
		Limit:    flags.limit,
		Offset:   offset,
		Status:   alertStatus,
		Severity: alertSeverity,
		ProbeID:  probeID,
	}

	// Call SDK to list alerts with timeout
	reqCtx, cancel := context.WithTimeout(ctx, alertListTimeout)
	defer cancel()

	result, err := client.ListAlerts(reqCtx, apiClient, opts)
	if err != nil {
		return fmt.Errorf("failed to list alerts: %w", err)
	}

	// Handle empty results
	if len(result.Alerts) == 0 {
		return output.PrintEmpty("No alerts found")
	}

	// Print the alerts using the configured output format
	return output.PrintAlerts(result.Alerts)
}
