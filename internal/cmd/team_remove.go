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

// teamRemoveTimeout is the maximum time to wait for API responses.
const teamRemoveTimeout = 30 * time.Second

// teamRemoveFlags holds the flag values for the team remove command.
type teamRemoveFlags struct {
	memberID uint
	email    string
	force    bool
}

// NewTeamRemoveCmd creates and returns the team remove subcommand.
func NewTeamRemoveCmd() *cobra.Command {
	flags := &teamRemoveFlags{}

	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove a team member",
		Long: `Remove a team member from your organization.

This permanently removes the member's access to your organization. The member
will no longer be able to view probes, alerts, or any other resources.

Identify the member by either their numeric ID (use 'stackeye team list -o wide'
to find IDs) or their email address.

Required Flags:
  --member-id OR --email   Identify the member to remove (one required)

Optional Flags:
  --force   Skip the confirmation prompt

Note: You cannot remove the organization owner or yourself.

Examples:
  # Remove a member by ID
  stackeye team remove --member-id 42

  # Remove a member by email address
  stackeye team remove --email {user_email}

  # Skip confirmation prompt
  stackeye team remove --member-id 42 --force

  # Output result as JSON
  stackeye team remove --member-id 42 --force -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTeamRemove(cmd.Context(), flags)
		},
	}

	// Flags for member identification (one required)
	cmd.Flags().UintVar(&flags.memberID, "member-id", 0, "member ID (use 'team list -o wide' to find)")
	cmd.Flags().StringVar(&flags.email, "email", "", "member email address")

	// Optional force flag to skip confirmation
	cmd.Flags().BoolVar(&flags.force, "force", false, "skip confirmation prompt")

	return cmd
}

// validateTeamRemoveFlags validates all flag values before making API calls.
// Returns an error if any flag value is invalid.
func validateTeamRemoveFlags(flags *teamRemoveFlags) error {
	// Validate that either member-id or email is provided (but not both)
	if flags.memberID == 0 && flags.email == "" {
		return fmt.Errorf("either --member-id or --email is required")
	}
	if flags.memberID != 0 && flags.email != "" {
		return fmt.Errorf("specify either --member-id or --email, not both")
	}

	return nil
}

// confirmRemoval prompts the user to confirm the member removal.
// Returns true if confirmed, false otherwise.
func confirmRemoval(email string) bool {
	fmt.Printf("Are you sure you want to remove %q from your organization? [y/N]: ", email)
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

// runTeamRemove executes the team remove command logic.
func runTeamRemove(ctx context.Context, flags *teamRemoveFlags) error {
	// Validate all flags before making any API calls
	if err := validateTeamRemoveFlags(flags); err != nil {
		return err
	}

	// Get authenticated API client (after validation passes)
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Create context with timeout
	reqCtx, cancel := context.WithTimeout(ctx, teamRemoveTimeout)
	defer cancel()

	// Resolve member ID (either directly or via email lookup)
	memberID, err := resolveMemberID(reqCtx, apiClient, flags.memberID, flags.email)
	if err != nil {
		return err
	}

	// Get the member's email for confirmation prompt if we only have ID
	memberEmail := flags.email
	if memberEmail == "" {
		// Look up the member to get their email for the confirmation message
		result, err := client.ListMembers(reqCtx, apiClient, &client.ListMembersOptions{
			Limit: 100,
		})
		if err != nil {
			return fmt.Errorf("failed to list team members: %w", err)
		}
		for _, m := range result.Members {
			if m.ID == memberID {
				memberEmail = m.Email
				break
			}
		}
		if memberEmail == "" {
			memberEmail = fmt.Sprintf("member ID %d", memberID)
		}
	}

	// Ask for confirmation unless --force or --no-input is specified
	if !flags.force && !GetNoInput() {
		if !confirmRemoval(memberEmail) {
			fmt.Println("Operation cancelled.")
			return nil
		}
	}

	// Call SDK to remove member
	if err := client.RemoveMember(reqCtx, apiClient, memberID); err != nil {
		return fmt.Errorf("failed to remove team member: %w", err)
	}

	// Print the success message
	return output.PrintMemberRemoved(memberID, memberEmail)
}
