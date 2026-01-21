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

// orgListTimeout is the maximum time to wait for the API response.
const orgListTimeout = 30 * time.Second

// NewOrgListCmd creates and returns the org list subcommand.
func NewOrgListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all organizations you belong to",
		Long: `List all organizations you belong to.

Displays organization name, slug, your role, and indicates which organization
is currently active. The active organization is used for all commands unless
overridden with --org flag or X-Organization-ID header.

Roles:
  owner    Full control including billing and deletion
  admin    Manage team members and all resources
  member   Create and manage probes, alerts, channels
  viewer   Read-only access to all resources

Examples:
  # List all organizations
  stackeye org list

  # Output as JSON for scripting
  stackeye org list -o json

  # Output as YAML
  stackeye org list -o yaml`,
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOrgList(cmd.Context())
		},
	}

	return cmd
}

// runOrgList executes the org list command logic.
func runOrgList(ctx context.Context) error {
	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Call SDK to list organizations with timeout
	reqCtx, cancel := context.WithTimeout(ctx, orgListTimeout)
	defer cancel()

	result, err := client.ListOrganizations(reqCtx, apiClient)
	if err != nil {
		return fmt.Errorf("failed to list organizations: %w", err)
	}

	// Handle empty results
	if len(result.Organizations) == 0 {
		return output.PrintEmpty("No organizations found. You may need to create or join an organization.")
	}

	// Print the organizations using the configured output format
	return output.Print(result.Organizations)
}
