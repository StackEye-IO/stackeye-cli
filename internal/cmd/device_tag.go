// Package cmd implements the CLI commands for StackEye.
// Task stackeye-5859.
package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/dryrun"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
)

// deviceTagTimeout is the maximum time to wait for each API response.
// Task stackeye-5859.
const deviceTagTimeout = 30 * time.Second

// NewDeviceTagCmd creates and returns the device tag subcommand.
// Task stackeye-5859.
func NewDeviceTagCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tag <device-id> <labels...>",
		Short: "Attach tags to a device",
		Long: `Attach one or more tags to a device.

Attaching a tag replaces tag_value if the key is already present on the
device. Idempotent on (device, tag key).

Tags can be specified in two formats:
  key=value    Key-value tag (e.g., env=production)
  key          Key-only tag with no value (e.g., critical)

The device can be specified by UUID or by exact name match. If the name
matches more than one device, use the UUID instead.

Examples:
  # Attach environment and tier tags
  stackeye device tag web-01 env=production tier=backend

  # Attach a key-only tag (no value)
  stackeye device tag web-01 critical

  # Use device UUID
  stackeye device tag 550e8400-e29b-41d4-a716-446655440000 env=dev

See also:
  stackeye device tag list    - List a device's tags
  stackeye device untag       - Remove tags from a device`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeviceTag(cmd.Context(), args[0], args[1:])
		},
	}

	cmd.AddCommand(NewDeviceTagListCmd())

	return cmd
}

// runDeviceTag executes the device tag command logic.
// Task stackeye-5859.
func runDeviceTag(ctx context.Context, deviceIDArg string, labelArgs []string) error {
	labels, err := parseLabelArgs(labelArgs)
	if err != nil {
		return err
	}

	if GetDryRun() {
		labelStrs := make([]string, len(labels))
		for i, l := range labels {
			if l.Value != nil {
				labelStrs[i] = fmt.Sprintf("%s=%s", l.Key, *l.Value)
			} else {
				labelStrs[i] = l.Key
			}
		}
		dryrun.PrintAction("attach tags to", "device",
			"Device", deviceIDArg,
			"Tags", strings.Join(labelStrs, ", "),
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

	for _, label := range labels {
		reqCtx, cancel := context.WithTimeout(ctx, deviceTagTimeout)
		err := client.AttachDeviceTag(reqCtx, apiClient, deviceID, label.Key, label.Value)
		cancel()
		if err != nil {
			return fmt.Errorf("failed to attach tag %q: %w", label.Key, err)
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
