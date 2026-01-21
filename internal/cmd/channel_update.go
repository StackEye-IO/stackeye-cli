// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// channelUpdateTimeout is the maximum time to wait for the API response.
const channelUpdateTimeout = 30 * time.Second

// channelUpdateFlags holds the flag values for the channel update command.
// All fields are pointers to support partial updates (nil = not specified).
type channelUpdateFlags struct {
	// Basic fields
	name    *string
	enabled *bool

	// Type-specific: email
	email *string

	// Type-specific: slack, discord, teams
	webhookURL *string

	// Type-specific: webhook
	url     *string
	method  *string
	headers *string

	// Type-specific: pagerduty
	routingKey *string
	severity   *string

	// Type-specific: sms
	phoneNumber *string

	// YAML file input
	fromFile string
}

// channelUpdateYAMLConfig represents the YAML structure for --from-file input on update.
type channelUpdateYAMLConfig struct {
	Name    *string `yaml:"name,omitempty"`
	Enabled *bool   `yaml:"enabled,omitempty"`
	Config  *struct {
		// Email
		Address string `yaml:"address,omitempty"`
		// Slack, Discord, Teams
		WebhookURL string `yaml:"webhook_url,omitempty"`
		// Webhook
		URL     string            `yaml:"url,omitempty"`
		Method  string            `yaml:"method,omitempty"`
		Headers map[string]string `yaml:"headers,omitempty"`
		// PagerDuty
		RoutingKey string `yaml:"routing_key,omitempty"`
		Severity   string `yaml:"severity,omitempty"`
		// SMS
		PhoneNumber string `yaml:"phone_number,omitempty"`
	} `yaml:"config,omitempty"`
}

// NewChannelUpdateCmd creates and returns the channel update subcommand.
func NewChannelUpdateCmd() *cobra.Command {
	flags := &channelUpdateFlags{}

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing notification channel",
		Long: `Update an existing notification channel configuration.

Only the specified flags will be updated; all other fields remain unchanged.
This allows for partial updates without needing to specify the entire configuration.

Note: Channel type cannot be changed after creation. To change the type,
delete the channel and create a new one.

Examples:
  # Update channel name
  stackeye channel update 550e8400-e29b-41d4-a716-446655440000 --name "New Name"

  # Disable a channel
  stackeye channel update 550e8400-e29b-41d4-a716-446655440000 --enabled=false

  # Re-enable a channel
  stackeye channel update 550e8400-e29b-41d4-a716-446655440000 --enabled=true

  # Update email address (for email channel)
  stackeye channel update 550e8400-e29b-41d4-a716-446655440000 --email {new-email}

  # Update webhook URL (for slack, discord, teams channels)
  stackeye channel update 550e8400-e29b-41d4-a716-446655440000 --webhook-url {new-webhook-url}

  # Update webhook channel config
  stackeye channel update 550e8400-e29b-41d4-a716-446655440000 \
    --url {new-endpoint-url} \
    --method PUT \
    --headers '{"Authorization":"Bearer {token}"}'

  # Update PagerDuty severity
  stackeye channel update 550e8400-e29b-41d4-a716-446655440000 --severity warning

  # Update from YAML file
  stackeye channel update 550e8400-e29b-41d4-a716-446655440000 --from-file channel-updates.yaml`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChannelUpdate(cmd, args[0], flags)
		},
	}

	// Basic flags (all optional for partial updates)
	cmd.Flags().StringVar(stringPtrVar(&flags.name), "name", "", "channel name")
	cmd.Flags().BoolVar(boolPtrVar(&flags.enabled), "enabled", false, "whether the channel is enabled")

	// Type-specific flags (used to update config)
	cmd.Flags().StringVar(stringPtrVar(&flags.email), "email", "", "email address (for email type)")
	cmd.Flags().StringVar(stringPtrVar(&flags.webhookURL), "webhook-url", "", "webhook URL (for slack, discord, teams types)")
	cmd.Flags().StringVar(stringPtrVar(&flags.url), "url", "", "endpoint URL (for webhook type)")
	cmd.Flags().StringVar(stringPtrVar(&flags.method), "method", "", "HTTP method (for webhook type)")
	cmd.Flags().StringVar(stringPtrVar(&flags.headers), "headers", "", "HTTP headers as JSON object (for webhook type)")
	cmd.Flags().StringVar(stringPtrVar(&flags.routingKey), "routing-key", "", "routing key (for pagerduty type)")
	cmd.Flags().StringVar(stringPtrVar(&flags.severity), "severity", "", "severity level (for pagerduty type)")
	cmd.Flags().StringVar(stringPtrVar(&flags.phoneNumber), "phone-number", "", "phone number in E.164 format (for sms type)")

	// File input
	cmd.Flags().StringVar(&flags.fromFile, "from-file", "", "update channel from YAML file")

	return cmd
}

