// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/dryrun"
	cliinteractive "github.com/StackEye-IO/stackeye-cli/internal/interactive"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// apiKeyDeleteTimeout is the maximum time to wait for a single API response.
const apiKeyDeleteTimeout = 30 * time.Second

// apiKeyDeleteFlags holds the flag values for the api-key delete command.
type apiKeyDeleteFlags struct {
	yes bool // Skip confirmation prompt
}

// NewAPIKeyDeleteCmd creates and returns the api-key delete subcommand.
func NewAPIKeyDeleteCmd() *cobra.Command {
	flags := &apiKeyDeleteFlags{}

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an API key",
		Long: `Delete an API key by its ID.

This permanently revokes the API key. Any applications or scripts using this key
will no longer be able to authenticate with the StackEye API.

IMPORTANT: This action cannot be undone. Once deleted, the key cannot be recovered
and a new key must be created if access is still needed.

By default, the command will prompt for confirmation before deleting. Use --yes
to skip the confirmation prompt for scripting or automation.

Examples:
  # Delete an API key (with confirmation)
  stackeye api-key delete 550e8400-e29b-41d4-a716-446655440000

  # Delete an API key without confirmation
  stackeye api-key delete 550e8400-e29b-41d4-a716-446655440000 --yes

  # Short form
  stackeye api-key delete 550e8400-e29b-41d4-a716-446655440000 -y`,
		Aliases: []string{"rm", "remove", "revoke"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAPIKeyDelete(cmd.Context(), args[0], flags)
		},
	}

	cmd.Flags().BoolVarP(&flags.yes, "yes", "y", false, "skip confirmation prompt")

	return cmd
}

// runAPIKeyDelete executes the api-key delete command logic.
func runAPIKeyDelete(ctx context.Context, idArg string, flags *apiKeyDeleteFlags) error {
	// Parse and validate UUID
	keyID, err := uuid.Parse(idArg)
	if err != nil {
		return fmt.Errorf("invalid API key ID %q: must be a valid UUID", idArg)
	}

	// Dry-run check: after validation, before API calls
	if GetDryRun() {
		dryrun.PrintAction("delete", "API key",
			"ID", keyID.String(),
		)
		return nil
	}

	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Fetch the API key to verify it exists and get its name for confirmation
	getCtx, cancelGet := context.WithTimeout(ctx, apiKeyDeleteTimeout)
	apiKey, err := client.GetAPIKey(getCtx, apiClient, keyID)
	cancelGet()
	if err != nil {
		return fmt.Errorf("failed to get API key: %w", err)
	}

	// Defensive check for nil response
	if apiKey == nil {
		return fmt.Errorf("API key %s not found", keyID)
	}

	// Prompt for confirmation unless --yes flag is set or --no-input is enabled
	message := fmt.Sprintf("Are you sure you want to delete API key %q (prefix: %s)?",
		apiKey.Name, apiKey.KeyPrefix)

	confirmed, err := cliinteractive.Confirm(message, cliinteractive.WithYesFlag(flags.yes))
	if err != nil {
		return err
	}
	if !confirmed {
		fmt.Println("Delete cancelled.")
		return nil
	}

	// Delete the API key
	deleteCtx, cancelDelete := context.WithTimeout(ctx, apiKeyDeleteTimeout)
	err = client.DeleteAPIKey(deleteCtx, apiClient, keyID)
	cancelDelete()

	if err != nil {
		return fmt.Errorf("failed to delete API key: %w", err)
	}

	fmt.Printf("Deleted API key %q (%s)\n", apiKey.Name, keyID)
	return nil
}
