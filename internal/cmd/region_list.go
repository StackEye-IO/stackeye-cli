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

// regionListTimeout is the maximum time to wait for the API response.
const regionListTimeout = 30 * time.Second

// NewRegionListCmd creates and returns the region list subcommand.
func NewRegionListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all available monitoring regions",
		Long: `List all available monitoring regions for StackEye probes.

Displays all regions where StackEye runs probe checks, organized by continent.
Each region has a unique code (e.g., nyc3, fra1) that can be used when creating
or updating probes.

Region Information:
  CODE          Unique region identifier (e.g., nyc3)
  NAME          Full region name (e.g., New York 3)
  DISPLAY       Short display name for UI (e.g., New York)
  COUNTRY       ISO 3166-1 alpha-2 country code (e.g., US)
  CONTINENT     Geographic grouping (e.g., north_america, europe)

Requires authentication via 'stackeye login' or API key.

Examples:
  # List all regions
  stackeye region list

  # Output as JSON for scripting
  stackeye region list -o json

  # Output as YAML
  stackeye region list -o yaml

  # Wide output with continent information
  stackeye region list -o wide`,
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRegionList(cmd.Context())
		},
	}

	return cmd
}

// runRegionList executes the region list command logic.
func runRegionList(ctx context.Context) error {
	// Get API client (regions endpoint is public, but we still use the client)
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Call SDK to list regions with timeout
	reqCtx, cancel := context.WithTimeout(ctx, regionListTimeout)
	defer cancel()

	result, err := client.ListRegions(reqCtx, apiClient)
	if err != nil {
		return fmt.Errorf("failed to list regions: %w", err)
	}

	// Handle empty results
	totalRegions := 0
	for _, regions := range result.Data {
		totalRegions += len(regions)
	}
	if totalRegions == 0 {
		return output.PrintEmpty("No regions available.")
	}

	// Print the regions using the configured output format
	return output.PrintRegions(result.Data)
}
