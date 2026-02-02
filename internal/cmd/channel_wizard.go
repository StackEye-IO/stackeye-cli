// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/mail"
	"regexp"
	"strings"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/StackEye-IO/stackeye-go-sdk/interactive"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// wizardChannelType represents a channel type option in the wizard.
type wizardChannelType struct {
	value       string
	label       string
	description string
}

// supportedChannelTypes defines the channel types available in the wizard.
var supportedChannelTypes = []wizardChannelType{
	{value: "email", label: "Email", description: "Send notifications to email addresses"},
	{value: "slack", label: "Slack", description: "Post to Slack channels via webhooks"},
	{value: "discord", label: "Discord", description: "Post to Discord channels via webhooks"},
	{value: "teams", label: "Microsoft Teams", description: "Post to Teams channels via webhooks"},
	{value: "webhook", label: "Webhook", description: "Send HTTP requests to custom endpoints"},
	{value: "pagerduty", label: "PagerDuty", description: "Create incidents in PagerDuty"},
	{value: "sms", label: "SMS", description: "Send SMS text messages (requires SMS plan)"},
}

// pagerDutySeverities defines valid PagerDuty severity levels.
var pagerDutySeverities = []string{"critical", "error", "warning", "info"}

// webhookMethods defines valid HTTP methods for webhook channels.
var webhookMethods = []string{"POST", "GET", "PUT", "PATCH", "DELETE"}

// NewChannelWizardCmd creates and returns the channel wizard subcommand.
func NewChannelWizardCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wizard",
		Short: "Interactive wizard for creating notification channels",
		Long: `Create a new notification channel using an interactive wizard.

The wizard guides you through the channel creation process step by step,
prompting for all required information based on the channel type you select.

Supported Channel Types:
  email       Send notifications to email addresses
              Prompts for: email address
  slack       Post to Slack channels via incoming webhooks
              Prompts for: webhook URL
  discord     Post to Discord channels via webhooks
              Prompts for: webhook URL
  teams       Post to Microsoft Teams channels
              Prompts for: webhook URL
  webhook     Send HTTP requests to custom endpoints
              Prompts for: URL, HTTP method, custom headers
  pagerduty   Create incidents in PagerDuty
              Prompts for: routing key, severity level
  sms         Send SMS text messages (requires SMS plan)
              Prompts for: phone number in E.164 format

After creating the channel, the wizard offers to send a test notification
to verify the configuration works correctly.

Examples:
  # Start the channel creation wizard
  stackeye channel wizard

  # The wizard will guide you through:
  # 1. Selecting a channel type
  # 2. Entering type-specific configuration
  # 3. Naming your channel
  # 4. Optionally sending a test notification`,
		Aliases: []string{"wiz", "w"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChannelWizard(cmd.Context())
		},
	}

	return cmd
}

