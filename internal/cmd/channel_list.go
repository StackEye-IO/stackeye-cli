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

// channelListTimeout is the maximum time to wait for the API response.
const channelListTimeout = 30 * time.Second

// channelListFlags holds the flag values for the channel list command.
type channelListFlags struct {
	channelType string
	enabled     string
	page        int
	limit       int
}

// NewChannelListCmd creates and returns the channel list subcommand.
func NewChannelListCmd() *cobra.Command {
	flags := &channelListFlags{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all notification channels",
		Long: `List all notification channels in your organization.

Displays channel name, type, enabled status, and number of linked probes.
Results are paginated and can be filtered by type or enabled status.

Channel Types:
  email       Send notifications to email addresses
  slack       Post to Slack channels via incoming webhooks
  webhook     Send HTTP requests to custom endpoints
  pagerduty   Create incidents in PagerDuty
  discord     Post to Discord channels via webhooks
  teams       Post to Microsoft Teams channels
  sms         Send SMS text messages (requires SMS plan)

Examples:
  # List all channels
  stackeye channel list

  # List only Slack channels
  stackeye channel list --type slack

  # List disabled channels
  stackeye channel list --enabled=false

  # Output as JSON for scripting
  stackeye channel list -o json

  # Paginate through results
  stackeye channel list --page 2 --limit 50`,
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChannelList(cmd.Context(), flags)
		},
	}

	// Define command-specific flags
	cmd.Flags().StringVar(&flags.channelType, "type", "", "filter by channel type: email, slack, webhook, pagerduty, discord, teams, sms")
	cmd.Flags().StringVar(&flags.enabled, "enabled", "", "filter by enabled status: true, false")
	cmd.Flags().IntVar(&flags.page, "page", 1, "page number for pagination")
	cmd.Flags().IntVar(&flags.limit, "limit", 20, "results per page (max: 100)")

	return cmd
}

// validateChannelListFlags validates all flag values before making API calls.
// Returns an error if any flag value is invalid.
func validateChannelListFlags(flags *channelListFlags) error {
	if flags.limit < 1 || flags.limit > 100 {
		return fmt.Errorf("invalid limit %d: must be between 1 and 100", flags.limit)
	}

	if flags.page < 1 {
		return fmt.Errorf("invalid page %d: must be at least 1", flags.page)
	}

	if flags.channelType != "" {
		switch flags.channelType {
		case "email", "slack", "webhook", "pagerduty", "discord", "teams", "sms":
			// Valid types
		default:
			return fmt.Errorf("invalid channel type %q: must be email, slack, webhook, pagerduty, discord, teams, or sms", flags.channelType)
		}
	}

	if flags.enabled != "" {
		switch flags.enabled {
		case "true", "false":
			// Valid values
		default:
			return fmt.Errorf("invalid enabled value %q: must be true or false", flags.enabled)
		}
	}

	return nil
}

// parseChannelType converts a string channel type to the SDK ChannelType.
// Returns empty ChannelType if input is empty (no filter).
func parseChannelType(typeStr string) client.ChannelType {
	switch typeStr {
	case "email":
		return client.ChannelTypeEmail
	case "slack":
		return client.ChannelTypeSlack
	case "webhook":
		return client.ChannelTypeWebhook
	case "pagerduty":
		return client.ChannelTypePagerDuty
	case "discord":
		return client.ChannelTypeDiscord
	case "teams":
		return client.ChannelTypeTeams
	case "sms":
		return client.ChannelTypeSMS
	default:
		return ""
	}
}

// parseEnabledFilter converts a string enabled value to a *bool for filtering.
// Returns nil if input is empty (no filter).
func parseEnabledFilter(enabledStr string) *bool {
	switch enabledStr {
	case "true":
		enabled := true
		return &enabled
	case "false":
		enabled := false
		return &enabled
	default:
		return nil
	}
}

// runChannelList executes the channel list command logic.
func runChannelList(ctx context.Context, flags *channelListFlags) error {
	// Validate all flags before making any API calls
	if err := validateChannelListFlags(flags); err != nil {
		return err
	}

	// Parse validated flags
	channelType := parseChannelType(flags.channelType)
	enabledFilter := parseEnabledFilter(flags.enabled)

	// Get authenticated API client (after validation passes)
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Build list options from validated flags
	// SDK uses offset-based pagination, convert page to offset
	offset := (flags.page - 1) * flags.limit
	opts := &client.ListChannelsOptions{
		Limit:   flags.limit,
		Offset:  offset,
		Type:    channelType,
		Enabled: enabledFilter,
	}

	// Call SDK to list channels with timeout
	reqCtx, cancel := context.WithTimeout(ctx, channelListTimeout)
	defer cancel()

	result, err := client.ListChannels(reqCtx, apiClient, opts)
	if err != nil {
		return fmt.Errorf("failed to list channels: %w", err)
	}

	// Handle empty results
	if len(result.Channels) == 0 {
		return output.PrintEmpty("No channels found. Create one with 'stackeye channel create'")
	}

	// Print the channels using the configured output format
	return output.Print(result.Channels)
}
