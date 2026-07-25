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

// NewDeviceTagListCmd creates and returns the device tag list subcommand.
// Task stackeye-5859.
func NewDeviceTagListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list <device-id>",
		Short:   "List a device's tags",
		Aliases: []string{"ls"},
		Long: `List every tag attached to a device, ordered by tag key.

The device can be specified by UUID or by exact name match.

Examples:
  # List a device's tags
  stackeye device tag list web-01

  # Output as JSON for scripting
  stackeye device tag list web-01 -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeviceTagList(cmd.Context(), args[0])
		},
	}

	return cmd
}

// runDeviceTagList executes the device tag list command logic.
// Task stackeye-5859.
func runDeviceTagList(ctx context.Context, deviceIDArg string) error {
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	deviceID, err := ResolveDeviceID(ctx, apiClient, deviceIDArg)
	if err != nil {
		return err
	}

	reqCtx, cancel := context.WithTimeout(ctx, deviceTagTimeout)
	defer cancel()

	tags, err := client.ListDeviceTags(reqCtx, apiClient, deviceID)
	if err != nil {
		return fmt.Errorf("failed to list device tags: %w", err)
	}

	return output.PrintDeviceTags(tags)
}