// runChannelWizard executes the interactive channel creation wizard.
func runChannelWizard(ctx context.Context) error {
	// Check for non-interactive mode
	if GetNoInput() {
		return runChannelWizardNonInteractive()
	}

	wiz := interactive.NewWizard(&interactive.WizardOptions{
		Title:         "Channel Creation Wizard",
		Description:   "Let's set up a new notification channel",
		ShowProgress:  true,
		ConfirmCancel: true,
	})

	wiz.AddSteps(
		&interactive.Step{
			Name:        "channel-type",
			Title:       "Select Channel Type",
			Description: "Choose how you want to receive notifications",
			Run:         stepSelectChannelType,
		},
		&interactive.Step{
			Name:        "email-config",
			Title:       "Email Configuration",
			Description: "Enter the email address for notifications",
			Run:         stepEmailConfig,
			Skip:        func(w *interactive.Wizard) bool { return w.GetDataString("channel_type") != "email" },
		},
		&interactive.Step{
			Name:        "slack-config",
			Title:       "Slack Configuration",
			Description: "Enter your Slack webhook URL",
			Run:         stepSlackConfig,
			Skip:        func(w *interactive.Wizard) bool { return w.GetDataString("channel_type") != "slack" },
		},
		&interactive.Step{
			Name:        "discord-config",
			Title:       "Discord Configuration",
			Description: "Enter your Discord webhook URL",
			Run:         stepDiscordConfig,
			Skip:        func(w *interactive.Wizard) bool { return w.GetDataString("channel_type") != "discord" },
		},
		&interactive.Step{
			Name:        "teams-config",
			Title:       "Teams Configuration",
			Description: "Enter your Microsoft Teams webhook URL",
			Run:         stepTeamsConfig,
			Skip:        func(w *interactive.Wizard) bool { return w.GetDataString("channel_type") != "teams" },
		},
		&interactive.Step{
			Name:        "webhook-config",
			Title:       "Webhook Configuration",
			Description: "Configure your custom webhook endpoint",
			Run:         stepWebhookConfig,
			Skip:        func(w *interactive.Wizard) bool { return w.GetDataString("channel_type") != "webhook" },
		},
		&interactive.Step{
			Name:        "pagerduty-config",
			Title:       "PagerDuty Configuration",
			Description: "Enter your PagerDuty integration details",
			Run:         stepPagerDutyConfig,
			Skip:        func(w *interactive.Wizard) bool { return w.GetDataString("channel_type") != "pagerduty" },
		},
		&interactive.Step{
			Name:        "sms-config",
			Title:       "SMS Configuration",
			Description: "Enter the phone number for SMS notifications",
			Run:         stepSMSConfig,
			Skip:        func(w *interactive.Wizard) bool { return w.GetDataString("channel_type") != "sms" },
		},
		&interactive.Step{
			Name:        "channel-name",
			Title:       "Channel Name",
			Description: "Give your channel a memorable name",
			Run:         stepChannelName,
		},
		&interactive.Step{
			Name:        "create-channel",
			Title:       "Create Channel",
			Description: "Creating your notification channel",
			Run:         stepCreateChannel,
		},
		&interactive.Step{
			Name:        "test-channel",
			Title:       "Test Channel",
			Description: "Send a test notification to verify configuration",
			Run:         stepTestChannel,
		},
	)

	return wiz.Run(ctx)
}

// stepSelectChannelType prompts the user to select a channel type.
func stepSelectChannelType(ctx context.Context, w *interactive.Wizard) error {
	options := make([]string, len(supportedChannelTypes))
	for i, ct := range supportedChannelTypes {
		options[i] = fmt.Sprintf("%s - %s", ct.label, ct.description)
	}

	selected, err := interactive.AskSelect(&interactive.SelectPromptOptions{
		Message:  "What type of channel do you want to create?",
		Options:  options,
		PageSize: 7,
	})
	if err != nil {
		return err
	}

	// Extract the channel type from the selection
	for _, ct := range supportedChannelTypes {
		if strings.HasPrefix(selected, ct.label+" - ") {
			w.SetData("channel_type", ct.value)
			w.SetData("channel_type_label", ct.label)
			return nil
		}
	}

	return fmt.Errorf("unexpected selection: %s", selected)
}

// stepEmailConfig prompts for email channel configuration.
func stepEmailConfig(_ context.Context, w *interactive.Wizard) error {
	email, err := interactive.AskString(&interactive.StringPromptOptions{
		Message: "Email address:",
		Help:    "The email address where notifications will be sent",
		Validate: func(s string) error {
			if s == "" {
				return fmt.Errorf("email address is required")
			}
			if _, err := mail.ParseAddress(s); err != nil {
				return fmt.Errorf("invalid email address: %w", err)
			}
			return nil
		},
	})
	if err != nil {
		return err
	}

	w.SetData("email", email)
	return nil
}