// runChannelUpdate executes the channel update command logic.
func runChannelUpdate(cmd *cobra.Command, idArg string, flags *channelUpdateFlags) error {
	// Parse and validate UUID
	channelID, err := uuid.Parse(idArg)
	if err != nil {
		return fmt.Errorf("invalid channel ID %q: must be a valid UUID", idArg)
	}

	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Handle --from-file if provided
	if flags.fromFile != "" {
		return runChannelUpdateFromFile(cmd.Context(), apiClient, channelID, flags.fromFile)
	}

	// Build request from flags
	req, err := buildChannelUpdateRequest(cmd, flags, apiClient, channelID)
	if err != nil {
		return err
	}

	// Call SDK to update channel with timeout
	ctx, cancel := context.WithTimeout(cmd.Context(), channelUpdateTimeout)
	defer cancel()

	channel, err := client.UpdateChannel(ctx, apiClient, channelID, req)
	if err != nil {
		return fmt.Errorf("failed to update channel: %w", err)
	}

	// Print the updated channel using the configured output format
	return output.Print(channel)
}

// buildChannelUpdateRequest constructs the API request from command flags.
func buildChannelUpdateRequest(cmd *cobra.Command, flags *channelUpdateFlags, apiClient *client.Client, channelID uuid.UUID) (*client.UpdateChannelRequest, error) {
	req := &client.UpdateChannelRequest{}
	hasUpdates := false

	// Check name flag
	if cmd.Flags().Changed("name") {
		if *flags.name == "" {
			return nil, fmt.Errorf("--name cannot be empty")
		}
		req.Name = flags.name
		hasUpdates = true
	}

	// Check enabled flag
	if cmd.Flags().Changed("enabled") {
		req.Enabled = flags.enabled
		hasUpdates = true
	}

	// Check if any config-related flags were provided
	configChanged := hasConfigFlags(cmd)
	if configChanged {
		// We need to fetch the current channel to know its type
		ctx, cancel := context.WithTimeout(cmd.Context(), channelUpdateTimeout)
		defer cancel()

		currentChannel, err := client.GetChannel(ctx, apiClient, channelID)
		if err != nil {
			return nil, fmt.Errorf("failed to get current channel: %w", err)
		}

		// Build new config based on channel type
		config, err := buildUpdatedConfig(cmd, flags, currentChannel)
		if err != nil {
			return nil, err
		}
		req.Config = config
		hasUpdates = true
	}

	// Require at least one update flag
	if !hasUpdates {
		return nil, fmt.Errorf("no update flags specified; use --help to see available options")
	}

	return req, nil
}

// hasConfigFlags checks if any config-related flags were changed.
func hasConfigFlags(cmd *cobra.Command) bool {
	configFlags := []string{
		"email", "webhook-url", "url", "method", "headers",
		"routing-key", "severity", "phone-number",
	}
	return slices.ContainsFunc(configFlags, func(flag string) bool {
		return cmd.Flags().Changed(flag)
	})
}

// buildUpdatedConfig builds an updated config for the channel based on its type.
func buildUpdatedConfig(cmd *cobra.Command, flags *channelUpdateFlags, channel *client.Channel) (json.RawMessage, error) {
	// Validate that provided config flags match the channel type
	if err := validateConfigFlagsForType(cmd, channel.Type); err != nil {
		return nil, err
	}

	switch channel.Type {
	case client.ChannelTypeEmail:
		return buildUpdatedEmailConfig(cmd, flags, channel)
	case client.ChannelTypeSlack:
		return buildUpdatedSlackConfig(cmd, flags, channel)
	case client.ChannelTypeWebhook:
		return buildUpdatedWebhookConfig(cmd, flags, channel)
	case client.ChannelTypePagerDuty:
		return buildUpdatedPagerDutyConfig(cmd, flags, channel)
	case client.ChannelTypeDiscord:
		return buildUpdatedDiscordConfig(cmd, flags, channel)
	case client.ChannelTypeTeams:
		return buildUpdatedTeamsConfig(cmd, flags, channel)
	case client.ChannelTypeSMS:
		return buildUpdatedSMSConfig(cmd, flags, channel)
	default:
		return nil, fmt.Errorf("unsupported channel type: %s", channel.Type)
	}
}

