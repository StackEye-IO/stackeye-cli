// Package cmd implements the CLI commands for StackEye.
// Task stackeye-5859.
package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/dryrun"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
)

// NewDeviceUntagCmd creates and returns the device untag subcommand.
// Task stackeye-5859.
func NewDeviceUntagCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "untag <device-id> <keys...>",
		Short: "Remove tags from a device",
		Long: `Remove one or more tags from a device by key name.

Tag keys that don't exist on the device are silently ignored (no error).
This operation does not affect other tags on the device.

The device can be specified by UUID or by exact name match.

Examples:
  # Remove a single tag
  stackeye device untag web-01 env

  # Remove multiple tags at once
  stackeye device untag web-01 env tier critical

  # Use device UUID
  stackeye device untag 550e8400-e29b-41d4-a716-446655440000 env`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeviceUntag(cmd.Context(), args[0], args[1:])
		},
	}

	return cmd
}

// runDeviceUntag executes the device untag command logic.
// Task stackeye-5859.
func runDeviceUntag(ctx context.Context, deviceIDArg string, keys []string) error {
	if GetDryRun() {
		dryrun.PrintAction("remove tags from", "device",
			"Device", deviceIDArg,
			"Tag Keys", strings.Join(keys, ", "),
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

	// RemoveDeviceTag is idempotent server-side (204 even if the tag is
	// absent), so no not-found special-casing is needed here — unlike
	// runProbeUnlabel, which has to work around the probe-label endpoint
	// returning 404 for an absent key.
	for _, key := range keys {
		reqCtx, cancel := context.WithTimeout(ctx, deviceTagTimeout)
		err := client.RemoveDeviceTag(reqCtx, apiClient, deviceID, key)
		cancel()
		if err != nil {
			return fmt.Errorf("failed to remove tag %q: %w", key, err)
		}
	}

	reqCtx, cancel := context.WithTimeout(ctx, deviceTagTimeout)
	defer cancel()

	tags, err := client.ListDeviceTags(reqCtx, apiClient, deviceID)
	if err != nil {
		return fmt.Errorf("failed to get updated tags: %w", err)
	}

	return output.PrintDeviceTags(tags)
}
