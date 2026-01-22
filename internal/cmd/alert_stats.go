// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
)

// alertStatsTimeout is the maximum time to wait for the API response.
const alertStatsTimeout = 30 * time.Second

// alertStatsFlags holds the flag values for the alert stats command.
type alertStatsFlags struct {
	period string
}

// NewAlertStatsCmd creates and returns the alert stats subcommand.
func NewAlertStatsCmd() *cobra.Command {
	flags := &alertStatsFlags{}

	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show alert statistics",
		Long: `Show alert statistics for your organization.

Displays aggregate alert metrics including total counts by status and severity,
plus mean time to acknowledge (MTTA) and mean time to resolve (MTTR).

Time Periods:
  24h   Last 24 hours (default)
  7d    Last 7 days
  30d   Last 30 days

Metrics Shown:
  Total Alerts      Total alerts in the period
  Active            Currently active (unresolved) alerts
  Acknowledged      Alerts that have been acknowledged
  Resolved          Alerts that have been resolved
  By Severity       Breakdown by critical/warning/info
  MTTA              Mean Time To Acknowledge (seconds)
  MTTR              Mean Time To Resolve (seconds)

Examples:
  # Show stats for the last 24 hours
  stackeye alert stats

  # Show stats for the last 7 days
  stackeye alert stats --period 7d

  # Show stats for the last 30 days
  stackeye alert stats --period 30d

  # Output as JSON for scripting
  stackeye alert stats -o json`,
		Aliases: []string{"statistics", "summary"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAlertStats(cmd.Context(), flags)
		},
	}

	// Define command-specific flags
	cmd.Flags().StringVar(&flags.period, "period", "24h", "time period: 24h, 7d, 30d")

	return cmd
}

// runAlertStats executes the alert stats command logic.
func runAlertStats(ctx context.Context, flags *alertStatsFlags) error {
	// Validate period flag
	switch flags.period {
	case "24h", "7d", "30d":
		// Valid periods
	default:
		return fmt.Errorf("invalid period %q: must be 24h, 7d, or 30d", flags.period)
	}

	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Call SDK to get alert stats with timeout
	reqCtx, cancel := context.WithTimeout(ctx, alertStatsTimeout)
	defer cancel()

	stats, err := client.GetAlertStats(reqCtx, apiClient, flags.period)
	if err != nil {
		return fmt.Errorf("failed to get alert statistics: %w", err)
	}

	// Print the stats using the configured output format
	return output.PrintAlertStats(stats)
}