// validateConfigFlagsForType ensures provided config flags are valid for the channel type.
// This prevents silent failures where mismatched flags (e.g., --email on a Slack channel) are ignored.
func validateConfigFlagsForType(cmd *cobra.Command, channelType client.ChannelType) error {
	// Define which flags are valid for each channel type
	validFlags := map[client.ChannelType][]string{
		client.ChannelTypeEmail:     {"email"},
		client.ChannelTypeSlack:     {"webhook-url"},
		client.ChannelTypeWebhook:   {"url", "method", "headers"},
		client.ChannelTypePagerDuty: {"routing-key", "severity"},
		client.ChannelTypeDiscord:   {"webhook-url"},
		client.ChannelTypeTeams:     {"webhook-url"},
		client.ChannelTypeSMS:       {"phone-number"},
	}

	// All config-related flags
	allConfigFlags := []string{
		"email", "webhook-url", "url", "method", "headers",
		"routing-key", "severity", "phone-number",
	}

	valid, ok := validFlags[channelType]
	if !ok {
		return fmt.Errorf("unsupported channel type: %s", channelType)
	}

	// Check each changed flag to ensure it's valid for this type
	var invalidFlags []string
	for _, flag := range allConfigFlags {
		if cmd.Flags().Changed(flag) {
			if !slices.Contains(valid, flag) {
				invalidFlags = append(invalidFlags, "--"+flag)
			}
		}
	}

	if len(invalidFlags) > 0 {
		validFlagsList := make([]string, len(valid))
		for i, f := range valid {
			validFlagsList[i] = "--" + f
		}
		return fmt.Errorf("flag(s) %s not valid for %s channel; valid config flags: %s",
			strings.Join(invalidFlags, ", "),
			channelType,
			strings.Join(validFlagsList, ", "))
	}

	return nil
}

// buildUpdatedEmailConfig builds an updated email config.
func buildUpdatedEmailConfig(cmd *cobra.Command, flags *channelUpdateFlags, channel *client.Channel) (json.RawMessage, error) {
	// Parse current config
	var current client.EmailChannelConfig
	if err := json.Unmarshal(channel.Config, &current); err != nil {
		return nil, fmt.Errorf("failed to parse current email config: %w", err)
	}

	// Apply updates
	if cmd.Flags().Changed("email") {
		if *flags.email == "" {
			return nil, fmt.Errorf("--email cannot be empty")
		}
		current.Address = *flags.email
	}

	// Validate and build new config
	return buildEmailConfig(current.Address)
}

// buildUpdatedSlackConfig builds an updated Slack config.
func buildUpdatedSlackConfig(cmd *cobra.Command, flags *channelUpdateFlags, channel *client.Channel) (json.RawMessage, error) {
	// Parse current config
	var current client.SlackChannelConfig
	if err := json.Unmarshal(channel.Config, &current); err != nil {
		return nil, fmt.Errorf("failed to parse current slack config: %w", err)
	}

	// Apply updates
	if cmd.Flags().Changed("webhook-url") {
		if *flags.webhookURL == "" {
			return nil, fmt.Errorf("--webhook-url cannot be empty")
		}
		current.WebhookURL = *flags.webhookURL
	}

	// Validate and build new config
	return buildSlackConfig(current.WebhookURL)
}

// buildUpdatedWebhookConfig builds an updated webhook config.
func buildUpdatedWebhookConfig(cmd *cobra.Command, flags *channelUpdateFlags, channel *client.Channel) (json.RawMessage, error) {
	// Parse current config
	var current client.WebhookChannelConfig
	if err := json.Unmarshal(channel.Config, &current); err != nil {
		return nil, fmt.Errorf("failed to parse current webhook config: %w", err)
	}

	// Apply updates
	if cmd.Flags().Changed("url") {
		if *flags.url == "" {
			return nil, fmt.Errorf("--url cannot be empty")
		}
		current.URL = *flags.url
	}
	if cmd.Flags().Changed("method") {
		current.Method = strings.ToUpper(*flags.method)
	}
	if cmd.Flags().Changed("headers") {
		if *flags.headers == "" {
			current.Headers = nil
		} else {
			var headers map[string]string
			if err := json.Unmarshal([]byte(*flags.headers), &headers); err != nil {
				return nil, fmt.Errorf("invalid --headers JSON: %w", err)
			}
			current.Headers = headers
		}
	}

	// Validate method if set
	if current.Method != "" {
		if err := validateWebhookMethod(current.Method); err != nil {
			return nil, err
		}
	}

	// Validate URL
	if err := validateHTTPURL(current.URL); err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	return client.NewWebhookConfig(current.URL, current.Method, current.Headers)
}

