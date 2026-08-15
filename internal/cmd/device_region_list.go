// Package cmd implements the CLI commands for StackEye.
// Task stackeye-5859.
package cmd

import (
	"context"
	"fmt"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
)

// NewDeviceRegionListCmd creates and returns the device region list subcommand.
// Task stackeye-5859.
func NewDeviceRegionListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list <device-id>",
		Short:   "List a device's region assignments",
		Aliases: []string{"ls"},
		Long: `List every region a device is explicitly assigned to, ordered by
region ID.

The device can be specified by UUID or by exact name match.

Examples:
  # List a device's region assignments
  stackeye device region list web-01

  # Output as JSON for scripting
  stackeye device region list web-01 -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeviceRegionList(cmd.Context(), args[0])
		},
	}

	return cmd
}

// runDeviceRegionList executes the device region list command logic.
// Task stackeye-5859.
func runDeviceRegionList(ctx context.Context, deviceIDArg string) error {
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	deviceID, err := ResolveDeviceID(ctx, apiClient, deviceIDArg)
	if err != nil {
		return err
	}

	reqCtx, cancel := context.WithTimeout(ctx, deviceRegionTimeout)
	defer cancel()

	regions, err := client.ListDeviceRegions(reqCtx, apiClient, deviceID)
	if err != nil {
		return fmt.Errorf("failed to list device region assignments: %w", err)
	}

	return output.PrintDeviceRegions(regions)
}
