// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/mail"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	clierrors "github.com/StackEye-IO/stackeye-cli/internal/errors"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// channelCreateTimeout is the maximum time to wait for the API response.
const channelCreateTimeout = 30 * time.Second

// channelCreateFlags holds the flag values for the channel create command.
type channelCreateFlags struct {
	// Required
	name        string
	channelType string

	// Type-specific: email
	email string

	// Type-specific: slack, discord, teams
	webhookURL string

	// Type-specific: webhook
	url     string
	method  string
	headers string

	// Type-specific: pagerduty
	routingKey string
	severity   string

	// Type-specific: sms
	phoneNumber string

	// Optional
	enabled  bool
	fromFile string
}

// channelYAMLConfig represents the YAML structure for --from-file input.
type channelYAMLConfig struct {
	Name    string `yaml:"name"`
	Type    string `yaml:"type"`
	Enabled *bool  `yaml:"enabled,omitempty"`
	Config  struct {
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
	} `yaml:"config"`
}

// NewChannelCreateCmd creates and returns the channel create subcommand.
func NewChannelCreateCmd() *cobra.Command {
	flags := &channelCreateFlags{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new notification channel",
		Long: `Create a new notification channel for receiving alerts.

Channels define where StackEye sends notifications when probes detect issues.
Each channel type requires specific configuration options.

Required Flags:
  --name       Human-readable name for the channel
  --type       Channel type (see supported types below)

Supported Channel Types:
  email       Send notifications to email addresses
              Requires: --email
  slack       Post to Slack channels via incoming webhooks
              Requires: --webhook-url
  webhook     Send HTTP requests to custom endpoints
              Requires: --url (optional: --method, --headers)
  pagerduty   Create incidents in PagerDuty
              Requires: --routing-key (optional: --severity)
  discord     Post to Discord channels via webhooks
              Requires: --webhook-url
  teams       Post to Microsoft Teams channels
              Requires: --webhook-url
  sms         Send SMS text messages (requires SMS plan)
              Requires: --phone-number

Examples:
  # Create an email channel
  stackeye channel create --name "Ops Team" --type email --email {your-email}

  # Create a Slack channel
  stackeye channel create --name "Alerts" --type slack \
    --webhook-url {slack-webhook-url}

  # Create a webhook channel with custom method and headers
  stackeye channel create --name "Custom API" --type webhook \
    --url {webhook-endpoint-url} \
    --method POST \
    --headers '{"Authorization":"Bearer {your-token}"}'

  # Create a PagerDuty channel
  stackeye channel create --name "On-Call" --type pagerduty \
    --routing-key {pagerduty-routing-key} --severity critical

  # Create a Discord channel
  stackeye channel create --name "Discord Alerts" --type discord \
    --webhook-url {discord-webhook-url}

  # Create a Teams channel
  stackeye channel create --name "Teams Alerts" --type teams \
    --webhook-url {teams-webhook-url}

  # Create an SMS channel
  stackeye channel create --name "Emergency" --type sms --phone-number {phone-e164}

  # Create a channel from YAML file
  stackeye channel create --from-file channel.yaml

  # Create a disabled channel (won't receive notifications until enabled)
  stackeye channel create --name "Staging" --type email --email {your-email} --enabled=false`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChannelCreate(cmd.Context(), flags)
		},
	}

	// Required flags (unless using --from-file)
	cmd.Flags().StringVar(&flags.name, "name", "", "channel name")
	cmd.Flags().StringVar(&flags.channelType, "type", "", "channel type: email, slack, webhook, pagerduty, discord, teams, sms")

	// Type-specific flags
	cmd.Flags().StringVar(&flags.email, "email", "", "email address (for email type)")
	cmd.Flags().StringVar(&flags.webhookURL, "webhook-url", "", "webhook URL (for slack, discord, teams types)")
	cmd.Flags().StringVar(&flags.url, "url", "", "endpoint URL (for webhook type)")
	cmd.Flags().StringVar(&flags.method, "method", "POST", "HTTP method (for webhook type)")
	cmd.Flags().StringVar(&flags.headers, "headers", "", "HTTP headers as JSON object (for webhook type)")
	cmd.Flags().StringVar(&flags.routingKey, "routing-key", "", "routing key (for pagerduty type)")
	cmd.Flags().StringVar(&flags.severity, "severity", "critical", "severity level: critical, error, warning, info (for pagerduty type)")
	cmd.Flags().StringVar(&flags.phoneNumber, "phone-number", "", "phone number in E.164 format (for sms type)")

	// Optional flags
	cmd.Flags().BoolVar(&flags.enabled, "enabled", true, "whether the channel is enabled")
	cmd.Flags().StringVar(&flags.fromFile, "from-file", "", "create channel from YAML file")

	return cmd
}

