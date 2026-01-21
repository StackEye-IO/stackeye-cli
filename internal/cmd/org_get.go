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

// orgGetTimeout is the maximum time to wait for the API response.
const orgGetTimeout = 30 * time.Second

// OrgDetail combines organization info with billing and usage details.
type OrgDetail struct {
	// Basic organization info
	ID        string `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Role      string `json:"role"`
	IsCurrent bool   `json:"is_current"`

	// Plan and billing info
	Plan          string  `json:"plan"`
	PlanStatus    string  `json:"plan_status"`
	TrialEndsAt   *string `json:"trial_ends_at,omitempty"`
	NextBillingAt *string `json:"next_billing_at,omitempty"`

	// Usage and limits
	MonitorCount    int `json:"monitor_count"`
	MonitorLimit    int `json:"monitor_limit"`
	TeamMemberCount int `json:"team_member_count"`
	TeamMemberLimit int `json:"team_member_limit"`
}

// NewOrgGetCmd creates and returns the org get subcommand.
func NewOrgGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [id|slug]",
		Short: "Get organization details",
		Long: `Get detailed information about an organization.

Shows organization settings, plan limits, current usage, and team member count.
If no ID or slug is provided, shows details for the current active organization.

The command accepts either the organization UUID or its URL-friendly slug.

Roles:
  owner    Full control including billing and deletion
  admin    Manage team members and all resources
  member   Create and manage probes, alerts, channels
  viewer   Read-only access to all resources

Plan Limits:
  - Monitor count: Number of probes/monitors you can create
  - Team members: Number of users that can be invited

Examples:
  # Get current organization details
  stackeye org get

  # Get organization by UUID
  stackeye org get 550e8400-e29b-41d4-a716-446655440000

  # Get organization by slug
  stackeye org get acme-corp

  # Output as JSON for scripting
  stackeye org get -o json

  # Output as YAML
  stackeye org get -o yaml`,
		Aliases: []string{"show", "info"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var identifier string
			if len(args) > 0 {
				identifier = args[0]
			}
			return runOrgGet(cmd.Context(), identifier)
		},
	}

	return cmd
}

// runOrgGet executes the org get command logic.
func runOrgGet(ctx context.Context, identifier string) error {
	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Create timeout context
	reqCtx, cancel := context.WithTimeout(ctx, orgGetTimeout)
	defer cancel()

	// Find the target organization
	org, err := findOrganization(reqCtx, apiClient, identifier)
	if err != nil {
		return err
	}

	// Build the detail response with additional info
	detail := OrgDetail{
		ID:        org.ID,
		Name:      org.Name,
		Slug:      org.Slug,
		Role:      org.Role,
		IsCurrent: org.IsCurrent,
	}

	// Fetch billing info (for plan and limits).
	// Errors are intentionally ignored - billing data is optional and may be
	// unavailable for users without billing access. We show what we can.
	billing, billingErr := client.GetBillingInfo(reqCtx, apiClient)
	if billingErr == nil {
		detail.Plan = billing.Plan
		detail.PlanStatus = billing.Status
		detail.MonitorCount = billing.MonitorCount
		detail.MonitorLimit = billing.MonitorLimit
		detail.TrialEndsAt = billing.TrialEndsAt
		detail.NextBillingAt = billing.NextBillingAt
	}

	// Fetch usage info (for team member count and limits).
	// Errors are intentionally ignored - usage data is optional and may be
	// unavailable for users without billing access. We show what we can.
	usage, usageErr := client.GetUsage(reqCtx, apiClient)
	if usageErr == nil {
		detail.TeamMemberCount = usage.TeamMembersCount
		detail.TeamMemberLimit = usage.TeamMembersLimit
		// Usage may have more accurate monitor counts
		if usage.MonitorsCount > 0 || detail.MonitorCount == 0 {
			detail.MonitorCount = usage.MonitorsCount
		}
		if usage.MonitorsLimit > 0 || detail.MonitorLimit == 0 {
			detail.MonitorLimit = usage.MonitorsLimit
		}
	}

	// Print the organization details
	return output.Print(detail)
}

// findOrganization locates an organization by ID, slug, or returns the current org.
func findOrganization(ctx context.Context, apiClient *client.Client, identifier string) (*client.Organization, error) {
	// List all organizations the user belongs to
	result, err := client.ListOrganizations(ctx, apiClient)
	if err != nil {
		return nil, fmt.Errorf("failed to list organizations: %w", err)
	}

	if len(result.Organizations) == 0 {
		return nil, fmt.Errorf("no organizations found")
	}

	// If no identifier provided, return the current organization
	if identifier == "" {
		for i := range result.Organizations {
			if result.Organizations[i].IsCurrent {
				return &result.Organizations[i], nil
			}
		}
		// If no current org marked, return the first one
		return &result.Organizations[0], nil
	}

	// Search by ID or slug
	for i := range result.Organizations {
		org := &result.Organizations[i]
		if org.ID == identifier || org.Slug == identifier {
			return org, nil
		}
	}

	return nil, fmt.Errorf("organization not found: %s", identifier)
}
