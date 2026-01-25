// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
)

// teamRevokeInvitationTimeout is the maximum time to wait for API responses.
const teamRevokeInvitationTimeout = 30 * time.Second

// teamRevokeInvitationFlags holds the flag values for the team revoke-invitation command.
type teamRevokeInvitationFlags struct {
	id    string
	force bool
}

// NewTeamRevokeInvitationCmd creates and returns the team revoke-invitation subcommand.
func NewTeamRevokeInvitationCmd() *cobra.Command {
	flags := &teamRevokeInvitationFlags{}

	cmd := &cobra.Command{
		Use:   "revoke-invitation",
		Short: "Revoke a pending team invitation",
		Long: `Revoke a pending invitation to your organization.

This permanently cancels the invitation. The recipient will no longer be able
to use the invite code to join your organization.

Use 'stackeye team invitations -o wide' to find invitation IDs.

Required Flags:
  --id   The invitation ID to revoke

Optional Flags:
  --force   Skip the confirmation prompt

Examples:
  # Revoke an invitation by ID
  stackeye team revoke-invitation --id abc123def456

  # Skip confirmation prompt
  stackeye team revoke-invitation --id abc123def456 --force

  # Output result as JSON
  stackeye team revoke-invitation --id abc123def456 --force -o json`,
		Aliases: []string{"revoke-invite", "cancel-invitation", "cancel-invite"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTeamRevokeInvitation(cmd.Context(), flags)
		},
	}

	// Flags
	cmd.Flags().StringVar(&flags.id, "id", "", "invitation ID (use 'team invitations -o wide' to find)")
	cmd.Flags().BoolVar(&flags.force, "force", false, "skip confirmation prompt")

	return cmd
}

// validateRevokeInvitationFlags validates all flag values before making API calls.
// Returns an error if any flag value is invalid.
func validateRevokeInvitationFlags(flags *teamRevokeInvitationFlags) error {
	if flags.id == "" {
		return fmt.Errorf("--id is required")
	}
	return nil
}

// confirmRevocation prompts the user to confirm the invitation revocation.
// Returns true if confirmed, false otherwise.
func confirmRevocation(invitationID, email string) bool {
	if email != "" {
		fmt.Printf("Are you sure you want to revoke the invitation for %q? [y/N]: ", email)
	} else {
		fmt.Printf("Are you sure you want to revoke invitation %q? [y/N]: ", invitationID)
	}
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

// runTeamRevokeInvitation executes the team revoke-invitation command logic.
func runTeamRevokeInvitation(ctx context.Context, flags *teamRevokeInvitationFlags) error {
	// Validate all flags before making any API calls
	if err := validateRevokeInvitationFlags(flags); err != nil {
		return err
	}

	// Get authenticated API client (after validation passes)
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Create context with timeout
	reqCtx, cancel := context.WithTimeout(ctx, teamRevokeInvitationTimeout)
	defer cancel()

	// Look up the invitation to get email for confirmation message
	var invitationEmail string
	invitations, err := client.ListInvitations(reqCtx, apiClient)
	if err == nil {
		for _, inv := range invitations.Invitations {
			if inv.ID == flags.id {
				invitationEmail = inv.Email
				break
			}
		}
	}

	// Ask for confirmation unless --force or --no-input is specified
	if !flags.force && !GetNoInput() {
		if !confirmRevocation(flags.id, invitationEmail) {
			fmt.Println("Operation cancelled.")
			return nil
		}
	}

	// Call SDK to revoke invitation
	if err := client.RevokeInvitation(reqCtx, apiClient, flags.id); err != nil {
		return fmt.Errorf("failed to revoke invitation: %w", err)
	}

	// Print the success message
	return output.PrintInvitationRevoked(flags.id, invitationEmail)
}