// runChannelCreate executes the channel create command logic.
func runChannelCreate(ctx context.Context, flags *channelCreateFlags) error {
	var req *client.CreateChannelRequest
	var err error

	// Handle --from-file if provided
	if flags.fromFile != "" {
		req, err = buildRequestFromYAML(flags.fromFile)
		if err != nil {
			return err
		}
	} else {
		// Build request from flags
		req, err = buildRequestFromFlags(flags)
		if err != nil {
			return err
		}
	}

	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Call SDK to create channel with timeout
	reqCtx, cancel := context.WithTimeout(ctx, channelCreateTimeout)
	defer cancel()

	channel, err := client.CreateChannel(reqCtx, apiClient, req)
	if err != nil {
		return fmt.Errorf("failed to create channel: %w", err)
	}

	// Print the created channel using the configured output format
	return output.Print(channel)
}

// buildRequestFromFlags constructs the API request from command flags.
func buildRequestFromFlags(flags *channelCreateFlags) (*client.CreateChannelRequest, error) {
	// Validate required fields
	if flags.name == "" {
		return nil, fmt.Errorf("--name is required")
	}
	if flags.channelType == "" {
		return nil, fmt.Errorf("--type is required")
	}

	// Validate channel type
	channelType := client.ChannelType(strings.ToLower(flags.channelType))
	if err := validateChannelType(channelType); err != nil {
		return nil, err
	}

	// Build type-specific config
	config, err := buildChannelConfig(channelType, flags)
	if err != nil {
		return nil, err
	}

	req := &client.CreateChannelRequest{
		Name:    flags.name,
		Type:    channelType,
		Config:  config,
		Enabled: &flags.enabled,
	}

	return req, nil
}

// buildRequestFromYAML constructs the API request from a YAML file.
func buildRequestFromYAML(filePath string) (*client.CreateChannelRequest, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %q: %w", filePath, err)
	}

	var cfg channelYAMLConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate required fields
	if cfg.Name == "" {
		return nil, fmt.Errorf("YAML config missing required field: name")
	}
	if cfg.Type == "" {
		return nil, fmt.Errorf("YAML config missing required field: type")
	}

	// Validate channel type
	channelType := client.ChannelType(strings.ToLower(cfg.Type))
	if err := validateChannelType(channelType); err != nil {
		return nil, err
	}

	// Build config from YAML
	config, err := buildChannelConfigFromYAML(channelType, &cfg)
	if err != nil {
		return nil, err
	}

	req := &client.CreateChannelRequest{
		Name:    cfg.Name,
		Type:    channelType,
		Config:  config,
		Enabled: cfg.Enabled,
	}

	return req, nil
}

// validateChannelType validates the channel type value.
func validateChannelType(t client.ChannelType) error {
	valid := map[client.ChannelType]bool{
		client.ChannelTypeEmail:     true,
		client.ChannelTypeSlack:     true,
		client.ChannelTypeWebhook:   true,
		client.ChannelTypePagerDuty: true,
		client.ChannelTypeDiscord:   true,
		client.ChannelTypeTeams:     true,
		client.ChannelTypeSMS:       true,
	}
	if !valid[t] {
		return clierrors.InvalidValueError("--type", string(t), clierrors.ValidChannelTypes)
	}
	return nil
}

