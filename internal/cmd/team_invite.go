// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"net/mail"
	"slices"
	"strings"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
)

// teamInviteTimeout is the maximum time to wait for the API response.
const teamInviteTimeout = 30 * time.Second

// validTeamRoles contains the allowed role values for team invitations.
var validTeamRoles = []string{"owner", "admin", "member", "viewer"}

// teamInviteFlags holds the flag values for the team invite command.
type teamInviteFlags struct {
	email string
	role  string
}

// NewTeamInviteCmd creates and returns the team invite subcommand.
func NewTeamInviteCmd() *cobra.Command {
	flags := &teamInviteFlags{}

	cmd := &cobra.Command{
		Use:   "invite",
		Short: "Invite a new team member",
		Long: `Invite a new team member to your organization by email.

An invitation email will be sent to the specified address with a unique
invite code. The invitation expires in 7 days.

Required Flags:
  --email   Email address of the person to invite
  --role    Role to assign (owner, admin, member, viewer)

Roles:
  owner    Full control including billing, settings, and member management
  admin    Manage team members and all organization resources
  member   Create and manage probes, alerts, and channels
  viewer   Read-only access to all resources

Examples:
  # Invite a new team member with admin role
  stackeye team invite --email {user_email} --role admin

  # Invite a viewer
  stackeye team invite --email {user_email} --role viewer

  # Output invitation details as JSON
  stackeye team invite --email {user_email} --role member -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTeamInvite(cmd.Context(), flags)
		},
	}

	// Required flags
	cmd.Flags().StringVar(&flags.email, "email", "", "email address to invite (required)")
	if err := cmd.MarkFlagRequired("email"); err != nil {
		panic(fmt.Sprintf("failed to mark email flag as required: %v", err))
	}

	cmd.Flags().StringVar(&flags.role, "role", "", "role to assign: owner, admin, member, viewer (required)")
	if err := cmd.MarkFlagRequired("role"); err != nil {
		panic(fmt.Sprintf("failed to mark role flag as required: %v", err))
	}

	return cmd
}

// validateTeamInviteFlags validates all flag values before making API calls.
// Returns an error if any flag value is invalid.
func validateTeamInviteFlags(flags *teamInviteFlags) error {
	// Validate email format using RFC 5322 parser
	if flags.email == "" {
		return fmt.Errorf("email is required")
	}
	addr, err := mail.ParseAddress(flags.email)
	if err != nil {
		return fmt.Errorf("invalid email format: %q", flags.email)
	}
	// Additional check: require a dot in the domain for practical email addresses
	// (RFC 5322 allows domains without dots, but real-world emails need TLDs)
	parts := strings.Split(addr.Address, "@")
	if len(parts) != 2 || !strings.Contains(parts[1], ".") {
		return fmt.Errorf("invalid email format: %q", flags.email)
	}

	// Validate role
	if flags.role == "" {
		return fmt.Errorf("role is required")
	}
	role := strings.ToLower(flags.role)
	if !slices.Contains(validTeamRoles, role) {
		return fmt.Errorf("invalid role %q: must be one of %v", flags.role, validTeamRoles)
	}

	return nil
}

// runTeamInvite executes the team invite command logic.
func runTeamInvite(ctx context.Context, flags *teamInviteFlags) error {
	// Validate all flags before making any API calls
	if err := validateTeamInviteFlags(flags); err != nil {
		return err
	}

	// Normalize role to lowercase
	role := strings.ToLower(flags.role)

	// Get authenticated API client (after validation passes)
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Build invite request
	req := &client.InviteMemberRequest{
		Email: flags.email,
		Role:  client.TeamRole(role),
	}

	// Call SDK to invite team member with timeout
	reqCtx, cancel := context.WithTimeout(ctx, teamInviteTimeout)
	defer cancel()

	invitation, err := client.InviteMember(reqCtx, apiClient, req)
	if err != nil {
		return fmt.Errorf("failed to invite team member: %w", err)
	}

	// Print the invitation details
	return output.PrintInvitationCreated(invitation)
}
