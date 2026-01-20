// Package output provides CLI output helpers for StackEye commands.
package output

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
)

// ChannelTableRow represents a row in the channel table output.
// The struct tags control column headers and wide mode display.
type ChannelTableRow struct {
	Status string `table:"STATUS"`
	Name   string `table:"NAME"`
	Type   string `table:"TYPE"`
	Target string `table:"TARGET"`
	Probes string `table:"PROBES"`
	// Wide mode columns
	Created string `table:"CREATED,wide"`
	ID      string `table:"ID,wide"`
}

// ChannelTableFormatter converts SDK Channel types to table-displayable rows
// with status coloring support.
type ChannelTableFormatter struct {
	colorMgr *sdkoutput.ColorManager
	isWide   bool
}

// NewChannelTableFormatter creates a new formatter for channel table output.
// The colorMode parameter controls whether status colors are applied.
// Set isWide to true for extended output with additional columns.
func NewChannelTableFormatter(colorMode sdkoutput.ColorMode, isWide bool) *ChannelTableFormatter {
	return &ChannelTableFormatter{
		colorMgr: sdkoutput.NewColorManager(colorMode),
		isWide:   isWide,
	}
}

// FormatChannels converts a slice of SDK Channels into table-displayable rows.
// Status fields are colored based on channel state:
//   - enabled: green
//   - disabled: red
func (f *ChannelTableFormatter) FormatChannels(channels []client.Channel) []ChannelTableRow {
	rows := make([]ChannelTableRow, 0, len(channels))
	for _, c := range channels {
		rows = append(rows, f.formatChannel(c))
	}
	return rows
}

// FormatChannel converts a single SDK Channel into a table-displayable row.
func (f *ChannelTableFormatter) FormatChannel(channel client.Channel) ChannelTableRow {
	return f.formatChannel(channel)
}

// formatChannel is the internal conversion function.
func (f *ChannelTableFormatter) formatChannel(c client.Channel) ChannelTableRow {
	return ChannelTableRow{
		Status:  f.formatStatus(c.Enabled),
		Name:    c.Name,
		Type:    formatChannelType(c.Type),
		Target:  extractChannelTarget(c.Type, c.Config),
		Probes:  formatProbeCount(c.ProbeCount),
		Created: formatCreatedTime(c.CreatedAt),
		ID:      c.ID.String(),
	}
}

// formatStatus applies color based on channel enabled state.
func (f *ChannelTableFormatter) formatStatus(enabled bool) string {
	if enabled {
		return f.colorMgr.StatusUp("ENABLED")
	}
	return f.colorMgr.StatusDown("DISABLED")
}

// formatChannelType converts the channel type to a human-readable string.
func formatChannelType(channelType client.ChannelType) string {
	switch channelType {
	case client.ChannelTypeEmail:
		return "Email"
	case client.ChannelTypeSlack:
		return "Slack"
	case client.ChannelTypeWebhook:
		return "Webhook"
	case client.ChannelTypePagerDuty:
		return "PagerDuty"
	case client.ChannelTypeDiscord:
		return "Discord"
	case client.ChannelTypeTeams:
		return "Teams"
	case client.ChannelTypeSMS:
		return "SMS"
	default:
		return string(channelType)
	}
}

