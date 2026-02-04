package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
)

func TestNewChannelCreateCmd(t *testing.T) {
	cmd := NewChannelCreateCmd()

	if cmd.Use != "create" {
		t.Errorf("expected Use='create', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Short != "Create a new notification channel" {
		t.Errorf("expected Short='Create a new notification channel', got %q", cmd.Short)
	}
}

func TestNewChannelCreateCmd_Flags(t *testing.T) {
	cmd := NewChannelCreateCmd()

	// Required flags
	requiredFlags := []string{"name", "type"}
	for _, name := range requiredFlags {
		flag := cmd.Flags().Lookup(name)
		if flag == nil {
			t.Errorf("expected flag --%s to exist", name)
		}
	}

	// Type-specific flags
	typeFlags := []string{
		"email", "webhook-url", "url", "method", "headers",
		"routing-key", "severity", "phone-number",
	}
	for _, name := range typeFlags {
		flag := cmd.Flags().Lookup(name)
		if flag == nil {
			t.Errorf("expected flag --%s to exist", name)
		}
	}

	// Optional flags
	optionalFlags := []string{"enabled", "from-file"}
	for _, name := range optionalFlags {
		flag := cmd.Flags().Lookup(name)
		if flag == nil {
			t.Errorf("expected flag --%s to exist", name)
		}
	}
}

func TestNewChannelCreateCmd_Long(t *testing.T) {
	cmd := NewChannelCreateCmd()

	long := cmd.Long

	// Should contain all channel types
	channelTypes := []string{"email", "slack", "webhook", "pagerduty", "discord", "teams", "sms"}
	for _, channelType := range channelTypes {
		if !strings.Contains(long, channelType) {
			t.Errorf("expected Long description to mention channel type %q", channelType)
		}
	}

	// Should have usage examples
	if !strings.Contains(long, "stackeye channel create") {
		t.Error("expected Long description to contain example commands")
	}

	// Should mention YAML file option
	if !strings.Contains(long, "--from-file") {
		t.Error("expected Long description to mention --from-file option")
	}
}

func TestValidateChannelType(t *testing.T) {
	validTypes := []client.ChannelType{
		client.ChannelTypeEmail,
		client.ChannelTypeSlack,
		client.ChannelTypeWebhook,
		client.ChannelTypePagerDuty,
		client.ChannelTypeDiscord,
		client.ChannelTypeTeams,
		client.ChannelTypeSMS,
	}

	for _, ct := range validTypes {
		t.Run(string(ct), func(t *testing.T) {
			err := validateChannelType(ct)
			if err != nil {
				t.Errorf("expected valid channel type %q to pass validation, got error: %v", ct, err)
			}
		})
	}
}

func TestValidateChannelType_Invalid(t *testing.T) {
	invalidTypes := []string{"invalid", "http", "EMAIL", "SLACK", ""}
	for _, ct := range invalidTypes {
		t.Run(ct, func(t *testing.T) {
			err := validateChannelType(client.ChannelType(ct))
			if err == nil {
				t.Errorf("expected invalid channel type %q to fail validation", ct)
			}
			if !strings.Contains(err.Error(), "for --type") {
				t.Errorf("expected error to mention 'for --type', got: %v", err)
			}
		})
	}
}

func TestValidateHTTPURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid https URL",
			url:     "https://hooks.slack.com/services/xxx",
			wantErr: false,
		},
		{
			name:    "valid http URL",
			url:     "http://localhost:8080/webhook",
			wantErr: false,
		},
		{
			name:    "missing scheme",
			url:     "hooks.slack.com/services/xxx",
			wantErr: true,
			errMsg:  "URL scheme must be http or https",
		},
		{
			name:    "invalid scheme",
			url:     "ftp://example.com/webhook",
			wantErr: true,
			errMsg:  "URL scheme must be http or https",
		},
		{
			name:    "missing host",
			url:     "https:///path",
			wantErr: true,
			errMsg:  "URL must include a host",
		},
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
			errMsg:  "URL scheme must be http or https",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateHTTPURL(tt.url)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for URL %q", tt.url)
					return
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing %q, got: %v", tt.errMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for URL %q: %v", tt.url, err)
				}
			}
		})
	}
}