// buildUpdatedPagerDutyConfig builds an updated PagerDuty config.
func buildUpdatedPagerDutyConfig(cmd *cobra.Command, flags *channelUpdateFlags, channel *client.Channel) (json.RawMessage, error) {
	// Parse current config
	var current client.PagerDutyChannelConfig
	if err := json.Unmarshal(channel.Config, &current); err != nil {
		return nil, fmt.Errorf("failed to parse current pagerduty config: %w", err)
	}

	// Apply updates
	if cmd.Flags().Changed("routing-key") {
		if *flags.routingKey == "" {
			return nil, fmt.Errorf("--routing-key cannot be empty")
		}
		current.RoutingKey = *flags.routingKey
	}
	if cmd.Flags().Changed("severity") {
		current.Severity = *flags.severity
	}

	// Validate severity if set
	if current.Severity != "" {
		if err := validatePagerDutySeverity(current.Severity); err != nil {
			return nil, err
		}
	}

	return client.NewPagerDutyConfig(current.RoutingKey, current.Severity)
}

// buildUpdatedDiscordConfig builds an updated Discord config.
func buildUpdatedDiscordConfig(cmd *cobra.Command, flags *channelUpdateFlags, channel *client.Channel) (json.RawMessage, error) {
	// Parse current config
	var current client.DiscordChannelConfig
	if err := json.Unmarshal(channel.Config, &current); err != nil {
		return nil, fmt.Errorf("failed to parse current discord config: %w", err)
	}

	// Apply updates
	if cmd.Flags().Changed("webhook-url") {
		if *flags.webhookURL == "" {
			return nil, fmt.Errorf("--webhook-url cannot be empty")
		}
		current.WebhookURL = *flags.webhookURL
	}

	// Validate and build new config
	return buildDiscordConfig(current.WebhookURL)
}

// buildUpdatedTeamsConfig builds an updated Teams config.
func buildUpdatedTeamsConfig(cmd *cobra.Command, flags *channelUpdateFlags, channel *client.Channel) (json.RawMessage, error) {
	// Parse current config
	var current client.TeamsChannelConfig
	if err := json.Unmarshal(channel.Config, &current); err != nil {
		return nil, fmt.Errorf("failed to parse current teams config: %w", err)
	}

	// Apply updates
	if cmd.Flags().Changed("webhook-url") {
		if *flags.webhookURL == "" {
			return nil, fmt.Errorf("--webhook-url cannot be empty")
		}
		current.WebhookURL = *flags.webhookURL
	}

	// Validate and build new config
	return buildTeamsConfig(current.WebhookURL)
}

// buildUpdatedSMSConfig builds an updated SMS config.
func buildUpdatedSMSConfig(cmd *cobra.Command, flags *channelUpdateFlags, channel *client.Channel) (json.RawMessage, error) {
	// Parse current config
	var current client.SMSChannelConfig
	if err := json.Unmarshal(channel.Config, &current); err != nil {
		return nil, fmt.Errorf("failed to parse current sms config: %w", err)
	}

	// Apply updates
	if cmd.Flags().Changed("phone-number") {
		if *flags.phoneNumber == "" {
			return nil, fmt.Errorf("--phone-number cannot be empty")
		}
		current.PhoneNumber = *flags.phoneNumber
	}

	// Validate and build new config
	return buildSMSConfig(current.PhoneNumber)
}

// runChannelUpdateFromFile executes channel update from a YAML file.
func runChannelUpdateFromFile(ctx context.Context, apiClient *client.Client, channelID uuid.UUID, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %q: %w", filePath, err)
	}

	var cfg channelUpdateYAMLConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	req := &client.UpdateChannelRequest{}
	hasUpdates := false

	// Apply name if provided
	if cfg.Name != nil {
		if *cfg.Name == "" {
			return fmt.Errorf("YAML name cannot be empty when specified")
		}
		req.Name = cfg.Name
		hasUpdates = true
	}

	// Apply enabled if provided
	if cfg.Enabled != nil {
		req.Enabled = cfg.Enabled
		hasUpdates = true
	}

	// Apply config if provided
	if cfg.Config != nil {
		// Fetch current channel to know its type
		reqCtx, cancel := context.WithTimeout(ctx, channelUpdateTimeout)
		channel, err := client.GetChannel(reqCtx, apiClient, channelID)
		cancel()
		if err != nil {
			return fmt.Errorf("failed to get current channel: %w", err)
		}

		config, err := buildConfigFromUpdateYAML(channel, cfg.Config)
		if err != nil {
			return err
		}
		req.Config = config
		hasUpdates = true
	}

	if !hasUpdates {
		return fmt.Errorf("YAML file contains no updates")
	}

	// Call SDK to update channel
	reqCtx, cancel := context.WithTimeout(ctx, channelUpdateTimeout)
	defer cancel()

	channel, err := client.UpdateChannel(reqCtx, apiClient, channelID, req)
	if err != nil {
		return fmt.Errorf("failed to update channel: %w", err)
	}

	return output.Print(channel)
}

