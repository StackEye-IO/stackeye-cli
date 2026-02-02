// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
)

// teamListTimeout is the maximum time to wait for the API response.
const teamListTimeout = 30 * time.Second

// teamListFlags holds the flag values for the team list command.
type teamListFlags struct {
	page  int
	limit int
	role  string // Filter by role: owner, admin, member, viewer
}

// NewTeamListCmd creates and returns the team list subcommand.
func NewTeamListCmd() *cobra.Command {
	flags := &teamListFlags{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all team members",
		Long: `List all team members in your organization.

Displays member name, email, role, and when they joined the organization.
Results are paginated and can be output in various formats.

Roles:
  owner    Full control including billing and deletion
  admin    Manage team members and all resources
  member   Create and manage probes, alerts, channels
  viewer   Read-only access to all resources

Note: Role filtering is applied client-side. When combining --role with
pagination, filtered results may be fewer than --limit if the page contains
mixed roles. To see all members with a specific role, omit pagination flags.

Examples:
  # List all team members
  stackeye team list

  # Filter by role (only shows admins)
  stackeye team list --role admin

  # Filter by different roles (run separately)
  stackeye team list --role member
  stackeye team list --role viewer

  # Output as JSON for scripting
  stackeye team list -o json

  # Output as YAML
  stackeye team list -o yaml

  # Paginate through results
  stackeye team list --page 2 --limit 50`,
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTeamList(cmd.Context(), flags)
		},
	}

	// Define command-specific flags
	cmd.Flags().IntVar(&flags.page, "page", 1, "page number for pagination")
	cmd.Flags().IntVar(&flags.limit, "limit", 20, "results per page (max: 100)")
	cmd.Flags().StringVar(&flags.role, "role", "", "filter by role: owner, admin, member, viewer")

	return cmd
}

// validateTeamListFlags validates all flag values before making API calls.
// Returns an error if any flag value is invalid.
func validateTeamListFlags(flags *teamListFlags) error {
	if flags.limit < 1 || flags.limit > 100 {
		return fmt.Errorf("invalid limit %d: must be between 1 and 100", flags.limit)
	}

	if flags.page < 1 {
		return fmt.Errorf("invalid page %d: must be at least 1", flags.page)
	}

	// Validate role if provided (uses validTeamRoles from team_invite.go)
	if flags.role != "" {
		role := strings.ToLower(flags.role)
		roleValid := false
		for _, valid := range validTeamRoles {
			if valid == role {
				roleValid = true
				break
			}
		}
		if !roleValid {
			return fmt.Errorf("invalid role %q: must be one of owner, admin, member, viewer", flags.role)
		}
	}

	return nil
}

// runTeamList executes the team list command logic.
func runTeamList(ctx context.Context, flags *teamListFlags) error {
	// Validate all flags before making any API calls
	if err := validateTeamListFlags(flags); err != nil {
		return err
	}

	// Get authenticated API client (after validation passes)
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Build list options from validated flags
	// SDK uses offset-based pagination, convert page to offset
	offset := (flags.page - 1) * flags.limit
	opts := &client.ListMembersOptions{
		Limit:  flags.limit,
		Offset: offset,
	}

	// Call SDK to list team members with timeout
	reqCtx, cancel := context.WithTimeout(ctx, teamListTimeout)
	defer cancel()

	result, err := client.ListMembers(reqCtx, apiClient, opts)
	if err != nil {
		return fmt.Errorf("failed to list team members: %w", err)
	}

	// Apply client-side role filter if specified
	members := result.Members
	if flags.role != "" {
		members = filterMembersByRole(members, strings.ToLower(flags.role))
	}

	// Handle empty results
	if len(members) == 0 {
		if flags.role != "" {
			return output.PrintEmpty(fmt.Sprintf("No team members found with role %q", flags.role))
		}
		return output.PrintEmpty("No team members found. Invite members with 'stackeye team invite'")
	}

	// Print the team members using the configured output format
	return output.PrintTeamMembers(members)
}

// filterMembersByRole filters team members by role (case-insensitive).
func filterMembersByRole(members []client.TeamMember, role string) []client.TeamMember {
	filtered := make([]client.TeamMember, 0)
	for _, m := range members {
		if strings.EqualFold(string(m.Role), role) {
			filtered = append(filtered, m)
		}
	}
	return filtered
}