func TestValidateWebhookMethod(t *testing.T) {
	validMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE"}
	for _, method := range validMethods {
		t.Run(method, func(t *testing.T) {
			err := validateWebhookMethod(method)
			if err != nil {
				t.Errorf("expected valid method %q to pass, got error: %v", method, err)
			}
		})
	}
}

func TestValidateWebhookMethod_Invalid(t *testing.T) {
	invalidMethods := []string{"HEAD", "OPTIONS", "CONNECT", "post", "get", ""}
	for _, method := range invalidMethods {
		t.Run(method, func(t *testing.T) {
			err := validateWebhookMethod(method)
			if err == nil {
				t.Errorf("expected invalid method %q to fail", method)
			}
		})
	}
}

func TestValidatePagerDutySeverity(t *testing.T) {
	validSeverities := []string{"critical", "error", "warning", "info"}
	for _, severity := range validSeverities {
		t.Run(severity, func(t *testing.T) {
			err := validatePagerDutySeverity(severity)
			if err != nil {
				t.Errorf("expected valid severity %q to pass, got error: %v", severity, err)
			}
		})
	}
}

func TestValidatePagerDutySeverity_CaseInsensitive(t *testing.T) {
	// Severity validation is case-insensitive
	validSeverities := []string{"CRITICAL", "ERROR", "WARNING", "INFO", "Critical", "Error"}
	for _, severity := range validSeverities {
		t.Run(severity, func(t *testing.T) {
			err := validatePagerDutySeverity(severity)
			if err != nil {
				t.Errorf("expected severity %q to pass (case-insensitive), got error: %v", severity, err)
			}
		})
	}
}

func TestValidatePagerDutySeverity_Invalid(t *testing.T) {
	invalidSeverities := []string{"high", "low", "medium", ""}
	for _, severity := range invalidSeverities {
		t.Run(severity, func(t *testing.T) {
			err := validatePagerDutySeverity(severity)
			if err == nil {
				t.Errorf("expected invalid severity %q to fail", severity)
			}
		})
	}
}

func TestValidatePhoneNumber(t *testing.T) {
	tests := []struct {
		name    string
		phone   string
		wantErr bool
	}{
		{
			name:    "valid US number",
			phone:   "+15551234567",
			wantErr: false,
		},
		{
			name:    "valid UK number",
			phone:   "+447911123456",
			wantErr: false,
		},
		{
			name:    "valid long number",
			phone:   "+861234567890123",
			wantErr: false,
		},
		{
			name:    "missing plus",
			phone:   "15551234567",
			wantErr: true,
		},
		{
			name:    "too short",
			phone:   "+1234567",
			wantErr: true,
		},
		{
			name:    "too long",
			phone:   "+12345678901234567",
			wantErr: true,
		},
		{
			name:    "leading zero in country code",
			phone:   "+05551234567",
			wantErr: true,
		},
		{
			name:    "contains letters",
			phone:   "+1555ABCDEFG",
			wantErr: true,
		},
		{
			name:    "empty",
			phone:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePhoneNumber(tt.phone)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for phone %q", tt.phone)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for phone %q: %v", tt.phone, err)
				}
			}
		})
	}
}

func TestBuildEmailConfig(t *testing.T) {
	t.Run("valid email", func(t *testing.T) {
		config, err := buildEmailConfig("test@stackeye.io")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if config == nil {
			t.Error("expected config to be non-nil")
		}
		// Verify JSON structure
		if !strings.Contains(string(config), `"address":"test@stackeye.io"`) {
			t.Errorf("expected config to contain address, got: %s", string(config))
		}
	})

	t.Run("empty email", func(t *testing.T) {
		_, err := buildEmailConfig("")
		if err == nil {
			t.Error("expected error for empty email")
		}
		if !strings.Contains(err.Error(), "--email is required") {
			t.Errorf("expected error about required email, got: %v", err)
		}
	})

	t.Run("invalid email format", func(t *testing.T) {
		_, err := buildEmailConfig("not-an-email")
		if err == nil {
			t.Error("expected error for invalid email")
		}
		if !strings.Contains(err.Error(), "invalid email address") {
			t.Errorf("expected error about invalid email, got: %v", err)
		}
	})
}

