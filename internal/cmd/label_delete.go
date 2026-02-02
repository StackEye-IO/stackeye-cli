// Package cmd implements the CLI commands for StackEye.
// Task #8067
package cmd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/StackEye-IO/stackeye-go-sdk/interactive"
	"github.com/spf13/cobra"
)

// labelDeleteTimeout is the maximum time to wait for the API response.
const labelDeleteTimeout = 30 * time.Second

// labelDeleteFlags holds the flag values for the label delete command.
type labelDeleteFlags struct {
	yes   bool // Skip confirmation prompt
	force bool // Force deletion without any prompt
}

// NewLabelDeleteCmd creates and returns the label delete subcommand.
func NewLabelDeleteCmd() *cobra.Command {
	flags := &labelDeleteFlags{}

	cmd := &cobra.Command{
		Use:   "delete <key>",
		Short: "Delete a label key",
		Long: `Delete a probe label key from your organization.

This permanently removes the label key and cascades to remove the label from
all probes that have it assigned. This action cannot be undone.

Before deletion, the command shows the number of probes that will be affected.
By default, the command will prompt for confirmation before deleting. Use --yes
to skip the confirmation prompt for scripting or automation.

Requires authentication via 'stackeye login' or API key.

Examples:
  # Delete a label key (with confirmation)
  stackeye label delete env

  # Delete a label key without confirmation
  stackeye label delete env --yes

  # Short form
  stackeye label delete env -y

  # Force deletion (alias for --yes)
  stackeye label delete env --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLabelDelete(cmd.Context(), args[0], flags)
		},
	}

	cmd.Flags().BoolVarP(&flags.yes, "yes", "y", false, "skip confirmation prompt")
	cmd.Flags().BoolVar(&flags.force, "force", false, "force deletion without confirmation (alias for --yes)")

	return cmd
}

// runLabelDelete executes the label delete command logic.
func runLabelDelete(ctx context.Context, key string, flags *labelDeleteFlags) error {
	// Validate key format locally before API call
	if err := validateLabelKey(key); err != nil {
		return err
	}

	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Fetch label key to check if it exists and get probe count for warning
	getCtx, cancelGet := context.WithTimeout(ctx, labelDeleteTimeout)
	result, err := client.GetLabelKey(getCtx, apiClient, key)
	cancelGet()
	if err != nil {
		return fmt.Errorf("failed to get label key: %w", err)
	}

	labelKey := result.LabelKey

	// Warn if label key is used by probes
	if labelKey.ProbeCount > 0 {
		fmt.Printf("Warning: This label key is used by %d probe(s).\n", labelKey.ProbeCount)
		fmt.Println("The label will be removed from all affected probes after deletion.")
	}

	// Prompt for confirmation unless --yes or --force flag is set or --no-input is enabled
	if !flags.yes && !flags.force && !GetNoInput() {
		displayName := key
		if labelKey.DisplayName != nil && *labelKey.DisplayName != "" {
			displayName = fmt.Sprintf("%s (%s)", *labelKey.DisplayName, key)
		}
		message := fmt.Sprintf("Are you sure you want to delete label key %q?", displayName)

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

	// Delete the label key
	deleteCtx, cancelDelete := context.WithTimeout(ctx, labelDeleteTimeout)
	err = client.DeleteLabelKey(deleteCtx, apiClient, key)
	cancelDelete()

	if err != nil {
		return fmt.Errorf("failed to delete label key: %w", err)
	}

	fmt.Printf("Deleted label key %q\n", key)
	return nil
}