// extractChannelTarget extracts the target destination from channel config.
// Returns a type-specific target string (email address, webhook URL, phone number, etc.)
func extractChannelTarget(channelType client.ChannelType, config json.RawMessage) string {
	if len(config) == 0 {
		return "-"
	}

	switch channelType {
	case client.ChannelTypeEmail:
		var cfg client.EmailChannelConfig
		if err := json.Unmarshal(config, &cfg); err == nil && cfg.Address != "" {
			return maskEmail(cfg.Address)
		}
	case client.ChannelTypeSlack:
		var cfg client.SlackChannelConfig
		if err := json.Unmarshal(config, &cfg); err == nil && cfg.WebhookURL != "" {
			return truncateWebhookURL(cfg.WebhookURL)
		}
	case client.ChannelTypeWebhook:
		var cfg client.WebhookChannelConfig
		if err := json.Unmarshal(config, &cfg); err == nil && cfg.URL != "" {
			return truncateWebhookURL(cfg.URL)
		}
	case client.ChannelTypePagerDuty:
		var cfg client.PagerDutyChannelConfig
		if err := json.Unmarshal(config, &cfg); err == nil && cfg.RoutingKey != "" {
			return maskRoutingKey(cfg.RoutingKey)
		}
	case client.ChannelTypeDiscord:
		var cfg client.DiscordChannelConfig
		if err := json.Unmarshal(config, &cfg); err == nil && cfg.WebhookURL != "" {
			return truncateWebhookURL(cfg.WebhookURL)
		}
	case client.ChannelTypeTeams:
		var cfg client.TeamsChannelConfig
		if err := json.Unmarshal(config, &cfg); err == nil && cfg.WebhookURL != "" {
			return truncateWebhookURL(cfg.WebhookURL)
		}
	case client.ChannelTypeSMS:
		var cfg client.SMSChannelConfig
		if err := json.Unmarshal(config, &cfg); err == nil && cfg.PhoneNumber != "" {
			return maskPhoneNumber(cfg.PhoneNumber)
		}
	}

	return "-"
}

// maskEmail partially masks an email address for privacy.
// Shows first 2 characters of local part, masks the rest before @.
func maskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}
	local := parts[0]
	domain := parts[1]

	if len(local) <= 2 {
		return local + "***@" + domain
	}
	return local[:2] + "***@" + domain
}

// truncateWebhookURL shortens a webhook URL for display.
// Example: https://hooks.slack.com/services/T00/B00/xxxx -> hooks.slack.com/...
func truncateWebhookURL(url string) string {
	// Remove protocol
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")

	if len(url) <= 30 {
		return url
	}

	// Find the first path segment
	slashIdx := strings.Index(url, "/")
	if slashIdx > 0 && slashIdx < len(url)-1 {
		return url[:slashIdx] + "/..."
	}

	return url[:27] + "..."
}

// maskRoutingKey partially masks a PagerDuty routing key.
// Example: abc123def456 -> abc1***6
func maskRoutingKey(key string) string {
	if len(key) <= 4 {
		return "***"
	}
	return key[:4] + "***" + key[len(key)-1:]
}

// maskPhoneNumber partially masks a phone number for privacy.
// Example: +1234567890 -> +123***7890
func maskPhoneNumber(phone string) string {
	if len(phone) <= 6 {
		return phone
	}
	// Show first 4 and last 4 characters
	return phone[:4] + "***" + phone[len(phone)-4:]
}

// formatProbeCount formats the probe count for display.
func formatProbeCount(count int64) string {
	if count == 0 {
		return "-"
	}
	return fmt.Sprintf("%d", count)
}

// formatCreatedTime formats the created timestamp for display.
func formatCreatedTime(t interface{ Format(string) string }) string {
	return t.Format("2006-01-02")
}

// PrintChannels is a convenience function that formats and prints channels
// using the CLI's configured output format. It handles status coloring
// and wide mode automatically based on configuration.
func PrintChannels(channels []client.Channel) error {
	printer := getPrinter()
	colorMode := sdkoutput.ColorAuto
	isWide := printer.Format() == sdkoutput.FormatWide

	// Get color mode from config if available
	if configGetter != nil {
		if cfg := configGetter(); cfg != nil && cfg.Preferences != nil {
			colorMode = sdkoutput.ColorMode(cfg.Preferences.Color)
		}
	}

	formatter := NewChannelTableFormatter(colorMode, isWide)
	rows := formatter.FormatChannels(channels)

	return printer.Print(rows)
}

// PrintChannel is a convenience function that formats and prints a single channel.
func PrintChannel(channel client.Channel) error {
	printer := getPrinter()
	colorMode := sdkoutput.ColorAuto
	isWide := printer.Format() == sdkoutput.FormatWide

	if configGetter != nil {
		if cfg := configGetter(); cfg != nil && cfg.Preferences != nil {
			colorMode = sdkoutput.ColorMode(cfg.Preferences.Color)
		}
	}

	formatter := NewChannelTableFormatter(colorMode, isWide)
	row := formatter.FormatChannel(channel)

	return printer.Print(row)
}