func TestBuildSlackConfig(t *testing.T) {
	t.Run("valid webhook URL", func(t *testing.T) {
		config, err := buildSlackConfig("https://hooks.slack.com/services/T00/B00/xxx")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if config == nil {
			t.Error("expected config to be non-nil")
		}
		if !strings.Contains(string(config), `"webhook_url"`) {
			t.Errorf("expected config to contain webhook_url, got: %s", string(config))
		}
	})

	t.Run("empty webhook URL", func(t *testing.T) {
		_, err := buildSlackConfig("")
		if err == nil {
			t.Error("expected error for empty webhook URL")
		}
		if !strings.Contains(err.Error(), "--webhook-url is required") {
			t.Errorf("expected error about required webhook URL, got: %v", err)
		}
	})

	t.Run("invalid URL scheme", func(t *testing.T) {
		_, err := buildSlackConfig("ftp://hooks.slack.com/services/xxx")
		if err == nil {
			t.Error("expected error for invalid URL scheme")
		}
	})
}

func TestBuildWebhookConfig(t *testing.T) {
	t.Run("valid config with headers", func(t *testing.T) {
		config, err := buildWebhookConfig(
			"https://api.example.com/webhook",
			"POST",
			`{"Authorization":"Bearer token"}`,
		)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if config == nil {
			t.Error("expected config to be non-nil")
		}
		configStr := string(config)
		if !strings.Contains(configStr, `"url"`) {
			t.Errorf("expected config to contain url, got: %s", configStr)
		}
		if !strings.Contains(configStr, `"method":"POST"`) {
			t.Errorf("expected config to contain method POST, got: %s", configStr)
		}
		if !strings.Contains(configStr, `"headers"`) {
			t.Errorf("expected config to contain headers, got: %s", configStr)
		}
	})

	t.Run("valid config without headers", func(t *testing.T) {
		config, err := buildWebhookConfig("https://api.example.com/webhook", "GET", "")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if config == nil {
			t.Error("expected config to be non-nil")
		}
	})

	t.Run("empty URL", func(t *testing.T) {
		_, err := buildWebhookConfig("", "POST", "")
		if err == nil {
			t.Error("expected error for empty URL")
		}
		if !strings.Contains(err.Error(), "--url is required") {
			t.Errorf("expected error about required URL, got: %v", err)
		}
	})

	t.Run("invalid method", func(t *testing.T) {
		_, err := buildWebhookConfig("https://api.example.com/webhook", "INVALID", "")
		if err == nil {
			t.Error("expected error for invalid method")
		}
	})

	t.Run("invalid headers JSON", func(t *testing.T) {
		_, err := buildWebhookConfig("https://api.example.com/webhook", "POST", "not-json")
		if err == nil {
			t.Error("expected error for invalid headers JSON")
		}
		if !strings.Contains(err.Error(), "invalid --headers JSON") {
			t.Errorf("expected error about invalid headers JSON, got: %v", err)
		}
	})
}

func TestBuildPagerDutyConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config, err := buildPagerDutyConfig("routing-key-123", "critical")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if config == nil {
			t.Error("expected config to be non-nil")
		}
		configStr := string(config)
		if !strings.Contains(configStr, `"routing_key":"routing-key-123"`) {
			t.Errorf("expected config to contain routing_key, got: %s", configStr)
		}
		if !strings.Contains(configStr, `"severity":"critical"`) {
			t.Errorf("expected config to contain severity, got: %s", configStr)
		}
	})

	t.Run("empty routing key", func(t *testing.T) {
		_, err := buildPagerDutyConfig("", "critical")
		if err == nil {
			t.Error("expected error for empty routing key")
		}
		if !strings.Contains(err.Error(), "--routing-key is required") {
			t.Errorf("expected error about required routing key, got: %v", err)
		}
	})

	t.Run("invalid severity", func(t *testing.T) {
		_, err := buildPagerDutyConfig("routing-key-123", "invalid")
		if err == nil {
			t.Error("expected error for invalid severity")
		}
	})
}

