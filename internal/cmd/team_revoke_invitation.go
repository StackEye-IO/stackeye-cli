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
	"github.com/StackEye-IO/stackeye-cli/internal/dryrun"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
)

// teamRevokeInvitationTimeout is the maximum time to wait for API responses.
const teamRevokeInvitationTimeout = 30 * time.Second

// teamRevokeInvitationFlags holds the flag values for the team revoke-invitation command.
type teamRevokeInvitationFlags struct {
	id    string
	email string
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

Identify the invitation using either --id or --email (not both).
Use 'stackeye team invitations -o wide' to find invitation IDs.

Required (one of):
  --id      The invitation ID to revoke
  --email   The email address of the pending invitation

Optional Flags:
  --force   Skip the confirmation prompt

Examples:
  # Revoke an invitation by ID
  stackeye team revoke-invitation --id abc123def456

  # Revoke an invitation by email
  stackeye team revoke-invitation --email {invitee_email}

  # Skip confirmation prompt
  stackeye team revoke-invitation --id abc123def456 --force

  # Output result as JSON
  stackeye team revoke-invitation --email {invitee_email} --force -o json`,
		Aliases: []string{"revoke-invite", "cancel-invitation", "cancel-invite"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTeamRevokeInvitation(cmd.Context(), flags)
		},
	}

	// Flags
	cmd.Flags().StringVar(&flags.id, "id", "", "invitation ID (use 'team invitations -o wide' to find)")
	cmd.Flags().StringVar(&flags.email, "email", "", "email address of the pending invitation")
	cmd.Flags().BoolVar(&flags.force, "force", false, "skip confirmation prompt")

	return cmd
}

// validateRevokeInvitationFlags validates all flag values before making API calls.
// Returns an error if any flag value is invalid.
func validateRevokeInvitationFlags(flags *teamRevokeInvitationFlags) error {
	if flags.id == "" && flags.email == "" {
		return fmt.Errorf("either --id or --email is required")
	}
	if flags.id != "" && flags.email != "" {
		return fmt.Errorf("cannot specify both --id and --email")
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

	// Dry-run check: after validation, before API calls
	if GetDryRun() {
		details := []string{}
		if flags.id != "" {
			details = append(details, "Invitation ID", flags.id)
		}
		if flags.email != "" {
			details = append(details, "Email", flags.email)
		}
		dryrun.PrintAction("revoke", "invitation", details...)
		return nil
	}

	// Get authenticated API client (after validation passes)
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Create context with timeout
	reqCtx, cancel := context.WithTimeout(ctx, teamRevokeInvitationTimeout)
	defer cancel()

	// Look up the invitation - required if email provided, optional for ID
	var invitationID string
	var invitationEmail string

	invitations, err := client.ListInvitations(reqCtx, apiClient)
	if err != nil {
		// If looking up by email, we need the invitation list
		if flags.email != "" {
			return fmt.Errorf("failed to list invitations: %w", err)
		}
		// If looking up by ID, we can proceed without the email
		invitationID = flags.id
	} else {
		// Search invitations for matching ID or email
		for _, inv := range invitations.Invitations {
			if flags.email != "" && strings.EqualFold(inv.Email, flags.email) {
				invitationID = inv.ID
				invitationEmail = inv.Email
				break
			}
			if flags.id != "" && inv.ID == flags.id {
				invitationID = inv.ID
				invitationEmail = inv.Email
				break
			}
		}

		// Handle email lookup not found
		if flags.email != "" && invitationID == "" {
			return fmt.Errorf("no pending invitation found for email %q", flags.email)
		}

		// If ID was provided but not found in list, still use it (might be valid)
		if flags.id != "" && invitationID == "" {
			invitationID = flags.id
		}
	}

	// Ask for confirmation unless --force or --no-input is specified
	if !flags.force && !GetNoInput() {
		if !confirmRevocation(invitationID, invitationEmail) {
			fmt.Println("Operation cancelled.")
			return nil
		}
	}

	// Call SDK to revoke invitation
	if err := client.RevokeInvitation(reqCtx, apiClient, invitationID); err != nil {
		return fmt.Errorf("failed to revoke invitation: %w", err)
	}

	// Print the success message
	return output.PrintInvitationRevoked(invitationID, invitationEmail)
}
