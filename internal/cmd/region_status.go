// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
)

// regionStatusTimeout is the maximum time to wait for all API responses.
const regionStatusTimeout = 60 * time.Second

// regionStatusFlags holds the flag values for the region status command.
type regionStatusFlags struct {
	region string
}

// NewRegionStatusCmd creates and returns the region status subcommand.
func NewRegionStatusCmd() *cobra.Command {
	flags := &regionStatusFlags{}

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show health status of monitoring regions",
		Long: `Show the operational and health status of monitoring regions.

Displays the current status and health indicators for all StackEye monitoring
regions. This helps you identify which regions are fully operational before
selecting them for your probes.

Status Values:
  active        Region is fully operational
  maintenance   Region is under scheduled maintenance
  disabled      Region is temporarily disabled

Health Indicators:
  healthy       All systems operating normally
  warning       Minor issues detected, monitoring continues
  degraded      Significant issues, results may be affected
  unknown       Status cannot be determined

Use the --region flag to check status of a specific region.

Examples:
  # Show status of all regions
  stackeye region status

  # Check status of a specific region
  stackeye region status --region nyc3

  # Output as JSON for scripting
  stackeye region status -o json

  # Wide output with maintenance details
  stackeye region status -o wide`,
		Aliases: []string{"health"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRegionStatus(cmd.Context(), flags)
		},
	}

	// Define command-specific flags
	cmd.Flags().StringVar(&flags.region, "region", "", "specific region ID to check (e.g., nyc3)")

	return cmd
}

// runRegionStatus executes the region status command logic.
func runRegionStatus(ctx context.Context, flags *regionStatusFlags) error {
	// Get API client (regions endpoint is public, but we still use the client)
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Create context with timeout
	reqCtx, cancel := context.WithTimeout(ctx, regionStatusTimeout)
	defer cancel()

	// If a specific region is requested, just fetch that one
	if flags.region != "" {
		return runSingleRegionStatus(reqCtx, apiClient, flags.region)
	}

	// Otherwise, fetch all regions and their statuses
	return runAllRegionStatuses(reqCtx, apiClient)
}

// runSingleRegionStatus fetches and displays status for a single region.
func runSingleRegionStatus(ctx context.Context, apiClient *client.Client, regionID string) error {
	result, err := client.GetRegionStatus(ctx, apiClient, regionID)
	if err != nil {
		return fmt.Errorf("failed to get status for region %q: %w", regionID, err)
	}

	return output.PrintRegionStatus(result.Data)
}

// runAllRegionStatuses fetches and displays status for all regions.
func runAllRegionStatuses(ctx context.Context, apiClient *client.Client) error {
	// First, get the list of all regions
	regions, err := client.GetAllRegionsFlat(ctx, apiClient)
	if err != nil {
		return fmt.Errorf("failed to list regions: %w", err)
	}

	if len(regions) == 0 {
		return output.PrintEmpty("No regions available.")
	}

	// Fetch status for each region concurrently
	statuses, err := fetchAllRegionStatuses(ctx, apiClient, regions)
	if err != nil {
		return err
	}

	if len(statuses) == 0 {
		return output.PrintEmpty("No region status information available.")
	}

	return output.PrintRegionStatuses(statuses)
}

// fetchAllRegionStatuses fetches status for all regions concurrently.
// It returns partial results if some regions fail to respond.
func fetchAllRegionStatuses(ctx context.Context, apiClient *client.Client, regions []client.Region) ([]client.RegionStatus, error) {
	type statusResult struct {
		status client.RegionStatus
		err    error
	}

	// Create channels for results
	results := make(chan statusResult, len(regions))
	var wg sync.WaitGroup

	// Fetch each region's status concurrently
	for _, region := range regions {
		wg.Add(1)
		go func(regionID string) {
			defer wg.Done()

			resp, err := client.GetRegionStatus(ctx, apiClient, regionID)
			if err != nil {
				results <- statusResult{err: fmt.Errorf("region %s: %w", regionID, err)}
				return
			}
			results <- statusResult{status: resp.Data}
		}(region.ID)
	}

	// Close results channel when all goroutines complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	statuses := make([]client.RegionStatus, 0, len(regions))
	var errors []error

	for result := range results {
		if result.err != nil {
			errors = append(errors, result.err)
			continue
		}
		statuses = append(statuses, result.status)
	}

	// If all requests failed, return an error
	if len(statuses) == 0 && len(errors) > 0 {
		return nil, fmt.Errorf("failed to get status for any region: %v", errors[0])
	}

	return statuses, nil
}