// buildConfigFromUpdateYAML builds config from YAML update structure.
func buildConfigFromUpdateYAML(channel *client.Channel, cfg *struct {
	Address     string            `yaml:"address,omitempty"`
	WebhookURL  string            `yaml:"webhook_url,omitempty"`
	URL         string            `yaml:"url,omitempty"`
	Method      string            `yaml:"method,omitempty"`
	Headers     map[string]string `yaml:"headers,omitempty"`
	RoutingKey  string            `yaml:"routing_key,omitempty"`
	Severity    string            `yaml:"severity,omitempty"`
	PhoneNumber string            `yaml:"phone_number,omitempty"`
}) (json.RawMessage, error) {
	switch channel.Type {
	case client.ChannelTypeEmail:
		// Parse current and merge
		var current client.EmailChannelConfig
		if err := json.Unmarshal(channel.Config, &current); err != nil {
			return nil, fmt.Errorf("failed to parse current email config: %w", err)
		}
		if cfg.Address != "" {
			current.Address = cfg.Address
		}
		return buildEmailConfig(current.Address)

	case client.ChannelTypeSlack:
		var current client.SlackChannelConfig
		if err := json.Unmarshal(channel.Config, &current); err != nil {
			return nil, fmt.Errorf("failed to parse current slack config: %w", err)
		}
		if cfg.WebhookURL != "" {
			current.WebhookURL = cfg.WebhookURL
		}
		return buildSlackConfig(current.WebhookURL)

	case client.ChannelTypeWebhook:
		var current client.WebhookChannelConfig
		if err := json.Unmarshal(channel.Config, &current); err != nil {
			return nil, fmt.Errorf("failed to parse current webhook config: %w", err)
		}
		if cfg.URL != "" {
			current.URL = cfg.URL
		}
		if cfg.Method != "" {
			current.Method = strings.ToUpper(cfg.Method)
		}
		if cfg.Headers != nil {
			current.Headers = cfg.Headers
		}
		// Validate
		if current.Method != "" {
			if err := validateWebhookMethod(current.Method); err != nil {
				return nil, err
			}
		}
		if err := validateHTTPURL(current.URL); err != nil {
			return nil, fmt.Errorf("invalid config.url: %w", err)
		}
		return client.NewWebhookConfig(current.URL, current.Method, current.Headers)

	case client.ChannelTypePagerDuty:
		var current client.PagerDutyChannelConfig
		if err := json.Unmarshal(channel.Config, &current); err != nil {
			return nil, fmt.Errorf("failed to parse current pagerduty config: %w", err)
		}
		if cfg.RoutingKey != "" {
			current.RoutingKey = cfg.RoutingKey
		}
		if cfg.Severity != "" {
			current.Severity = cfg.Severity
		}
		if current.Severity != "" {
			if err := validatePagerDutySeverity(current.Severity); err != nil {
				return nil, err
			}
		}
		return client.NewPagerDutyConfig(current.RoutingKey, current.Severity)

	case client.ChannelTypeDiscord:
		var current client.DiscordChannelConfig
		if err := json.Unmarshal(channel.Config, &current); err != nil {
			return nil, fmt.Errorf("failed to parse current discord config: %w", err)
		}
		if cfg.WebhookURL != "" {
			current.WebhookURL = cfg.WebhookURL
		}
		return buildDiscordConfig(current.WebhookURL)

	case client.ChannelTypeTeams:
		var current client.TeamsChannelConfig
		if err := json.Unmarshal(channel.Config, &current); err != nil {
			return nil, fmt.Errorf("failed to parse current teams config: %w", err)
		}
		if cfg.WebhookURL != "" {
			current.WebhookURL = cfg.WebhookURL
		}
		return buildTeamsConfig(current.WebhookURL)

	case client.ChannelTypeSMS:
		var current client.SMSChannelConfig
		if err := json.Unmarshal(channel.Config, &current); err != nil {
			return nil, fmt.Errorf("failed to parse current sms config: %w", err)
		}
		if cfg.PhoneNumber != "" {
			current.PhoneNumber = cfg.PhoneNumber
		}
		return buildSMSConfig(current.PhoneNumber)

	default:
		return nil, fmt.Errorf("unsupported channel type: %s", channel.Type)
	}
}
