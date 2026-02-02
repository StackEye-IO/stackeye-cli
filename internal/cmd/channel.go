// Package cmd implements the CLI commands for StackEye.
package cmd

import "github.com/spf13/cobra"

// NewChannelCmd creates and returns the channel parent command.
// This command provides management of notification channels.
func NewChannelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "channel",
		Short: "Manage notification channels",
		Long: `Manage notification channels for your organization.

Channels are the destinations where StackEye sends alert notifications when
a probe detects an issue. Configure multiple channels to ensure your team
gets notified through their preferred communication methods.

Supported Channel Types:
  email       Send notifications to email addresses
  slack       Post to Slack channels via incoming webhooks
  webhook     Send HTTP requests to custom endpoints
  pagerduty   Create incidents in PagerDuty
  discord     Post to Discord channels via webhooks
  teams       Post to Microsoft Teams channels
  sms         Send SMS text messages (requires SMS plan)

Available Commands:
  list        List all notification channels
  get         Get details of a specific channel
  create      Create a new notification channel
  update      Update an existing channel
  delete      Delete a notification channel
  test        Send a test notification through a channel
  wizard      Interactive wizard for creating channels

Examples:
  # List all channels
  stackeye channel list

  # List only Slack channels
  stackeye channel list --type slack

  # Get details for a specific channel
  stackeye channel get <channel-id>

  # Create a new email channel
  stackeye channel create --name "Ops Team" --type email --email {your-email}

  # Create a Slack channel
  stackeye channel create --name "Alerts" --type slack --webhook-url {slack-webhook-url}

  # Create a webhook channel
  stackeye channel create --name "Custom" --type webhook --url {webhook-url}

  # Update a channel name
  stackeye channel update <channel-id> --name "New Name"

  # Disable a channel
  stackeye channel update <channel-id> --enabled=false

  # Delete a channel
  stackeye channel delete <channel-id>

  # Delete without confirmation
  stackeye channel delete <channel-id> --yes

For more information about a specific command:
  stackeye channel [command] --help`,
		Aliases: []string{"channels", "ch"},
	}

	// Register subcommands
	cmd.AddCommand(NewChannelListCmd())
	cmd.AddCommand(NewChannelGetCmd())
	cmd.AddCommand(NewChannelCreateCmd())
	cmd.AddCommand(NewChannelUpdateCmd())
	cmd.AddCommand(NewChannelDeleteCmd())
	cmd.AddCommand(NewChannelTestCmd())
	cmd.AddCommand(NewChannelWizardCmd())

	return cmd
}
