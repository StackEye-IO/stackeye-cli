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

// teamInvitationsTimeout is the maximum time to wait for the API response.
const teamInvitationsTimeout = 30 * time.Second

// NewTeamInvitationsCmd creates and returns the team invitations subcommand.
func NewTeamInvitationsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "invitations",
		Short: "List pending team invitations",
		Long: `List all pending invitations for your organization.

Displays invitation email, assigned role, expiration time, and who sent the invitation.
Invitations expire after 7 days if not accepted.

Roles that can be assigned via invitation:
  admin    Manage team members and all resources
  member   Create and manage probes, alerts, channels
  viewer   Read-only access to all resources

Examples:
  # List all pending invitations
  stackeye team invitations

  # Output as JSON for scripting
  stackeye team invitations -o json

  # Output as YAML
  stackeye team invitations -o yaml

  # Wide output with invitation codes
  stackeye team invitations -o wide`,
		Aliases: []string{"invites", "inv"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTeamInvitations(cmd.Context())
		},
	}

	return cmd
}

// runTeamInvitations executes the team invitations command logic.
func runTeamInvitations(ctx context.Context) error {
	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Call SDK to list invitations with timeout
	reqCtx, cancel := context.WithTimeout(ctx, teamInvitationsTimeout)
	defer cancel()

	result, err := client.ListInvitations(reqCtx, apiClient)
	if err != nil {
		return fmt.Errorf("failed to list invitations: %w", err)
	}

	// Handle empty results
	if len(result.Invitations) == 0 {
		return output.PrintEmpty("No pending invitations. Invite team members with 'stackeye team invite'")
	}

	// Print the invitations using the configured output format
	return output.PrintInvitations(result.Invitations)
}
