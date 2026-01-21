// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// alertHistoryTimeout is the maximum time to wait for the API response.
const alertHistoryTimeout = 30 * time.Second

// alertHistoryFlags holds the flag values for the alert history command.
type alertHistoryFlags struct {
	probeID string
	since   string
	until   string
	page    int
	limit   int
}

// NewAlertHistoryCmd creates and returns the alert history subcommand.
func NewAlertHistoryCmd() *cobra.Command {
	flags := &alertHistoryFlags{}

	cmd := &cobra.Command{
		Use:   "history",
		Short: "Show historical alerts",
		Long: `Show historical (resolved) alerts for your organization.

Displays past alerts that have been resolved, including their duration,
resolution time, and related probe information. Useful for post-incident
reviews, SLA reporting, and trend analysis.

Time Range Filtering:
  --since    Show alerts triggered after this time
  --until    Show alerts triggered before this time

Time formats accepted:
  Relative:  24h, 7d, 30d (hours or days ago from now)
  Absolute:  RFC3339 format (e.g., 2024-01-15T10:30:00Z)

Examples:
  # Show alerts from the last 24 hours
  stackeye alert history --since 24h

  # Show alerts from the last 7 days
  stackeye alert history --since 7d

  # Show alerts for a specific probe
  stackeye alert history --probe abc123-def456-...

  # Show alerts within a date range
  stackeye alert history --since 2024-01-01T00:00:00Z --until 2024-01-31T23:59:59Z

  # Output as JSON for scripting
  stackeye alert history --since 7d -o json

  # Paginate through results
  stackeye alert history --since 30d --page 2 --limit 50`,
		Aliases: []string{"hist"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAlertHistory(cmd.Context(), flags)
		},
	}

	// Define command-specific flags
	cmd.Flags().StringVar(&flags.probeID, "probe", "", "filter by probe ID")
	cmd.Flags().StringVar(&flags.since, "since", "", "show alerts triggered after this time (e.g., 24h, 7d, or RFC3339)")
	cmd.Flags().StringVar(&flags.until, "until", "", "show alerts triggered before this time (e.g., 24h, 7d, or RFC3339)")
	cmd.Flags().IntVar(&flags.page, "page", 1, "page number for pagination")
	cmd.Flags().IntVar(&flags.limit, "limit", 20, "results per page (max: 100)")

	return cmd
}

// parseTimeFlag parses a time flag value into a time.Time pointer.
// Accepts relative durations (e.g., "24h", "7d") or RFC3339 timestamps.
func parseTimeFlag(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}

	// Try relative duration first (e.g., "24h", "7d", "30d")
	if daysStr, found := strings.CutSuffix(value, "d"); found {
		// Parse days manually since time.ParseDuration doesn't support "d"
		var days int
		_, err := fmt.Sscanf(daysStr, "%d", &days)
		if err != nil || days < 0 {
			return nil, fmt.Errorf("invalid duration %q: days must be a non-negative integer", value)
		}
		duration := time.Duration(days) * 24 * time.Hour
		t := time.Now().Add(-duration)
		return &t, nil
	}

	if strings.HasSuffix(value, "h") {
		duration, err := time.ParseDuration(value)
		if err != nil {
			return nil, fmt.Errorf("invalid duration %q: %w", value, err)
		}
		if duration < 0 {
			return nil, fmt.Errorf("invalid duration %q: must be positive", value)
		}
		t := time.Now().Add(-duration)
		return &t, nil
	}

	// Try RFC3339 format
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, fmt.Errorf("invalid time %q: use relative (24h, 7d) or RFC3339 format", value)
	}

	return &t, nil
}

// runAlertHistory executes the alert history command logic.
func runAlertHistory(ctx context.Context, flags *alertHistoryFlags) error {
	// Validate all flags before making any API calls
	if flags.limit < 1 || flags.limit > 100 {
		return fmt.Errorf("invalid limit %d: must be between 1 and 100", flags.limit)
	}

	if flags.page < 1 {
		return fmt.Errorf("invalid page %d: must be at least 1", flags.page)
	}

	var probeID *uuid.UUID
	if flags.probeID != "" {
		parsed, err := uuid.Parse(flags.probeID)
		if err != nil {
			return fmt.Errorf("invalid probe ID %q: %w", flags.probeID, err)
		}
		probeID = &parsed
	}

	sinceTime, err := parseTimeFlag(flags.since)
	if err != nil {
		return fmt.Errorf("invalid --since flag: %w", err)
	}

	untilTime, err := parseTimeFlag(flags.until)
	if err != nil {
		return fmt.Errorf("invalid --until flag: %w", err)
	}

	// Validate time range if both are provided
	if sinceTime != nil && untilTime != nil {
		if sinceTime.After(*untilTime) {
			return fmt.Errorf("invalid time range: --since (%s) must be before --until (%s)",
				sinceTime.Format(time.RFC3339), untilTime.Format(time.RFC3339))
		}
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
		Limit:   flags.limit,
		Offset:  offset,
		Status:  client.AlertStatusResolved, // History shows resolved alerts
		ProbeID: probeID,
		From:    sinceTime,
		To:      untilTime,
	}

	// Call SDK to list alerts with timeout
	reqCtx, cancel := context.WithTimeout(ctx, alertHistoryTimeout)
	defer cancel()

	result, err := client.ListAlerts(reqCtx, apiClient, opts)
	if err != nil {
		return fmt.Errorf("failed to fetch alert history: %w", err)
	}

	// Handle empty results
	if len(result.Alerts) == 0 {
		return output.PrintEmpty("No historical alerts found")
	}

	// Print the alerts using the configured output format
	return output.Print(result.Alerts)
}
