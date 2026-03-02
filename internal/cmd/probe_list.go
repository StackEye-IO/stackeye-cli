// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	clierrors "github.com/StackEye-IO/stackeye-cli/internal/errors"
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
	labels string // Task #8070: Comma-separated label filters
	// paginationSet tracks whether user explicitly passed --page/--limit.
	// If not set, list defaults to returning all probes across pages.
	paginationSet bool
}

// NewProbeListCmd creates and returns the probe list subcommand.
func NewProbeListCmd() *cobra.Command {
	flags := &probeListFlags{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all monitoring probes",
		Long: `List all monitoring probes in your organization.

Displays probe status, name, target URL, check interval, and last check time.
By default, returns all probes and can be filtered by status or labels.
Use --page/--limit for manual pagination.

Status Values:
  up        Probe is healthy and responding
  down      Probe is failing checks
  degraded  Probe is experiencing intermittent issues
  paused    Monitoring is temporarily disabled
  pending   Probe has not been checked yet

Label Filters:
  Labels can be specified as comma-separated filters:
  - key=value    Match probes with label key having specific value
  - key          Match probes with label key (any value or no value)

  Multiple labels use AND logic - probes must match ALL specified labels.

Examples:
  # List all probes
  stackeye probe list

  # List only probes that are down
  stackeye probe list -s down

  # List probes with specific labels
  stackeye probe list -l "env=production,tier=web"

  # List probes with a label key (any value)
  stackeye probe list --labels "pci"

  # Combine status and label filters
  stackeye probe list -s up -l "env=production"

  # List probes with uptime stats for last 7 days
  stackeye probe list -p 7d

  # Output as JSON for scripting
  stackeye probe list -o json

  # Paginate through results
  stackeye probe list --page 2 --limit 50`,
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			flags.paginationSet = cmd.Flags().Changed("page") || cmd.Flags().Changed("limit")
			return runProbeList(cmd.Context(), flags)
		},
	}

	// Define command-specific flags
	cmd.Flags().StringVarP(&flags.status, "status", "s", "", "filter by status: up, down, degraded, paused, pending")
	cmd.Flags().IntVar(&flags.page, "page", 1, "page number for pagination")
	cmd.Flags().IntVar(&flags.limit, "limit", 20, "results per page (max: 100)")
	cmd.Flags().StringVarP(&flags.period, "period", "p", "", "include uptime stats for period: 24h, 7d, 30d")
	cmd.Flags().StringVarP(&flags.labels, "labels", "l", "", "filter by labels: key=value,key2=value2 (AND logic)")

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
			return clierrors.InvalidValueError("--period", flags.period, clierrors.ValidPeriods)
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
			return clierrors.InvalidValueError("--status", flags.status, clierrors.ValidProbeStatusFilters)
		}
	}

	// Task #8070: Parse label filters
	labelFilters, err := parseLabelFilters(flags.labels)
	if err != nil {
		return err
	}

	// Get authenticated API client (after validation passes)
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Call SDK to list probes with timeout
	reqCtx, cancel := context.WithTimeout(ctx, probeListTimeout)
	defer cancel()

	var probes []client.Probe
	if flags.paginationSet {
		// Respect explicit pagination flags and fetch only one page.
		opts := &client.ListProbesOptions{
			Page:   flags.page,
			Limit:  flags.limit,
			Period: flags.period,
			Status: probeStatus,
			Labels: labelFilters,
		}
		result, err := client.ListProbes(reqCtx, apiClient, opts)
		if err != nil {
			return fmt.Errorf("failed to list probes: %w", err)
		}
		probes = result.Probes
	} else {
		// Default behavior: fetch all pages so "probe list" actually lists all probes.
		probes, err = fetchAllProbesForList(reqCtx, apiClient, probeStatus, flags.period, labelFilters)
		if err != nil {
			return err
		}
	}

	// Handle empty results
	if len(probes) == 0 {
		if len(labelFilters) > 0 {
			return output.PrintEmpty(fmt.Sprintf("No probes found with labels: %s", flags.labels))
		}
		return output.PrintEmpty("No probes found. Create one with 'stackeye probe create'")
	}

	// Task #8070: Show count message when label filters are applied
	if len(labelFilters) > 0 {
		fmt.Printf("Showing %d probes with labels: %s\n\n", len(probes), flags.labels)
	}

	// Print the probes using the configured output format
	return output.PrintProbes(probes)
}

// fetchAllProbesForList fetches all probes across pages.
func fetchAllProbesForList(
	ctx context.Context,
	apiClient *client.Client,
	status client.ProbeStatus,
	period string,
	labels map[string]string,
) ([]client.Probe, error) {
	var allProbes []client.Probe
	page := 1
	limit := 100

	for {
		opts := &client.ListProbesOptions{
			Page:   page,
			Limit:  limit,
			Period: period,
			Status: status,
			Labels: labels,
		}

		result, err := client.ListProbes(ctx, apiClient, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list probes: %w", err)
		}

		allProbes = append(allProbes, result.Probes...)

		if len(result.Probes) < limit {
			break
		}
		page++
	}

	return allProbes, nil
}

// parseLabelFilters parses a comma-separated label filter string into a map.
// Accepts formats: "key=value" (exact match) or "key" (any value).
// Task #8070
func parseLabelFilters(labelsStr string) (map[string]string, error) {
	if labelsStr == "" {
		return nil, nil
	}

	filters := make(map[string]string)
	parts := strings.Split(labelsStr, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Check for key=value format
		if idx := strings.Index(part, "="); idx > 0 {
			key := part[:idx]
			value := part[idx+1:]

			if err := validateLabelKey(key); err != nil {
				return nil, err
			}
			filters[key] = value
		} else {
			// Key-only format (matches any value)
			if err := validateLabelKey(part); err != nil {
				return nil, err
			}
			filters[part] = "" // Empty string means "any value"
		}
	}

	return filters, nil
}
