// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/config"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/spf13/cobra"
)

// orgSwitchTimeout is the maximum time to wait for the API response.
const orgSwitchTimeout = 30 * time.Second

// SwitchResult represents the result of an organization switch.
type SwitchResult struct {
	OrganizationID   string `json:"organization_id"`
	OrganizationName string `json:"organization_name"`
	OrganizationSlug string `json:"organization_slug"`
	Role             string `json:"role"`
	Message          string `json:"message"`
}

// NewOrgSwitchCmd creates and returns the org switch subcommand.
func NewOrgSwitchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "switch <id|slug>",
		Short: "Switch to a different organization",
		Long: `Switch your active organization context.

This command changes which organization is used for subsequent CLI commands.
You can specify the organization by its UUID or URL-friendly slug.

The switch only affects the current CLI context. If you need to work with
multiple organizations simultaneously, consider using different contexts
(see 'stackeye context' commands).

Examples:
  # Switch by UUID
  stackeye org switch 550e8400-e29b-41d4-a716-446655440000

  # Switch by slug
  stackeye org switch acme-corp

  # Verify the switch worked
  stackeye org get`,
		Aliases: []string{"use", "select"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOrgSwitch(cmd.Context(), args[0])
		},
	}

	return cmd
}

// runOrgSwitch executes the org switch command logic.
func runOrgSwitch(ctx context.Context, identifier string) error {
	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Create timeout context
	reqCtx, cancel := context.WithTimeout(ctx, orgSwitchTimeout)
	defer cancel()

	// Find the target organization
	org, err := findOrganization(reqCtx, apiClient, identifier)
	if err != nil {
		return err
	}

	// Load the current config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get the current context
	currentCtx, err := cfg.GetCurrentContext()
	if err != nil {
		return fmt.Errorf("failed to get current context: %w", err)
	}

	// Check if already switched to this org
	if currentCtx.OrganizationID == org.ID {
		return output.Print(SwitchResult{
			OrganizationID:   org.ID,
			OrganizationName: org.Name,
			OrganizationSlug: org.Slug,
			Role:             org.Role,
			Message:          fmt.Sprintf("Already using organization %q", org.Name),
		})
	}

	// Update the context with the new organization
	currentCtx.OrganizationID = org.ID
	currentCtx.OrganizationName = org.Name

	// Save the updated config
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Print the result
	return output.Print(SwitchResult{
		OrganizationID:   org.ID,
		OrganizationName: org.Name,
		OrganizationSlug: org.Slug,
		Role:             org.Role,
		Message:          fmt.Sprintf("Switched to organization %q", org.Name),
	})
}
