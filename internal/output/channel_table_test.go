package output

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestChannelTableFormatter_FormatChannels(t *testing.T) {
	now := time.Now()

	emailConfig, _ := json.Marshal(client.EmailChannelConfig{Address: "alerts@company.io"})
	slackConfig, _ := json.Marshal(client.SlackChannelConfig{WebhookURL: "https://hooks.slack.com/services/T00/B00/xxxx"})
	webhookConfig, _ := json.Marshal(client.WebhookChannelConfig{URL: "https://api.company.io/webhooks/monitoring/alerts"})

	channels := []client.Channel{
		{
			ID:         uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
			Name:       "Email Alerts",
			Type:       client.ChannelTypeEmail,
			Config:     emailConfig,
			Enabled:    true,
			ProbeCount: 5,
			CreatedAt:  now.Add(-30 * 24 * time.Hour),
		},
		{
			ID:         uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
			Name:       "Slack Notifications",
			Type:       client.ChannelTypeSlack,
			Config:     slackConfig,
			Enabled:    true,
			ProbeCount: 12,
			CreatedAt:  now.Add(-7 * 24 * time.Hour),
		},
		{
			ID:         uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc"),
			Name:       "Disabled Webhook",
			Type:       client.ChannelTypeWebhook,
			Config:     webhookConfig,
			Enabled:    false,
			ProbeCount: 0,
			CreatedAt:  now.Add(-1 * 24 * time.Hour),
		},
	}

	// Test with colors disabled for predictable output
	formatter := NewChannelTableFormatter(sdkoutput.ColorNever, false)
	rows := formatter.FormatChannels(channels)

	assert.Len(t, rows, 3)

	// First channel - Email, ENABLED
	assert.Equal(t, "ENABLED", rows[0].Status)
	assert.Equal(t, "Email Alerts", rows[0].Name)
	assert.Equal(t, "Email", rows[0].Type)
	assert.Equal(t, "al***@company.io", rows[0].Target)
	assert.Equal(t, "5", rows[0].Probes)

	// Second channel - Slack, ENABLED
	assert.Equal(t, "ENABLED", rows[1].Status)
	assert.Equal(t, "Slack Notifications", rows[1].Name)
	assert.Equal(t, "Slack", rows[1].Type)
	assert.Equal(t, "hooks.slack.com/...", rows[1].Target)
	assert.Equal(t, "12", rows[1].Probes)

	// Third channel - Webhook, DISABLED
	assert.Equal(t, "DISABLED", rows[2].Status)
	assert.Equal(t, "Disabled Webhook", rows[2].Name)
	assert.Equal(t, "Webhook", rows[2].Type)
	assert.Equal(t, "api.company.io/...", rows[2].Target)
	assert.Equal(t, "-", rows[2].Probes)
}

