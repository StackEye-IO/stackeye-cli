// Package cmd implements the CLI commands for StackEye.
// Task stackeye-5859.
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/dryrun"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
)

// deviceRegionTimeout is the maximum time to wait for each API response.
// Task stackeye-5859.
const deviceRegionTimeout = 30 * time.Second

// NewDeviceRegionAssignCmd creates and returns the device region assign subcommand.
// Task stackeye-5859.
func NewDeviceRegionAssignCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "assign <device-id> <region-id>",
		Short: "Assign a device to a region",
		Long: `Assign a device to a monitoring region.

The region must be a public region or one owned by your organization;
anything else is reported as not found. Idempotent on (device, region).

The device can be specified by UUID or by exact name match.

Examples:
  # Assign a device to a region
  stackeye device region assign web-01 us-east

  # Use device UUID
  stackeye device region assign 550e8400-e29b-41d4-a716-446655440000 eu-west

See also:
  stackeye device region list        - List a device's region assignments
  stackeye device region unassign    - Remove a device's region assignment`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeviceRegionAssign(cmd.Context(), args[0], args[1])
		},
	}

	return cmd
}

// runDeviceRegionAssign executes the device region assign command logic.
// Task stackeye-5859.
func runDeviceRegionAssign(ctx context.Context, deviceIDArg, regionID string) error {
	if GetDryRun() {
		dryrun.PrintAction("assign", "device to region",
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
	err = client.AssignDeviceRegion(reqCtx, apiClient, deviceID, regionID)
	cancel()
	if err != nil {
		return fmt.Errorf("failed to assign region %q: %w", regionID, err)
	}

	reqCtx, cancel = context.WithTimeout(ctx, deviceRegionTimeout)
	defer cancel()

	regions, err := client.ListDeviceRegions(reqCtx, apiClient, deviceID)
	if err != nil {
		return fmt.Errorf("failed to get updated region assignments: %w", err)
	}

	return output.PrintDeviceRegions(regions)
}
