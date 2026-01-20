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

// probeListTimeout is the maximum time to wait for the API response.
const probeListTimeout = 30 * time.Second

// probeListFlags holds the flag values for the probe list command.
type probeListFlags struct {
	status string
	page   int
	limit  int
	period string
}

// NewProbeListCmd creates and returns the probe list subcommand.
func NewProbeListCmd() *cobra.Command {
	flags := &probeListFlags{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all monitoring probes",
		Long: `List all monitoring probes in your organization.

Displays probe status, name, target URL, check interval, and last check time.
Results are paginated and can be filtered by status.

Status Values:
  up        Probe is healthy and responding
  down      Probe is failing checks
  degraded  Probe is experiencing intermittent issues
  paused    Monitoring is temporarily disabled
  pending   Probe has not been checked yet

Examples:
  # List all probes
  stackeye probe list

  # List only probes that are down
  stackeye probe list --status down

  # List probes with uptime stats for last 7 days
  stackeye probe list --period 7d

  # Output as JSON for scripting
  stackeye probe list -o json

  # Paginate through results
  stackeye probe list --page 2 --limit 50`,
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProbeList(cmd.Context(), flags)
		},
	}

	// Define command-specific flags
	cmd.Flags().StringVar(&flags.status, "status", "", "filter by status: up, down, degraded, paused, pending")
	cmd.Flags().IntVar(&flags.page, "page", 1, "page number for pagination")
	cmd.Flags().IntVar(&flags.limit, "limit", 20, "results per page (max: 100)")
	cmd.Flags().StringVar(&flags.period, "period", "", "include uptime stats for period: 24h, 7d, 30d")

	return cmd
}

// runProbeList executes the probe list command logic.
func runProbeList(ctx context.Context, flags *probeListFlags) error {
	// Validate all flags before making any API calls
	if flags.limit < 1 || flags.limit > 100 {
		return fmt.Errorf("invalid limit %d: must be between 1 and 100", flags.limit)
	}

	if flags.page < 1 {
		return fmt.Errorf("invalid page %d: must be at least 1", flags.page)
	}

	if flags.period != "" {
		switch flags.period {
		case "24h", "7d", "30d":
			// Valid period
		default:
			return fmt.Errorf("invalid period %q: must be 24h, 7d, or 30d", flags.period)
		}
	}

	var probeStatus client.ProbeStatus
	if flags.status != "" {
		switch flags.status {
		case "up":
			probeStatus = client.ProbeStatusUp
		case "down":
			probeStatus = client.ProbeStatusDown
		case "degraded":
			probeStatus = client.ProbeStatusDegraded
		case "paused":
			probeStatus = client.ProbeStatusPaused
		case "pending":
			probeStatus = client.ProbeStatusPending
		default:
			return fmt.Errorf("invalid status %q: must be up, down, degraded, paused, or pending", flags.status)
		}
	}

	// Get authenticated API client (after validation passes)
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Build list options from validated flags
	opts := &client.ListProbesOptions{
		Page:   flags.page,
		Limit:  flags.limit,
		Period: flags.period,
		Status: probeStatus,
	}

	// Call SDK to list probes with timeout
	reqCtx, cancel := context.WithTimeout(ctx, probeListTimeout)
	defer cancel()

	result, err := client.ListProbes(reqCtx, apiClient, opts)
	if err != nil {
		return fmt.Errorf("failed to list probes: %w", err)
	}

	// Handle empty results
	if len(result.Probes) == 0 {
		return output.PrintEmpty("No probes found. Create one with 'stackeye probe create'")
	}

	// Print the probes using the configured output format
	return output.Print(result.Probes)
}