func TestBuildDiscordConfig(t *testing.T) {
	t.Run("valid webhook URL", func(t *testing.T) {
		config, err := buildDiscordConfig("https://discord.com/api/webhooks/xxx/yyy")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if config == nil {
			t.Error("expected config to be non-nil")
		}
	})

	t.Run("empty webhook URL", func(t *testing.T) {
		_, err := buildDiscordConfig("")
		if err == nil {
			t.Error("expected error for empty webhook URL")
		}
		if !strings.Contains(err.Error(), "--webhook-url is required for discord") {
			t.Errorf("expected error about required webhook URL, got: %v", err)
		}
	})
}

func TestBuildTeamsConfig(t *testing.T) {
	t.Run("valid webhook URL", func(t *testing.T) {
		config, err := buildTeamsConfig("https://outlook.office.com/webhook/xxx")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if config == nil {
			t.Error("expected config to be non-nil")
		}
	})

	t.Run("empty webhook URL", func(t *testing.T) {
		_, err := buildTeamsConfig("")
		if err == nil {
			t.Error("expected error for empty webhook URL")
		}
		if !strings.Contains(err.Error(), "--webhook-url is required for teams") {
			t.Errorf("expected error about required webhook URL, got: %v", err)
		}
	})
}

func TestBuildSMSConfig(t *testing.T) {
	t.Run("valid phone number", func(t *testing.T) {
		config, err := buildSMSConfig("+15551234567")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if config == nil {
			t.Error("expected config to be non-nil")
		}
		if !strings.Contains(string(config), `"phone_number":"+15551234567"`) {
			t.Errorf("expected config to contain phone_number, got: %s", string(config))
		}
	})

	t.Run("empty phone number", func(t *testing.T) {
		_, err := buildSMSConfig("")
		if err == nil {
			t.Error("expected error for empty phone number")
		}
		if !strings.Contains(err.Error(), "--phone-number is required") {
			t.Errorf("expected error about required phone number, got: %v", err)
		}
	})

	t.Run("invalid phone number format", func(t *testing.T) {
		_, err := buildSMSConfig("5551234567")
		if err == nil {
			t.Error("expected error for invalid phone number")
		}
		if !strings.Contains(err.Error(), "invalid phone number") {
			t.Errorf("expected error about invalid phone number, got: %v", err)
		}
	})
}

func TestBuildRequestFromFlags_Validation(t *testing.T) {
	t.Run("missing name", func(t *testing.T) {
		flags := &channelCreateFlags{
			channelType: "email",
			email:       "test@stackeye.io",
		}
		_, err := buildRequestFromFlags(flags)
		if err == nil {
			t.Error("expected error for missing name")
		}
		if !strings.Contains(err.Error(), "--name is required") {
			t.Errorf("expected error about required name, got: %v", err)
		}
	})

	t.Run("missing type", func(t *testing.T) {
		flags := &channelCreateFlags{
			name:  "Test Channel",
			email: "test@stackeye.io",
		}
		_, err := buildRequestFromFlags(flags)
		if err == nil {
			t.Error("expected error for missing type")
		}
		if !strings.Contains(err.Error(), "--type is required") {
			t.Errorf("expected error about required type, got: %v", err)
		}
	})

	t.Run("invalid type", func(t *testing.T) {
		flags := &channelCreateFlags{
			name:        "Test Channel",
			channelType: "invalid",
		}
		_, err := buildRequestFromFlags(flags)
		if err == nil {
			t.Error("expected error for invalid type")
		}
	})

	t.Run("valid email channel", func(t *testing.T) {
		flags := &channelCreateFlags{
			name:        "Test Email",
			channelType: "email",
			email:       "test@stackeye.io",
			enabled:     true,
		}
		req, err := buildRequestFromFlags(flags)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if req == nil {
			t.Fatal("expected request to be non-nil")
		}
		if req.Name != "Test Email" {
			t.Errorf("expected name 'Test Email', got %q", req.Name)
		}
		if req.Type != client.ChannelTypeEmail {
			t.Errorf("expected type email, got %q", req.Type)
		}
	})

	t.Run("valid slack channel", func(t *testing.T) {
		flags := &channelCreateFlags{
			name:        "Test Slack",
			channelType: "slack",
			webhookURL:  "https://hooks.slack.com/services/xxx",
			enabled:     true,
		}
		req, err := buildRequestFromFlags(flags)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if req == nil {
			t.Fatal("expected request to be non-nil")
		}
		if req.Type != client.ChannelTypeSlack {
			t.Errorf("expected type slack, got %q", req.Type)
		}
	})

	t.Run("channel type case insensitive", func(t *testing.T) {
		flags := &channelCreateFlags{
			name:        "Test Email",
			channelType: "EMAIL",
			email:       "test@stackeye.io",
			enabled:     true,
		}
		req, err := buildRequestFromFlags(flags)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if req.Type != client.ChannelTypeEmail {
			t.Errorf("expected type email, got %q", req.Type)
		}
	})
}

