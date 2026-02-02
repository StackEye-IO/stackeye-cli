// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/StackEye-IO/stackeye-go-sdk/interactive"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// statusPageRemoveProbeTimeout is the maximum time to wait for the remove-probe API response.
const statusPageRemoveProbeTimeout = 30 * time.Second

// statusPageRemoveProbeFlags holds the flag values for the status-page remove-probe command.
type statusPageRemoveProbeFlags struct {
	probeID string
	yes     bool // Skip confirmation prompt
}

// NewStatusPageRemoveProbeCmd creates and returns the status-page remove-probe subcommand.
func NewStatusPageRemoveProbeCmd() *cobra.Command {
	flags := &statusPageRemoveProbeFlags{}

	cmd := &cobra.Command{
		Use:   "remove-probe <status-page-id>",
		Short: "Remove a probe from a status page",
		Long: `Remove a probe from a status page.

This command removes an existing probe association from a status page. The probe
itself is not deleted, only its display on the status page is removed.

By default, the command will prompt for confirmation before removing. Use --yes
to skip the confirmation prompt for scripting or automation.

Flags:
  --probe-id    Required. The UUID of the probe to remove.
  --yes, -y     Skip confirmation prompt.

Examples:
  # Remove a probe from status page 123 (with confirmation)
  stackeye status-page remove-probe 123 --probe-id {probe_uuid}

  # Remove a probe without confirmation
  stackeye status-page remove-probe 123 --probe-id {probe_uuid} --yes`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatusPageRemoveProbe(cmd.Context(), args[0], flags)
		},
	}

	cmd.Flags().StringVar(&flags.probeID, "probe-id", "", "probe UUID to remove (required)")
	cmd.Flags().BoolVarP(&flags.yes, "yes", "y", false, "skip confirmation prompt")

	_ = cmd.MarkFlagRequired("probe-id")

	return cmd
}

// runStatusPageRemoveProbe executes the status-page remove-probe command logic.
func runStatusPageRemoveProbe(ctx context.Context, idArg string, flags *statusPageRemoveProbeFlags) error {
	// Parse and validate status page ID
	id, err := strconv.ParseUint(idArg, 10, 64)
	if err != nil || id == 0 {
		return fmt.Errorf("invalid status page ID %q: must be a positive integer", idArg)
	}

	// Validate probe ID is provided and is a valid UUID
	if flags.probeID == "" {
		return fmt.Errorf("--probe-id is required")
	}

	probeUUID, err := uuid.Parse(flags.probeID)
	if err != nil {
		return fmt.Errorf("invalid probe ID %q: must be a valid UUID", flags.probeID)
	}

	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Prompt for confirmation unless --yes flag is set or --no-input is enabled
	if !flags.yes && !GetNoInput() {
		message := fmt.Sprintf("Are you sure you want to remove probe %s from status page %d?", probeUUID.String(), id)

		confirmed, err := interactive.AskConfirm(&interactive.ConfirmPromptOptions{
			Message: message,
			Default: false,
		})
		if err != nil {
			if errors.Is(err, interactive.ErrPromptCancelled) {
				return fmt.Errorf("operation cancelled by user")
			}
			return fmt.Errorf("failed to prompt for confirmation: %w", err)
		}

		if !confirmed {
			fmt.Println("Remove cancelled.")
			return nil
		}
	}

	// Remove the probe from the status page
	removeCtx, cancel := context.WithTimeout(ctx, statusPageRemoveProbeTimeout)
	err = client.RemoveProbeFromStatusPage(removeCtx, apiClient, uint(id), probeUUID)
	cancel()

	if err != nil {
		return fmt.Errorf("failed to remove probe from status page: %w", err)
	}

	// Display success message
	fmt.Printf("Removed probe %s from status page %d\n", probeUUID.String(), id)

	return nil
}
