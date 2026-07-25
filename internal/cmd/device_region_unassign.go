// Package cmd implements the CLI commands for StackEye.
// Task stackeye-5859.
package cmd

import (
	"context"
	"fmt"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/dryrun"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
)

// NewDeviceRegionUnassignCmd creates and returns the device region unassign subcommand.
// Task stackeye-5859.
func NewDeviceRegionUnassignCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unassign <device-id> <region-id>",
		Short: "Remove a device's region assignment",
		Long: `Remove a device's assignment to a monitoring region.

Idempotent: no error if the device isn't assigned to the region.

The device can be specified by UUID or by exact name match.

Examples:
  # Remove a region assignment
  stackeye device region unassign web-01 us-east

  # Use device UUID
  stackeye device region unassign 550e8400-e29b-41d4-a716-446655440000 us-east`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeviceRegionUnassign(cmd.Context(), args[0], args[1])
		},
	}

	return cmd
}

// runDeviceRegionUnassign executes the device region unassign command logic.
// Task stackeye-5859.
func runDeviceRegionUnassign(ctx context.Context, deviceIDArg, regionID string) error {
	if GetDryRun() {
		dryrun.PrintAction("unassign", "device from region",
			"Device", deviceIDArg,
			"Region", regionID,
		)
		return nil
	}

	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	deviceID, err := ResolveDeviceID(ctx, apiClient, deviceIDArg)
	if err != nil {
		return err
	}

	reqCtx, cancel := context.WithTimeout(ctx, deviceRegionTimeout)
	err = client.UnassignDeviceRegion(reqCtx, apiClient, deviceID, regionID)
	cancel()
	if err != nil {
		return fmt.Errorf("failed to unassign region %q: %w", regionID, err)
	}

	reqCtx, cancel = context.WithTimeout(ctx, deviceRegionTimeout)
	defer cancel()

	regions, err := client.ListDeviceRegions(reqCtx, apiClient, deviceID)
	if err != nil {
		return fmt.Errorf("failed to get updated region assignments: %w", err)
	}

	return output.PrintDeviceRegions(regions)
}
