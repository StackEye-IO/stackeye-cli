// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/StackEye-IO/stackeye-go-sdk/interactive"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// muteDeleteTimeout is the maximum time to wait for a single delete API response.
const muteDeleteTimeout = 30 * time.Second

// muteDeleteFlags holds the flag values for the mute delete command.
type muteDeleteFlags struct {
	yes bool // Skip confirmation prompt
}

// NewMuteDeleteCmd creates and returns the mute delete subcommand.
func NewMuteDeleteCmd() *cobra.Command {
	flags := &muteDeleteFlags{}

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an alert mute period",
		Long: `Delete an alert mute period by its ID.

This permanently removes the mute configuration. Alerts that were silenced by
this mute will resume normal notification behavior after deletion.

Note: Deleting a mute is different from expiring it. Expiration leaves a record
of the mute in history, while deletion removes it entirely. For audit purposes,
consider using 'stackeye mute expire' instead.

By default, the command will prompt for confirmation before deleting. Use --yes
to skip the confirmation prompt for scripting or automation.

Examples:
  # Delete a mute (with confirmation)
  stackeye mute delete 550e8400-e29b-41d4-a716-446655440000

  # Delete a mute without confirmation
  stackeye mute delete 550e8400-e29b-41d4-a716-446655440000 --yes

  # Short form
  stackeye mute delete 550e8400-e29b-41d4-a716-446655440000 -y`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMuteDelete(cmd.Context(), args[0], flags)
		},
	}

	cmd.Flags().BoolVarP(&flags.yes, "yes", "y", false, "skip confirmation prompt")

	return cmd
}

// runMuteDelete executes the mute delete command logic.
func runMuteDelete(ctx context.Context, idArg string, flags *muteDeleteFlags) error {
	// Parse and validate UUID
	muteID, err := uuid.Parse(idArg)
	if err != nil {
		return fmt.Errorf("invalid mute ID %q: must be a valid UUID", idArg)
	}

	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Fetch mute to check if it exists and get details for display
	getCtx, cancelGet := context.WithTimeout(ctx, muteDeleteTimeout)
	mute, err := client.GetMute(getCtx, apiClient, muteID)
	cancelGet()
	if err != nil {
		return fmt.Errorf("failed to get mute: %w", err)
	}

	// Defensive check for nil mute
	if mute == nil {
		return fmt.Errorf("mute %s not found", muteID)
	}

	// Prompt for confirmation unless --yes flag is set or --no-input is enabled
	if !flags.yes && !GetNoInput() {
		message := fmt.Sprintf("Are you sure you want to delete mute %s (%s scope)?", muteID, mute.ScopeType)

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
			fmt.Println("Delete cancelled.")
			return nil
		}
	}

	// Delete the mute
	deleteCtx, cancelDelete := context.WithTimeout(ctx, muteDeleteTimeout)
	err = client.DeleteMute(deleteCtx, apiClient, muteID)
	cancelDelete()

	if err != nil {
		return fmt.Errorf("failed to delete mute: %w", err)
	}

	fmt.Printf("Deleted mute %s (%s scope)\n", muteID, mute.ScopeType)
	return nil
}
