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

// statusPageListTimeout is the maximum time to wait for the API response.
const statusPageListTimeout = 30 * time.Second

// statusPageListFlags holds the flag values for the status-page list command.
type statusPageListFlags struct {
	page    int
	limit   int
	enabled bool
	public  bool
	search  string
}

// NewStatusPageListCmd creates and returns the status-page list subcommand.
func NewStatusPageListCmd() *cobra.Command {
	flags := &statusPageListFlags{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all status pages",
		Long: `List all public status pages in your organization.

Displays status pages with their name, slug, theme, visibility settings,
probe count, and optional custom domain. Use the filter flags to narrow
results by enabled state, public visibility, or search term.

Status pages provide public-facing displays of system health. Each organization
can create multiple status pages with customizable branding and custom domains.

Status Columns:
  NAME      Status page display name
  SLUG      URL slug (e.g., acme → acme.stackeye.io)
  THEME     Display theme (Light, Dark, System)
  PUBLIC    Public visibility (Yes/No)
  ENABLED   Page is active (Yes/No)
  PROBES    Number of probes on the page

Wide Mode Columns (--output wide):
  DOMAIN    Custom domain if configured
  UPTIME%   Shows uptime percentage (Yes/No)
  ID        Status page ID
  CREATED   Creation date

Examples:
  # List all status pages
  stackeye status-page list

  # List only enabled status pages
  stackeye status-page list --enabled

  # List only public status pages
  stackeye status-page list --public

  # Search for status pages by name or slug
  stackeye status-page list --search "api"

  # Combine filters
  stackeye status-page list --enabled --public

  # Output as JSON for scripting
  stackeye status-page list -o json

  # Wide output with additional columns
  stackeye status-page list -o wide

  # Paginate through results
  stackeye status-page list --page 2 --limit 50`,
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatusPageList(cmd, flags)
		},
	}

	// Define command-specific flags
	cmd.Flags().IntVar(&flags.page, "page", 1, "page number for pagination")
	cmd.Flags().IntVar(&flags.limit, "limit", 20, "results per page (max: 100)")
	cmd.Flags().StringVar(&flags.search, "search", "", "search by name or slug")

	// Boolean filter flags - use cmd.Flags().Changed() to detect if explicitly set
	cmd.Flags().BoolVar(&flags.enabled, "enabled", false, "show only enabled status pages")
	cmd.Flags().BoolVar(&flags.public, "public", false, "show only public status pages")

	return cmd
}

// runStatusPageList executes the status-page list command logic.
func runStatusPageList(cmd *cobra.Command, flags *statusPageListFlags) error {
	ctx := cmd.Context()
	// Validate all flags before making any API calls
	if flags.limit < 1 || flags.limit > 100 {
		return fmt.Errorf("invalid limit %d: must be between 1 and 100", flags.limit)
	}

	if flags.page < 1 {
		return fmt.Errorf("invalid page %d: must be at least 1", flags.page)
	}

	// Get authenticated API client (after validation passes)
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Build list options - SDK uses offset-based pagination
	offset := (flags.page - 1) * flags.limit
	opts := &client.ListStatusPagesOptions{
		Limit:  flags.limit,
		Offset: offset,
	}

	// Add optional filters - only apply if explicitly set by user
	if cmd.Flags().Changed("enabled") {
		opts.Enabled = &flags.enabled
	}
	if cmd.Flags().Changed("public") {
		opts.IsPublic = &flags.public
	}
	if flags.search != "" {
		opts.Search = flags.search
	}

	// Call SDK to list status pages with timeout
	reqCtx, cancel := context.WithTimeout(ctx, statusPageListTimeout)
	defer cancel()

	result, err := client.ListStatusPages(reqCtx, apiClient, opts)
	if err != nil {
		return fmt.Errorf("failed to list status pages: %w", err)
	}

	// Handle empty results
	if len(result.StatusPages) == 0 {
		msg := "No status pages found"
		if flags.search != "" {
			msg = fmt.Sprintf("No status pages found matching '%s'", flags.search)
		}
		return output.PrintEmpty(msg)
	}

	// Print the status pages using the table formatter
	return output.PrintStatusPages(result.StatusPages)
}
