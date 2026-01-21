// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
)

// statusPageGetTimeout is the maximum time to wait for the API response.
const statusPageGetTimeout = 30 * time.Second

// NewStatusPageGetCmd creates and returns the status-page get subcommand.
func NewStatusPageGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get details of a status page",
		Long: `Get detailed information about a specific status page.

Displays the full status page configuration including name, slug, visibility
settings, branding (logo, favicon, header/footer text), theme, custom domain,
and the list of probes assigned to it.

Status Page Fields:
  NAME                Status page display name
  SLUG                URL slug (e.g., acme → acme.stackeye.io)
  THEME               Display theme (Light, Dark, System)
  PUBLIC              Public visibility (Yes/No)
  ENABLED             Page is active (Yes/No)
  PROBES              Number of probes on the page

Wide Mode Fields (--output wide):
  DOMAIN              Custom domain if configured
  UPTIME%             Shows uptime percentage (Yes/No)
  ID                  Status page ID
  CREATED             Creation date

Examples:
  # Get status page details by ID
  stackeye status-page get 123

  # Output as JSON for scripting
  stackeye status-page get 123 -o json

  # Output as YAML
  stackeye status-page get 123 -o yaml

  # Wide output with additional fields
  stackeye status-page get 123 -o wide`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatusPageGet(cmd.Context(), args[0])
		},
	}

	return cmd
}

// runStatusPageGet executes the status-page get command logic.
func runStatusPageGet(ctx context.Context, idArg string) error {
	// Parse and validate status page ID (uint)
	id, err := strconv.ParseUint(idArg, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid status page ID %q: must be a positive integer", idArg)
	}
	if id == 0 {
		return fmt.Errorf("invalid status page ID: must be greater than 0")
	}

	// Get authenticated API client (after validation passes)
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Call SDK to get status page with timeout
	reqCtx, cancel := context.WithTimeout(ctx, statusPageGetTimeout)
	defer cancel()

	statusPage, err := client.GetStatusPage(reqCtx, apiClient, uint(id))
	if err != nil {
		return fmt.Errorf("failed to get status page: %w", err)
	}

	// Defensive check for nil status page
	if statusPage == nil {
		return fmt.Errorf("status page %d not found", id)
	}

	// Print the status page using the configured output format
	return output.PrintStatusPage(*statusPage)
}
