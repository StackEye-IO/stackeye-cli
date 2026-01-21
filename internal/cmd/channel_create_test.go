package cmd

import (
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
			if !strings.Contains(err.Error(), "invalid --type") {
				t.Errorf("expected error to mention 'invalid --type', got: %v", err)
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