// stepSlackConfig prompts for Slack channel configuration.
func stepSlackConfig(_ context.Context, w *interactive.Wizard) error {
	webhookURL, err := interactive.AskString(&interactive.StringPromptOptions{
		Message: "Slack webhook URL:",
		Help:    "The incoming webhook URL from your Slack workspace",
		Validate: func(s string) error {
			if s == "" {
				return fmt.Errorf("webhook URL is required")
			}
			if err := validateHTTPURL(s); err != nil {
				return fmt.Errorf("invalid URL: %w", err)
			}
			return nil
		},
	})
	if err != nil {
		return err
	}

	w.SetData("webhook_url", webhookURL)
	return nil
}

// stepDiscordConfig prompts for Discord channel configuration.
func stepDiscordConfig(_ context.Context, w *interactive.Wizard) error {
	webhookURL, err := interactive.AskString(&interactive.StringPromptOptions{
		Message: "Discord webhook URL:",
		Help:    "The webhook URL from your Discord server settings",
		Validate: func(s string) error {
			if s == "" {
				return fmt.Errorf("webhook URL is required")
			}
			if err := validateHTTPURL(s); err != nil {
				return fmt.Errorf("invalid URL: %w", err)
			}
			return nil
		},
	})
	if err != nil {
		return err
	}

	w.SetData("webhook_url", webhookURL)
	return nil
}

// stepTeamsConfig prompts for Microsoft Teams channel configuration.
func stepTeamsConfig(_ context.Context, w *interactive.Wizard) error {
	webhookURL, err := interactive.AskString(&interactive.StringPromptOptions{
		Message: "Microsoft Teams webhook URL:",
		Help:    "The incoming webhook URL from your Teams channel connector",
		Validate: func(s string) error {
			if s == "" {
				return fmt.Errorf("webhook URL is required")
			}
			if err := validateHTTPURL(s); err != nil {
				return fmt.Errorf("invalid URL: %w", err)
			}
			return nil
		},
	})
	if err != nil {
		return err
	}

	w.SetData("webhook_url", webhookURL)
	return nil
}

// stepWebhookConfig prompts for custom webhook channel configuration.
func stepWebhookConfig(_ context.Context, w *interactive.Wizard) error {
	// Get the endpoint URL
	endpointURL, err := interactive.AskString(&interactive.StringPromptOptions{
		Message: "Webhook endpoint URL:",
		Help:    "The HTTP/HTTPS URL where notifications will be sent",
		Validate: func(s string) error {
			if s == "" {
				return fmt.Errorf("URL is required")
			}
			if err := validateHTTPURL(s); err != nil {
				return fmt.Errorf("invalid URL: %w", err)
			}
			return nil
		},
	})
	if err != nil {
		return err
	}
	w.SetData("webhook_endpoint_url", endpointURL)

	// Get the HTTP method
	method, err := interactive.AskSelect(&interactive.SelectPromptOptions{
		Message:  "HTTP method:",
		Options:  webhookMethods,
		Default:  "POST",
		Help:     "The HTTP method to use when sending notifications",
		PageSize: 5,
	})
	if err != nil {
		return err
	}
	w.SetData("webhook_method", method)

	// Ask if custom headers are needed
	addHeaders, err := interactive.AskConfirm(&interactive.ConfirmPromptOptions{
		Message: "Do you want to add custom HTTP headers?",
		Default: false,
		Help:    "Add custom headers like Authorization tokens",
	})
	if err != nil {
		return err
	}

	if addHeaders {
		headersJSON, err := interactive.AskString(&interactive.StringPromptOptions{
			Message: "Headers (JSON object):",
			Default: `{"Content-Type": "application/json"}`,
			Help:    `Enter headers as a JSON object, e.g., {"Authorization": "Bearer token"}`,
			Validate: func(s string) error {
				if s == "" {
					return nil // Headers are optional
				}
				var h map[string]string
				if err := json.Unmarshal([]byte(s), &h); err != nil {
					return fmt.Errorf("invalid JSON: %w", err)
				}
				return nil
			},
		})
		if err != nil {
			return err
		}
		w.SetData("webhook_headers", headersJSON)
	}

	return nil
}