// buildChannelConfig builds the JSON config for the specified channel type from flags.
func buildChannelConfig(t client.ChannelType, flags *channelCreateFlags) (json.RawMessage, error) {
	switch t {
	case client.ChannelTypeEmail:
		return buildEmailConfig(flags.email)
	case client.ChannelTypeSlack:
		return buildSlackConfig(flags.webhookURL)
	case client.ChannelTypeWebhook:
		return buildWebhookConfig(flags.url, flags.method, flags.headers)
	case client.ChannelTypePagerDuty:
		return buildPagerDutyConfig(flags.routingKey, flags.severity)
	case client.ChannelTypeDiscord:
		return buildDiscordConfig(flags.webhookURL)
	case client.ChannelTypeTeams:
		return buildTeamsConfig(flags.webhookURL)
	case client.ChannelTypeSMS:
		return buildSMSConfig(flags.phoneNumber)
	default:
		return nil, fmt.Errorf("unsupported channel type: %s", t)
	}
}

// buildChannelConfigFromYAML builds the JSON config for the specified channel type from YAML.
func buildChannelConfigFromYAML(t client.ChannelType, cfg *channelYAMLConfig) (json.RawMessage, error) {
	switch t {
	case client.ChannelTypeEmail:
		return buildEmailConfig(cfg.Config.Address)
	case client.ChannelTypeSlack:
		return buildSlackConfig(cfg.Config.WebhookURL)
	case client.ChannelTypeWebhook:
		return buildWebhookConfigFromYAML(cfg.Config.URL, cfg.Config.Method, cfg.Config.Headers)
	case client.ChannelTypePagerDuty:
		return buildPagerDutyConfig(cfg.Config.RoutingKey, cfg.Config.Severity)
	case client.ChannelTypeDiscord:
		return buildDiscordConfig(cfg.Config.WebhookURL)
	case client.ChannelTypeTeams:
		return buildTeamsConfig(cfg.Config.WebhookURL)
	case client.ChannelTypeSMS:
		return buildSMSConfig(cfg.Config.PhoneNumber)
	default:
		return nil, fmt.Errorf("unsupported channel type: %s", t)
	}
}

// buildEmailConfig creates and validates email channel config.
func buildEmailConfig(email string) (json.RawMessage, error) {
	if email == "" {
		return nil, fmt.Errorf("--email is required for email channel type")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, fmt.Errorf("invalid email address %q: %w", email, err)
	}
	return client.NewEmailConfig(email)
}

// buildSlackConfig creates and validates Slack channel config.
func buildSlackConfig(webhookURL string) (json.RawMessage, error) {
	if webhookURL == "" {
		return nil, fmt.Errorf("--webhook-url is required for slack channel type")
	}
	if err := validateWebhookURL(webhookURL, "slack"); err != nil {
		return nil, err
	}
	return client.NewSlackConfig(webhookURL)
}

// buildWebhookConfig creates and validates webhook channel config from flags.
func buildWebhookConfig(webhookURL, method, headersJSON string) (json.RawMessage, error) {
	if webhookURL == "" {
		return nil, fmt.Errorf("--url is required for webhook channel type")
	}
	if err := validateHTTPURL(webhookURL); err != nil {
		return nil, fmt.Errorf("invalid --url: %w", err)
	}

	// Validate method
	method = strings.ToUpper(method)
	if err := validateWebhookMethod(method); err != nil {
		return nil, err
	}

	// Parse headers if provided
	var headers map[string]string
	if headersJSON != "" {
		if err := json.Unmarshal([]byte(headersJSON), &headers); err != nil {
			return nil, fmt.Errorf("invalid --headers JSON: %w", err)
		}
	}

	return client.NewWebhookConfig(webhookURL, method, headers)
}

