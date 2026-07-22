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

// enrollmentKeyListTimeout is the maximum time to wait for the API response.
const enrollmentKeyListTimeout = 30 * time.Second

// NewEnrollmentKeyListCmd creates and returns the enrollment-key list subcommand.
func NewEnrollmentKeyListCmd() *cobra.Command {
	var limit int
	var offset int

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List Station enrollment keys for the organization",
		Aliases: []string{"ls"},
		Long: `List all Station enrollment keys for your organization.

Displays each key's name, prefix (for identification), mode (class),
capabilities, usage, and expiry. Full plaintext key values are never
displayed after creation — only the key prefix is shown.

Examples:
  # List all enrollment keys
  stackeye enrollment-key list

  # List as JSON for scripting
  stackeye enrollment-key list -o json

  # Paginate
  stackeye enrollment-key list --limit 20 --offset 20`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEnrollmentKeyList(cmd.Context(), limit, offset)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 0, "items per page (1-100, default 20)")
	cmd.Flags().IntVar(&offset, "offset", 0, "offset for pagination (0-based)")

	return cmd
}

// runEnrollmentKeyList executes the enrollment-key list command logic.
func runEnrollmentKeyList(ctx context.Context, limit, offset int) error {
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, enrollmentKeyListTimeout)
	defer cancel()

	result, err := client.ListEnrollmentKeys(reqCtx, apiClient, client.ListEnrollmentKeysParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return fmt.Errorf("failed to list enrollment keys: %w", err)
	}

	if len(result.EnrollmentKeys) == 0 {
		return output.PrintEmpty("No enrollment keys found. Create one with 'stackeye enrollment-key create'")
	}

	return output.PrintEnrollmentKeys(result.EnrollmentKeys)
}