// stepPagerDutyConfig prompts for PagerDuty channel configuration.
func stepPagerDutyConfig(_ context.Context, w *interactive.Wizard) error {
	// Get the routing key
	routingKey, err := interactive.AskString(&interactive.StringPromptOptions{
		Message: "PagerDuty routing key:",
		Help:    "The integration/routing key from your PagerDuty service",
		Validate: func(s string) error {
			if s == "" {
				return fmt.Errorf("routing key is required")
			}
			return nil
		},
	})
	if err != nil {
		return err
	}
	w.SetData("pagerduty_routing_key", routingKey)

	// Get the severity level
	severity, err := interactive.AskSelect(&interactive.SelectPromptOptions{
		Message:  "Default severity level:",
		Options:  pagerDutySeverities,
		Default:  "critical",
		Help:     "The severity level for incidents created in PagerDuty",
		PageSize: 4,
	})
	if err != nil {
		return err
	}
	w.SetData("pagerduty_severity", severity)

	return nil
}

// stepSMSConfig prompts for SMS channel configuration.
func stepSMSConfig(_ context.Context, w *interactive.Wizard) error {
	phoneNumber, err := interactive.AskString(&interactive.StringPromptOptions{
		Message: "Phone number (E.164 format):",
		Help:    "Phone number with country code, e.g., +15551234567",
		Validate: func(s string) error {
			if s == "" {
				return fmt.Errorf("phone number is required")
			}
			// E.164 format: + followed by country code and number, 8-15 digits total
			e164Pattern := regexp.MustCompile(`^\+[1-9]\d{7,14}$`)
			if !e164Pattern.MatchString(s) {
				return fmt.Errorf("must be in E.164 format (e.g., +15551234567)")
			}
			return nil
		},
	})
	if err != nil {
		return err
	}

	w.SetData("phone_number", phoneNumber)
	return nil
}

// stepChannelName prompts for the channel name.
func stepChannelName(_ context.Context, w *interactive.Wizard) error {
	channelType := w.GetDataString("channel_type_label")
	defaultName := fmt.Sprintf("My %s Channel", channelType)

	name, err := interactive.AskString(&interactive.StringPromptOptions{
		Message: "Channel name:",
		Default: defaultName,
		Help:    "A memorable name to identify this channel",
		Validate: func(s string) error {
			if s == "" {
				return fmt.Errorf("channel name is required")
			}
			if len(s) > 100 {
				return fmt.Errorf("name must be 100 characters or less")
			}
			return nil
		},
	})
	if err != nil {
		return err
	}

	w.SetData("channel_name", name)
	return nil
}

// stepCreateChannel creates the channel via the API.
func stepCreateChannel(ctx context.Context, w *interactive.Wizard) error {
	// Get API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Build the channel config based on type
	channelType := w.GetDataString("channel_type")
	config, err := buildWizardChannelConfig(w, channelType)
	if err != nil {
		return fmt.Errorf("failed to build channel config: %w", err)
	}

	// Create the request
	enabled := true
	req := &client.CreateChannelRequest{
		Name:    w.GetDataString("channel_name"),
		Type:    client.ChannelType(channelType),
		Config:  config,
		Enabled: &enabled,
	}

	// Create the channel
	channel, err := client.CreateChannel(ctx, apiClient, req)
	if err != nil {
		return fmt.Errorf("failed to create channel: %w", err)
	}

	// Store the channel ID for the test step
	w.SetData("channel_id", channel.ID.String())

	// Print success message
	fmt.Fprintf(w.Writer(), "\nChannel created successfully!\n\n")
	if err := output.Print(channel); err != nil {
		return fmt.Errorf("failed to print channel: %w", err)
	}
	fmt.Fprintln(w.Writer())

	return nil
}