// writeYAMLFile is a test helper that writes YAML content to a temp file and returns the path.
func writeYAMLFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "channel.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp YAML file: %v", err)
	}
	return path
}

func TestBuildRequestFromYAML_Email(t *testing.T) {
	path := writeYAMLFile(t, `
name: Ops Email
type: email
config:
  address: ops@stackeye.io
`)
	req, err := buildRequestFromYAML(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Name != "Ops Email" {
		t.Errorf("expected name 'Ops Email', got %q", req.Name)
	}
	if req.Type != client.ChannelTypeEmail {
		t.Errorf("expected type email, got %q", req.Type)
	}
	if !strings.Contains(string(req.Config), `"address":"ops@stackeye.io"`) {
		t.Errorf("expected config to contain address, got: %s", string(req.Config))
	}
	if req.Enabled != nil {
		t.Errorf("expected Enabled to be nil when omitted, got %v", *req.Enabled)
	}
}

func TestBuildRequestFromYAML_Slack(t *testing.T) {
	path := writeYAMLFile(t, `
name: Slack Alerts
type: slack
config:
  webhook_url: https://hooks.slack.com/services/T00/B00/xxx
`)
	req, err := buildRequestFromYAML(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Type != client.ChannelTypeSlack {
		t.Errorf("expected type slack, got %q", req.Type)
	}
	if !strings.Contains(string(req.Config), `"webhook_url"`) {
		t.Errorf("expected config to contain webhook_url, got: %s", string(req.Config))
	}
}

func TestBuildRequestFromYAML_Webhook(t *testing.T) {
	path := writeYAMLFile(t, `
name: Custom Webhook
type: webhook
config:
  url: https://api.example.com/hook
  method: PUT
  headers:
    Authorization: Bearer token123
    Content-Type: application/json
`)
	req, err := buildRequestFromYAML(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Type != client.ChannelTypeWebhook {
		t.Errorf("expected type webhook, got %q", req.Type)
	}
	configStr := string(req.Config)
	if !strings.Contains(configStr, `"url":"https://api.example.com/hook"`) {
		t.Errorf("expected config to contain url, got: %s", configStr)
	}
	if !strings.Contains(configStr, `"method":"PUT"`) {
		t.Errorf("expected config to contain method PUT, got: %s", configStr)
	}
	if !strings.Contains(configStr, `"headers"`) {
		t.Errorf("expected config to contain headers, got: %s", configStr)
	}
}

func TestBuildRequestFromYAML_WebhookDefaultMethod(t *testing.T) {
	path := writeYAMLFile(t, `
name: Webhook No Method
type: webhook
config:
  url: https://api.example.com/hook
`)
	req, err := buildRequestFromYAML(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(req.Config), `"method":"POST"`) {
		t.Errorf("expected default method POST, got: %s", string(req.Config))
	}
}

func TestBuildRequestFromYAML_PagerDuty(t *testing.T) {
	path := writeYAMLFile(t, `
name: On-Call PD
type: pagerduty
config:
  routing_key: abc123def456
  severity: warning
`)
	req, err := buildRequestFromYAML(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Type != client.ChannelTypePagerDuty {
		t.Errorf("expected type pagerduty, got %q", req.Type)
	}
	configStr := string(req.Config)
	if !strings.Contains(configStr, `"routing_key":"abc123def456"`) {
		t.Errorf("expected config to contain routing_key, got: %s", configStr)
	}
	if !strings.Contains(configStr, `"severity":"warning"`) {
		t.Errorf("expected config to contain severity, got: %s", configStr)
	}
}

func TestBuildRequestFromYAML_Discord(t *testing.T) {
	path := writeYAMLFile(t, `
name: Discord Alerts
type: discord
config:
  webhook_url: https://discord.com/api/webhooks/xxx/yyy
`)
	req, err := buildRequestFromYAML(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Type != client.ChannelTypeDiscord {
		t.Errorf("expected type discord, got %q", req.Type)
	}
}

func TestBuildRequestFromYAML_Teams(t *testing.T) {
	path := writeYAMLFile(t, `
name: Teams Alerts
type: teams
config:
  webhook_url: https://outlook.office.com/webhook/xxx
`)
	req, err := buildRequestFromYAML(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Type != client.ChannelTypeTeams {
		t.Errorf("expected type teams, got %q", req.Type)
	}
}

func TestBuildRequestFromYAML_SMS(t *testing.T) {
	path := writeYAMLFile(t, `
name: Emergency SMS
type: sms
config:
  phone_number: "+15551234567"
`)
	req, err := buildRequestFromYAML(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Type != client.ChannelTypeSMS {
		t.Errorf("expected type sms, got %q", req.Type)
	}
	if !strings.Contains(string(req.Config), `"phone_number":"+15551234567"`) {
		t.Errorf("expected config to contain phone_number, got: %s", string(req.Config))
	}
}

func TestBuildRequestFromYAML_MissingName(t *testing.T) {
	path := writeYAMLFile(t, `
type: email
config:
  address: ops@stackeye.io
`)
	_, err := buildRequestFromYAML(path)
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if !strings.Contains(err.Error(), "missing required field: name") {
		t.Errorf("expected error about missing name, got: %v", err)
	}
}

func TestBuildRequestFromYAML_MissingType(t *testing.T) {
	path := writeYAMLFile(t, `
name: Test Channel
config:
  address: ops@stackeye.io
`)
	_, err := buildRequestFromYAML(path)
	if err == nil {
		t.Fatal("expected error for missing type")
	}
	if !strings.Contains(err.Error(), "missing required field: type") {
		t.Errorf("expected error about missing type, got: %v", err)
	}
}

func TestBuildRequestFromYAML_InvalidType(t *testing.T) {
	path := writeYAMLFile(t, `
name: Bad Channel
type: carrier_pigeon
config: {}
`)
	_, err := buildRequestFromYAML(path)
	if err == nil {
		t.Fatal("expected error for invalid channel type")
	}
	if !strings.Contains(err.Error(), "for --type") {
		t.Errorf("expected error mentioning invalid type, got: %v", err)
	}
}

func TestBuildRequestFromYAML_MalformedYAML(t *testing.T) {
	path := writeYAMLFile(t, `
name: "unterminated
  [this is not valid yaml
  !!!
`)
	_, err := buildRequestFromYAML(path)
	if err == nil {
		t.Fatal("expected error for malformed YAML")
	}
	if !strings.Contains(err.Error(), "failed to parse YAML") {
		t.Errorf("expected YAML parse error, got: %v", err)
	}
}

func TestBuildRequestFromYAML_FileNotFound(t *testing.T) {
	_, err := buildRequestFromYAML("/nonexistent/path/channel.yaml")
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}
	if !strings.Contains(err.Error(), "failed to read file") {
		t.Errorf("expected file read error, got: %v", err)
	}
}

func TestBuildRequestFromYAML_EmptyFile(t *testing.T) {
	path := writeYAMLFile(t, "")
	_, err := buildRequestFromYAML(path)
	if err == nil {
		t.Fatal("expected error for empty file")
	}
	if !strings.Contains(err.Error(), "missing required field") {
		t.Errorf("expected missing field error, got: %v", err)
	}
}

func TestBuildRequestFromYAML_EnabledExplicit(t *testing.T) {
	path := writeYAMLFile(t, `
name: Disabled Channel
type: email
enabled: false
config:
  address: ops@stackeye.io
`)
	req, err := buildRequestFromYAML(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Enabled == nil {
		t.Fatal("expected Enabled to be non-nil when explicitly set")
	}
	if *req.Enabled != false {
		t.Errorf("expected Enabled to be false, got %v", *req.Enabled)
	}
}

func TestBuildRequestFromYAML_EnabledTrue(t *testing.T) {
	path := writeYAMLFile(t, `
name: Enabled Channel
type: email
enabled: true
config:
  address: ops@stackeye.io
`)
	req, err := buildRequestFromYAML(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Enabled == nil {
		t.Fatal("expected Enabled to be non-nil when explicitly set")
	}
	if *req.Enabled != true {
		t.Errorf("expected Enabled to be true, got %v", *req.Enabled)
	}
}

func TestBuildRequestFromYAML_TypeCaseInsensitive(t *testing.T) {
	path := writeYAMLFile(t, `
name: Upper Case Type
type: EMAIL
config:
  address: ops@stackeye.io
`)
	req, err := buildRequestFromYAML(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Type != client.ChannelTypeEmail {
		t.Errorf("expected type email, got %q", req.Type)
	}
}

func TestBuildRequestFromYAML_WebhookMissingURL(t *testing.T) {
	path := writeYAMLFile(t, `
name: Bad Webhook
type: webhook
config:
  method: POST
`)
	_, err := buildRequestFromYAML(path)
	if err == nil {
		t.Fatal("expected error for webhook missing URL")
	}
	if !strings.Contains(err.Error(), "config.url is required") {
		t.Errorf("expected error about required URL, got: %v", err)
	}
}

func TestBuildRequestFromYAML_EmailMissingAddress(t *testing.T) {
	path := writeYAMLFile(t, `
name: Bad Email
type: email
config: {}
`)
	_, err := buildRequestFromYAML(path)
	if err == nil {
		t.Fatal("expected error for email missing address")
	}
	if !strings.Contains(err.Error(), "--email is required") {
		t.Errorf("expected error about required email, got: %v", err)
	}
}

func TestBuildRequestFromYAML_SlackMissingWebhookURL(t *testing.T) {
	path := writeYAMLFile(t, `
name: Bad Slack
type: slack
config: {}
`)
	_, err := buildRequestFromYAML(path)
	if err == nil {
		t.Fatal("expected error for slack missing webhook URL")
	}
	if !strings.Contains(err.Error(), "--webhook-url is required") {
		t.Errorf("expected error about required webhook URL, got: %v", err)
	}
}

func TestBuildRequestFromYAML_PagerDutyMissingRoutingKey(t *testing.T) {
	path := writeYAMLFile(t, `
name: Bad PD
type: pagerduty
config:
  severity: critical
`)
	_, err := buildRequestFromYAML(path)
	if err == nil {
		t.Fatal("expected error for pagerduty missing routing key")
	}
	if !strings.Contains(err.Error(), "--routing-key is required") {
		t.Errorf("expected error about required routing key, got: %v", err)
	}
}

func TestBuildRequestFromYAML_SMSMissingPhoneNumber(t *testing.T) {
	path := writeYAMLFile(t, `
name: Bad SMS
type: sms
config: {}
`)
	_, err := buildRequestFromYAML(path)
	if err == nil {
		t.Fatal("expected error for sms missing phone number")
	}
	if !strings.Contains(err.Error(), "--phone-number is required") {
		t.Errorf("expected error about required phone number, got: %v", err)
	}
}
