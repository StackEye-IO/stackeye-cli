// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
)

// apiKeyListTimeout is the maximum time to wait for the API response.
const apiKeyListTimeout = 30 * time.Second

// NewAPIKeyListCmd creates and returns the api-key list subcommand.
func NewAPIKeyListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all API keys for the organization",
		Long: `List all API keys for your organization.

Displays key name, prefix (for identification), permissions, and last used date.
API keys are shown without pagination since typical organizations have fewer than 50 keys.

Security Note:
  Full API key values are never displayed - only the key prefix (first 8 characters)
  is shown for identification purposes. This prevents accidental exposure of credentials.

Examples:
  # List all API keys
  stackeye api-key list

  # List as JSON for scripting
  stackeye api-key list -o json

  # List with wide output for more details
  stackeye api-key list -o wide`,
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAPIKeyList(cmd.Context())
		},
	}

	return cmd
}

// runAPIKeyList executes the api-key list command logic.
func runAPIKeyList(ctx context.Context) error {
	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Call SDK to list API keys with timeout
	reqCtx, cancel := context.WithTimeout(ctx, apiKeyListTimeout)
	defer cancel()

	result, err := client.ListAPIKeys(reqCtx, apiClient)
	if err != nil {
		return fmt.Errorf("failed to list API keys: %w", err)
	}

	// Handle empty results
	if len(result.Data) == 0 {
		return output.PrintEmpty("No API keys found. Create one with 'stackeye api-key create'")
	}

	// Print the API keys using the dedicated formatter
	return output.PrintAPIKeys(result.Data)
}