func TestFormatChannelType(t *testing.T) {
	tests := []struct {
		channelType client.ChannelType
		expected    string
	}{
		{client.ChannelTypeEmail, "Email"},
		{client.ChannelTypeSlack, "Slack"},
		{client.ChannelTypeWebhook, "Webhook"},
		{client.ChannelTypePagerDuty, "PagerDuty"},
		{client.ChannelTypeDiscord, "Discord"},
		{client.ChannelTypeTeams, "Teams"},
		{client.ChannelTypeSMS, "SMS"},
		{client.ChannelType("unknown"), "unknown"},
	}

	for _, tt := range tests {
		t.Run(string(tt.channelType), func(t *testing.T) {
			result := formatChannelType(tt.channelType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractChannelTarget(t *testing.T) {
	tests := []struct {
		name        string
		channelType client.ChannelType
		config      json.RawMessage
		expected    string
	}{
		{
			name:        "empty config",
			channelType: client.ChannelTypeEmail,
			config:      nil,
			expected:    "-",
		},
		{
			name:        "email address",
			channelType: client.ChannelTypeEmail,
			config:      mustMarshal(client.EmailChannelConfig{Address: "alerts@company.io"}),
			expected:    "al***@company.io",
		},
		{
			name:        "slack webhook",
			channelType: client.ChannelTypeSlack,
			config:      mustMarshal(client.SlackChannelConfig{WebhookURL: "https://hooks.slack.com/services/T00/B00/xxxx"}),
			expected:    "hooks.slack.com/...",
		},
		{
			name:        "webhook URL",
			channelType: client.ChannelTypeWebhook,
			config:      mustMarshal(client.WebhookChannelConfig{URL: "https://api.company.io/webhooks/monitoring/alerts"}),
			expected:    "api.company.io/...",
		},
		{
			name:        "pagerduty routing key",
			channelType: client.ChannelTypePagerDuty,
			config:      mustMarshal(client.PagerDutyChannelConfig{RoutingKey: "abc123def456ghi789"}),
			expected:    "abc1***9",
		},
		{
			name:        "discord webhook",
			channelType: client.ChannelTypeDiscord,
			config:      mustMarshal(client.DiscordChannelConfig{WebhookURL: "https://discord.com/api/webhooks/123456/abcdef"}),
			expected:    "discord.com/...",
		},
		{
			name:        "teams webhook",
			channelType: client.ChannelTypeTeams,
			config:      mustMarshal(client.TeamsChannelConfig{WebhookURL: "https://outlook.office.com/webhook/abc123"}),
			expected:    "outlook.office.com/...",
		},
		{
			name:        "sms phone number",
			channelType: client.ChannelTypeSMS,
			config:      mustMarshal(client.SMSChannelConfig{PhoneNumber: "+15551234567"}),
			expected:    "+155***4567",
		},
		{
			name:        "invalid json",
			channelType: client.ChannelTypeEmail,
			config:      json.RawMessage(`{invalid`),
			expected:    "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractChannelTarget(tt.channelType, tt.config)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMaskEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{"standard email", "user@domain.com", "us***@domain.com"},
		{"short local part", "ab@domain.com", "ab***@domain.com"},
		{"single char local", "a@domain.com", "a***@domain.com"},
		{"invalid email no @", "invalid", "invalid"},
		{"long local part", "longusername@domain.com", "lo***@domain.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskEmail(tt.email)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTruncateWebhookURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{"short url", "https://api.io/hook", "api.io/hook"},
		{"slack webhook", "https://hooks.slack.com/services/T00/B00/xxxx", "hooks.slack.com/..."},
		{"discord webhook", "https://discord.com/api/webhooks/123456/abcdef", "discord.com/..."},
		{"no path", "https://shorturl.com", "shorturl.com"},
		{"http url", "http://insecure.com/webhook/monitoring/alerts", "insecure.com/..."},
		{"very long no path", "https://this-is-a-very-long-domain-name-that-exceeds-thirty.com", "this-is-a-very-long-domain-..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateWebhookURL(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMaskRoutingKey(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{"standard key", "abc123def456ghi789", "abc1***9"},
		{"short key", "abcd", "***"},
		{"very short key", "ab", "***"},
		{"exact 4 chars", "abcd", "***"},
		{"5 chars", "abcde", "abcd***e"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskRoutingKey(tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMaskPhoneNumber(t *testing.T) {
	tests := []struct {
		name     string
		phone    string
		expected string
	}{
		{"us number", "+15551234567", "+155***4567"},
		{"short number", "12345", "12345"},
		{"exact 6 chars", "123456", "123456"},
		{"7 chars", "1234567", "1234***4567"},
		{"international", "+44207123456", "+442***3456"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskPhoneNumber(tt.phone)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatProbeCount(t *testing.T) {
	tests := []struct {
		name     string
		count    int64
		expected string
	}{
		{"zero", 0, "-"},
		{"one", 1, "1"},
		{"many", 42, "42"},
		{"large number", 1000, "1000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatProbeCount(tt.count)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatCreatedTime(t *testing.T) {
	testTime := time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC)
	result := formatCreatedTime(testTime)
	assert.Equal(t, "2025-06-15", result)
}

func TestChannelTableFormatter_StatusColoring(t *testing.T) {
	formatter := NewChannelTableFormatter(sdkoutput.ColorAlways, false)

	tests := []struct {
		name    string
		enabled bool
	}{
		{"enabled", true},
		{"disabled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			channel := client.Channel{
				ID:        uuid.New(),
				Name:      "Test Channel",
				Type:      client.ChannelTypeEmail,
				Config:    mustMarshal(client.EmailChannelConfig{Address: "test@test.io"}),
				Enabled:   tt.enabled,
				CreatedAt: time.Now(),
			}

			row := formatter.FormatChannel(channel)

			// Status should contain ANSI escape codes when colors enabled
			assert.Contains(t, row.Status, "\x1b[")
		})
	}
}

func TestChannelTableFormatter_NoColor(t *testing.T) {
	formatter := NewChannelTableFormatter(sdkoutput.ColorNever, false)

	channel := client.Channel{
		ID:        uuid.New(),
		Name:      "Test Channel",
		Type:      client.ChannelTypeSlack,
		Config:    mustMarshal(client.SlackChannelConfig{WebhookURL: "https://hooks.slack.com/test"}),
		Enabled:   true,
		CreatedAt: time.Now(),
	}

	row := formatter.FormatChannel(channel)

	// When colors are disabled, should be plain text without ANSI codes
	assert.Equal(t, "ENABLED", row.Status)
	assert.NotContains(t, row.Status, "\x1b[")
}

func TestNewChannelTableFormatter(t *testing.T) {
	// Test constructor creates valid formatter
	formatter := NewChannelTableFormatter(sdkoutput.ColorAuto, true)

	assert.NotNil(t, formatter)
	assert.NotNil(t, formatter.colorMgr)
	assert.True(t, formatter.isWide)

	formatterNoWide := NewChannelTableFormatter(sdkoutput.ColorNever, false)
	assert.False(t, formatterNoWide.isWide)
}

func TestChannelTableFormatter_WideMode(t *testing.T) {
	now := time.Now()
	channel := client.Channel{
		ID:        uuid.MustParse("12345678-1234-1234-1234-123456789abc"),
		Name:      "Wide Test",
		Type:      client.ChannelTypeWebhook,
		Config:    mustMarshal(client.WebhookChannelConfig{URL: "https://api.test.io/webhook"}),
		Enabled:   true,
		CreatedAt: now,
	}

	formatter := NewChannelTableFormatter(sdkoutput.ColorNever, true)
	row := formatter.FormatChannel(channel)

	// Wide mode should include ID and Created fields
	assert.Equal(t, "12345678-1234-1234-1234-123456789abc", row.ID)
	assert.Equal(t, now.Format("2006-01-02"), row.Created)
}

func TestChannelTableFormatter_AllChannelTypes(t *testing.T) {
	formatter := NewChannelTableFormatter(sdkoutput.ColorNever, false)

	channelTypes := []struct {
		channelType client.ChannelType
		config      json.RawMessage
		expected    string
	}{
		{client.ChannelTypeEmail, mustMarshal(client.EmailChannelConfig{Address: "test@test.io"}), "Email"},
		{client.ChannelTypeSlack, mustMarshal(client.SlackChannelConfig{WebhookURL: "https://hooks.slack.com/test"}), "Slack"},
		{client.ChannelTypeWebhook, mustMarshal(client.WebhookChannelConfig{URL: "https://api.test.io/hook"}), "Webhook"},
		{client.ChannelTypePagerDuty, mustMarshal(client.PagerDutyChannelConfig{RoutingKey: "routing123key"}), "PagerDuty"},
		{client.ChannelTypeDiscord, mustMarshal(client.DiscordChannelConfig{WebhookURL: "https://discord.com/api/webhooks/123"}), "Discord"},
		{client.ChannelTypeTeams, mustMarshal(client.TeamsChannelConfig{WebhookURL: "https://outlook.office.com/webhook/123"}), "Teams"},
		{client.ChannelTypeSMS, mustMarshal(client.SMSChannelConfig{PhoneNumber: "+15551234567"}), "SMS"},
	}

	for _, ct := range channelTypes {
		t.Run(string(ct.channelType), func(t *testing.T) {
			channel := client.Channel{
				ID:        uuid.New(),
				Name:      "Test " + string(ct.channelType),
				Type:      ct.channelType,
				Config:    ct.config,
				Enabled:   true,
				CreatedAt: time.Now(),
			}

			row := formatter.FormatChannel(channel)
			assert.Equal(t, ct.expected, row.Type)
			assert.NotEqual(t, "-", row.Target, "target should be extracted for %s", ct.channelType)
		})
	}
}

// Helper function to marshal config structs
func mustMarshal(v any) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}