// stepTestChannel offers to send a test notification.
func stepTestChannel(ctx context.Context, w *interactive.Wizard) error {
	sendTest, err := interactive.AskConfirm(&interactive.ConfirmPromptOptions{
		Message: "Would you like to send a test notification?",
		Default: true,
		Help:    "This will send a test message to verify your channel is configured correctly",
	})
	if err != nil {
		return err
	}

	channelIDStr := w.GetDataString("channel_id")
	if !sendTest {
		fmt.Fprintf(w.Writer(), "\nSkipping test notification. You can test later with:\n")
		fmt.Fprintf(w.Writer(), "  stackeye channel test %s\n", channelIDStr)
		return nil
	}

	// Get API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Send test notification
	channelID, err := uuid.Parse(channelIDStr)
	if err != nil {
		return fmt.Errorf("invalid channel ID: %w", err)
	}
	fmt.Fprintf(w.Writer(), "\nSending test notification...\n")

	_, err = client.TestChannel(ctx, apiClient, channelID)
	if err != nil {
		fmt.Fprintf(w.Writer(), "Test notification failed: %v\n", err)
		fmt.Fprintf(w.Writer(), "Please verify your channel configuration and try again with:\n")
		fmt.Fprintf(w.Writer(), "  stackeye channel test %s\n", channelIDStr)
		return nil // Don't fail the wizard, the channel was still created
	}

	fmt.Fprintf(w.Writer(), "Test notification sent successfully!\n")
	return nil
}

// runChannelWizardNonInteractive provides guidance when wizard is run in non-interactive mode.
func runChannelWizardNonInteractive() error {
	fmt.Println("Channel Creation Wizard (Non-Interactive Mode)")
	fmt.Println()
	fmt.Println("The channel wizard requires interactive input.")
	fmt.Println()
	fmt.Println("To create a channel non-interactively, use:")
	fmt.Println()
	fmt.Println("  stackeye channel create --name <name> --type <type> [options]")
	fmt.Println()
	fmt.Println("Required flags:")
	fmt.Println("  --name       Channel name")
	fmt.Println("  --type       Channel type: email, slack, webhook, pagerduty, discord, teams, sms")
	fmt.Println()
	fmt.Println("Type-specific options:")
	fmt.Println("  email:       --email <address>")
	fmt.Println("  slack:       --webhook-url <url>")
	fmt.Println("  discord:     --webhook-url <url>")
	fmt.Println("  teams:       --webhook-url <url>")
	fmt.Println("  webhook:     --url <url> [--method <method>] [--headers <json>]")
	fmt.Println("  pagerduty:   --routing-key <key> [--severity <level>]")
	fmt.Println("  sms:         --phone-number <e164>")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  stackeye channel create --name \"Ops Team\" --type email --email ops@company.com")
	fmt.Println("  stackeye channel create --name \"Alerts\" --type slack --webhook-url <webhook-url>")
	fmt.Println()
	return nil
}

// buildWizardChannelConfig builds the JSON config for the channel type from wizard data.
func buildWizardChannelConfig(w *interactive.Wizard, channelType string) (json.RawMessage, error) {
	switch channelType {
	case "email":
		return client.NewEmailConfig(w.GetDataString("email"))

	case "slack":
		return client.NewSlackConfig(w.GetDataString("webhook_url"))

	case "discord":
		return client.NewDiscordConfig(w.GetDataString("webhook_url"))

	case "teams":
		return client.NewTeamsConfig(w.GetDataString("webhook_url"))

	case "webhook":
		var headers map[string]string
		headersJSON := w.GetDataString("webhook_headers")
		if headersJSON != "" {
			if err := json.Unmarshal([]byte(headersJSON), &headers); err != nil {
				return nil, fmt.Errorf("invalid headers JSON: %w", err)
			}
		}
		return client.NewWebhookConfig(
			w.GetDataString("webhook_endpoint_url"),
			w.GetDataString("webhook_method"),
			headers,
		)

	case "pagerduty":
		return client.NewPagerDutyConfig(
			w.GetDataString("pagerduty_routing_key"),
			w.GetDataString("pagerduty_severity"),
		)

	case "sms":
		return client.NewSMSConfig(w.GetDataString("phone_number"))

	default:
		return nil, fmt.Errorf("unsupported channel type: %s", channelType)
	}
}