// buildWebhookConfigFromYAML creates and validates webhook channel config from YAML.
func buildWebhookConfigFromYAML(webhookURL, method string, headers map[string]string) (json.RawMessage, error) {
	if webhookURL == "" {
		return nil, fmt.Errorf("config.url is required for webhook channel type")
	}
	if err := validateHTTPURL(webhookURL); err != nil {
		return nil, fmt.Errorf("invalid config.url: %w", err)
	}

	// Default method to POST if not specified
	if method == "" {
		method = "POST"
	}
	method = strings.ToUpper(method)
	if err := validateWebhookMethod(method); err != nil {
		return nil, err
	}

	return client.NewWebhookConfig(webhookURL, method, headers)
}

// buildPagerDutyConfig creates and validates PagerDuty channel config.
func buildPagerDutyConfig(routingKey, severity string) (json.RawMessage, error) {
	if routingKey == "" {
		return nil, fmt.Errorf("--routing-key is required for pagerduty channel type")
	}
	if err := validatePagerDutySeverity(severity); err != nil {
		return nil, err
	}
	return client.NewPagerDutyConfig(routingKey, severity)
}

// buildDiscordConfig creates and validates Discord channel config.
func buildDiscordConfig(webhookURL string) (json.RawMessage, error) {
	if webhookURL == "" {
		return nil, fmt.Errorf("--webhook-url is required for discord channel type")
	}
	if err := validateWebhookURL(webhookURL, "discord"); err != nil {
		return nil, err
	}
	return client.NewDiscordConfig(webhookURL)
}

// buildTeamsConfig creates and validates Teams channel config.
func buildTeamsConfig(webhookURL string) (json.RawMessage, error) {
	if webhookURL == "" {
		return nil, fmt.Errorf("--webhook-url is required for teams channel type")
	}
	if err := validateWebhookURL(webhookURL, "teams"); err != nil {
		return nil, err
	}
	return client.NewTeamsConfig(webhookURL)
}

// buildSMSConfig creates and validates SMS channel config.
func buildSMSConfig(phoneNumber string) (json.RawMessage, error) {
	if phoneNumber == "" {
		return nil, fmt.Errorf("--phone-number is required for sms channel type")
	}
	if err := validatePhoneNumber(phoneNumber); err != nil {
		return nil, err
	}
	return client.NewSMSConfig(phoneNumber)
}

// validateHTTPURL validates that the URL is a valid HTTP/HTTPS URL.
func validateHTTPURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL %q: %w", rawURL, err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("URL scheme must be http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return fmt.Errorf("URL must include a host: %q", rawURL)
	}
	return nil
}

// validateWebhookURL validates a webhook URL for a specific service.
// We validate the URL format but don't enforce domain restrictions as users
// may use proxies, custom domains, or self-hosted alternatives.
func validateWebhookURL(rawURL, _ string) error {
	return validateHTTPURL(rawURL)
}

// validateWebhookMethod validates the HTTP method for webhook channels.
func validateWebhookMethod(method string) error {
	valid := map[string]bool{
		"GET":    true,
		"POST":   true,
		"PUT":    true,
		"PATCH":  true,
		"DELETE": true,
	}
	if !valid[method] {
		return clierrors.InvalidValueError("--method", method, clierrors.ValidWebhookMethods)
	}
	return nil
}

// validatePagerDutySeverity validates the PagerDuty severity level.
func validatePagerDutySeverity(severity string) error {
	valid := map[string]bool{
		"critical": true,
		"error":    true,
		"warning":  true,
		"info":     true,
	}
	if !valid[strings.ToLower(severity)] {
		return clierrors.InvalidValueError("--severity", severity, clierrors.ValidPagerDutySeverities)
	}
	return nil
}

// validatePhoneNumber validates a phone number in E.164 format.
func validatePhoneNumber(phone string) error {
	// E.164 format: + followed by country code and number, 8-15 digits total
	e164Pattern := regexp.MustCompile(`^\+[1-9]\d{7,14}$`)
	if !e164Pattern.MatchString(phone) {
		return fmt.Errorf("invalid phone number %q: must be in E.164 format (e.g., +15551234567)", phone)
	}
	return nil
}
