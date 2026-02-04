// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/dryrun"
	clierrors "github.com/StackEye-IO/stackeye-cli/internal/errors"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
)

// teamUpdateRoleTimeout is the maximum time to wait for API responses.
const teamUpdateRoleTimeout = 30 * time.Second

// teamUpdateRoleFlags holds the flag values for the team update-role command.
type teamUpdateRoleFlags struct {
	memberID uint
	email    string
	role     string
}

// NewTeamUpdateRoleCmd creates and returns the team update-role subcommand.
func NewTeamUpdateRoleCmd() *cobra.Command {
	flags := &teamUpdateRoleFlags{}

	cmd := &cobra.Command{
		Use:   "update-role",
		Short: "Update a team member's role",
		Long: `Update the role of an existing team member in your organization.

Identify the member by either their numeric ID (use 'stackeye team list -o wide'
to find IDs) or their email address.

Required Flags:
  --member-id OR --email   Identify the member to update (one required)
  --role                   New role to assign (required)

Roles:
  owner    Full control including billing, settings, and member management
  admin    Manage team members and all organization resources
  member   Create and manage probes, alerts, and channels
  viewer   Read-only access to all resources

Examples:
  # Update role by member ID
  stackeye team update-role --member-id 42 --role admin

  # Update role by email address
  stackeye team update-role --email {user_email} --role viewer

  # Output result as JSON
  stackeye team update-role --member-id 42 --role member -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTeamUpdateRole(cmd.Context(), flags)
		},
	}

	// Flags for member identification (one required)
	cmd.Flags().UintVar(&flags.memberID, "member-id", 0, "member ID (use 'team list -o wide' to find)")
	cmd.Flags().StringVar(&flags.email, "email", "", "member email address")

	// Required role flag
	cmd.Flags().StringVar(&flags.role, "role", "", "new role: owner, admin, member, viewer (required)")
	if err := cmd.MarkFlagRequired("role"); err != nil {
		panic(fmt.Sprintf("failed to mark role flag as required: %v", err))
	}

	return cmd
}

// validateTeamUpdateRoleFlags validates all flag values before making API calls.
// Returns an error if any flag value is invalid.
func validateTeamUpdateRoleFlags(flags *teamUpdateRoleFlags) error {
	// Validate that either member-id or email is provided (but not both)
	if flags.memberID == 0 && flags.email == "" {
		return fmt.Errorf("either --member-id or --email is required")
	}
	if flags.memberID != 0 && flags.email != "" {
		return fmt.Errorf("specify either --member-id or --email, not both")
	}

	// Validate role
	if flags.role == "" {
		return fmt.Errorf("role is required")
	}
	role := strings.ToLower(flags.role)
	if !slices.Contains(validTeamRoles, role) {
		return clierrors.InvalidValueError("--role", flags.role, clierrors.ValidTeamRoles)
	}

	return nil
}

// resolveMemberID resolves the member ID from either the direct ID or email lookup.
// If memberID is provided, it is returned directly.
// If email is provided, ListMembers is called to find the matching member.
func resolveMemberID(ctx context.Context, apiClient *client.Client, memberID uint, email string) (uint, error) {
	// Direct ID provided
	if memberID != 0 {
		return memberID, nil
	}

	// Look up by email - fetch all members and find match
	// Note: API does not support filtering by email, so we fetch and search locally
	result, err := client.ListMembers(ctx, apiClient, &client.ListMembersOptions{
		Limit: 100, // Fetch up to 100 members for lookup
	})
	if err != nil {
		return 0, fmt.Errorf("failed to list team members: %w", err)
	}

	// Search for matching email (case-insensitive)
	emailLower := strings.ToLower(email)
	for _, member := range result.Members {
		if strings.ToLower(member.Email) == emailLower {
			return member.ID, nil
		}
	}

	return 0, fmt.Errorf("no team member found with email %q", email)
}

// runTeamUpdateRole executes the team update-role command logic.
func runTeamUpdateRole(ctx context.Context, flags *teamUpdateRoleFlags) error {
	// Validate all flags before making any API calls
	if err := validateTeamUpdateRoleFlags(flags); err != nil {
		return err
	}

	// Normalize role to lowercase
	role := strings.ToLower(flags.role)

	// Dry-run check: after validation, before API calls
	if GetDryRun() {
		details := []string{
			"Role", role,
		}
		if flags.memberID != 0 {
			details = append(details, "Member ID", fmt.Sprintf("%d", flags.memberID))
		}
		if flags.email != "" {
			details = append(details, "Email", flags.email)
		}
		dryrun.PrintAction("update role for", "team member", details...)
		return nil
	}

	// Get authenticated API client (after validation passes)
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Create context with timeout
	reqCtx, cancel := context.WithTimeout(ctx, teamUpdateRoleTimeout)
	defer cancel()

	// Resolve member ID (either directly or via email lookup)
	memberID, err := resolveMemberID(reqCtx, apiClient, flags.memberID, flags.email)
	if err != nil {
		return err
	}

	// Build update request
	req := &client.UpdateMemberRoleRequest{
		Role: client.TeamRole(role),
	}

	// Call SDK to update member role
	result, err := client.UpdateMemberRole(reqCtx, apiClient, memberID, req)
	if err != nil {
		return fmt.Errorf("failed to update member role: %w", err)
	}

	// Print the success message
	return output.PrintRoleUpdated(result)
}
