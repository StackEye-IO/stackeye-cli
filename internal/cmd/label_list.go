// Package cmd implements the CLI commands for StackEye.
// Task #8065
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

// labelListTimeout is the maximum time to wait for the API response.
const labelListTimeout = 30 * time.Second

// NewLabelListCmd creates and returns the label list subcommand.
func NewLabelListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all label keys in your organization",
		Long: `List all probe label keys defined in your organization.

Displays all label keys with their display names, colors, values in use,
and the number of probes using each key.

Output columns:
  KEY            Label key identifier (e.g., env, tier)
  DISPLAY NAME   Human-readable name for the key
  COLOR          Color indicator for UI badge display
  VALUES IN USE  Distinct values assigned to probes (or "key-only" for tags)
  PROBES         Number of probes using this label key

Requires authentication via 'stackeye login' or API key.

Examples:
  # List all label keys
  stackeye label list

  # Output as JSON for scripting
  stackeye label list -o json

  # Output as YAML
  stackeye label list -o yaml`,
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLabelList(cmd.Context())
		},
	}

	return cmd
}

// runLabelList executes the label list command logic.
func runLabelList(ctx context.Context) error {
	// Get API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Call SDK to list label keys with timeout
	reqCtx, cancel := context.WithTimeout(ctx, labelListTimeout)
	defer cancel()

	result, err := client.ListLabelKeys(reqCtx, apiClient)
	if err != nil {
		return fmt.Errorf("failed to list label keys: %w", err)
	}

	// Handle empty results
	if len(result.LabelKeys) == 0 {
		return output.PrintEmpty("No label keys found. Create one with 'stackeye label create'")
	}

	// Print the label keys using the configured output format
	return output.PrintLabelKeys(result.LabelKeys)
}
