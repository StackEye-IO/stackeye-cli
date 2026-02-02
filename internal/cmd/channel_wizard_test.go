package cmd

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/interactive"
)

func TestNewChannelWizardCmd(t *testing.T) {
	cmd := NewChannelWizardCmd()

	if cmd.Use != "wizard" {
		t.Errorf("expected Use='wizard', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Short != "Interactive wizard for creating notification channels" {
		t.Errorf("expected Short='Interactive wizard for creating notification channels', got %q", cmd.Short)
	}

	// Verify aliases
	if len(cmd.Aliases) != 2 {
		t.Errorf("expected 2 aliases, got %d", len(cmd.Aliases))
	}
	expectedAliases := map[string]bool{"wiz": true, "w": true}
	for _, alias := range cmd.Aliases {
		if !expectedAliases[alias] {
			t.Errorf("unexpected alias %q", alias)
		}
	}
}

func TestNewChannelWizardCmd_Long(t *testing.T) {
	cmd := NewChannelWizardCmd()

	long := cmd.Long

	// Should mention all supported channel types
	channelTypes := []string{"email", "slack", "discord", "teams", "webhook", "pagerduty", "sms"}
	for _, ct := range channelTypes {
		if !strings.Contains(long, ct) {
			t.Errorf("expected Long description to mention channel type %q", ct)
		}
	}

	// Should mention the wizard command
	if !strings.Contains(long, "stackeye channel wizard") {
		t.Error("expected Long description to contain example command")
	}

	// Should mention test notification
	if !strings.Contains(long, "test notification") {
		t.Error("expected Long description to mention test notification feature")
	}
}

func TestSupportedChannelTypes(t *testing.T) {
	// Verify all expected channel types are present
	expectedTypes := map[string]bool{
		"email":     false,
		"slack":     false,
		"discord":   false,
		"teams":     false,
		"webhook":   false,
		"pagerduty": false,
		"sms":       false,
	}

	for _, ct := range supportedChannelTypes {
		if _, exists := expectedTypes[ct.value]; !exists {
			t.Errorf("unexpected channel type %q", ct.value)
		}
		expectedTypes[ct.value] = true

		// Each type should have a label and description
		if ct.label == "" {
			t.Errorf("channel type %q missing label", ct.value)
		}
		if ct.description == "" {
			t.Errorf("channel type %q missing description", ct.value)
		}
	}

	// Check all expected types are present
	for typeName, found := range expectedTypes {
		if !found {
			t.Errorf("expected channel type %q not found", typeName)
		}
	}
}

func TestPagerDutySeverities(t *testing.T) {
	expectedSeverities := []string{"critical", "error", "warning", "info"}

	if len(pagerDutySeverities) != len(expectedSeverities) {
		t.Errorf("expected %d severities, got %d", len(expectedSeverities), len(pagerDutySeverities))
	}

	for i, expected := range expectedSeverities {
		if pagerDutySeverities[i] != expected {
			t.Errorf("expected severity[%d]=%q, got %q", i, expected, pagerDutySeverities[i])
		}
	}
}

func TestWebhookMethods(t *testing.T) {
	expectedMethods := []string{"POST", "GET", "PUT", "PATCH", "DELETE"}

	if len(webhookMethods) != len(expectedMethods) {
		t.Errorf("expected %d methods, got %d", len(expectedMethods), len(webhookMethods))
	}

	for i, expected := range expectedMethods {
		if webhookMethods[i] != expected {
			t.Errorf("expected method[%d]=%q, got %q", i, expected, webhookMethods[i])
		}
	}
}

func TestBuildWizardChannelConfig_Email(t *testing.T) {
	wiz := interactive.NewWizard(nil)
	wiz.SetData("email", "test@example.com")

	config, err := buildWizardChannelConfig(wiz, "email")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var emailConfig struct {
		Address string `json:"address"`
	}
	if err := json.Unmarshal(config, &emailConfig); err != nil {
		t.Fatalf("failed to unmarshal config: %v", err)
	}

	if emailConfig.Address != "test@example.com" {
		t.Errorf("expected address='test@example.com', got %q", emailConfig.Address)
	}
}

func TestBuildWizardChannelConfig_Slack(t *testing.T) {
	wiz := interactive.NewWizard(nil)
	wiz.SetData("webhook_url", "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXX")

	config, err := buildWizardChannelConfig(wiz, "slack")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var slackConfig struct {
		WebhookURL string `json:"webhook_url"`
	}
	if err := json.Unmarshal(config, &slackConfig); err != nil {
		t.Fatalf("failed to unmarshal config: %v", err)
	}

	if slackConfig.WebhookURL != "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXX" {
		t.Errorf("unexpected webhook URL: %q", slackConfig.WebhookURL)
	}
}

func TestBuildWizardChannelConfig_Discord(t *testing.T) {
	wiz := interactive.NewWizard(nil)
	wiz.SetData("webhook_url", "https://discord.com/api/webhooks/123456/abcdef")

	config, err := buildWizardChannelConfig(wiz, "discord")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var discordConfig struct {
		WebhookURL string `json:"webhook_url"`
	}
	if err := json.Unmarshal(config, &discordConfig); err != nil {
		t.Fatalf("failed to unmarshal config: %v", err)
	}

	if discordConfig.WebhookURL != "https://discord.com/api/webhooks/123456/abcdef" {
		t.Errorf("unexpected webhook URL: %q", discordConfig.WebhookURL)
	}
}

func TestBuildWizardChannelConfig_Teams(t *testing.T) {
	wiz := interactive.NewWizard(nil)
	wiz.SetData("webhook_url", "https://outlook.office.com/webhook/...")

	config, err := buildWizardChannelConfig(wiz, "teams")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var teamsConfig struct {
		WebhookURL string `json:"webhook_url"`
	}
	if err := json.Unmarshal(config, &teamsConfig); err != nil {
		t.Fatalf("failed to unmarshal config: %v", err)
	}

	if teamsConfig.WebhookURL != "https://outlook.office.com/webhook/..." {
		t.Errorf("unexpected webhook URL: %q", teamsConfig.WebhookURL)
	}
}

func TestBuildWizardChannelConfig_Webhook(t *testing.T) {
	wiz := interactive.NewWizard(nil)
	wiz.SetData("webhook_endpoint_url", "https://api.example.com/notify")
	wiz.SetData("webhook_method", "POST")
	wiz.SetData("webhook_headers", `{"Authorization": "Bearer token123"}`)

	config, err := buildWizardChannelConfig(wiz, "webhook")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var webhookConfig struct {
		URL     string            `json:"url"`
		Method  string            `json:"method"`
		Headers map[string]string `json:"headers"`
	}
	if err := json.Unmarshal(config, &webhookConfig); err != nil {
		t.Fatalf("failed to unmarshal config: %v", err)
	}

	if webhookConfig.URL != "https://api.example.com/notify" {
		t.Errorf("unexpected URL: %q", webhookConfig.URL)
	}
	if webhookConfig.Method != "POST" {
		t.Errorf("expected method='POST', got %q", webhookConfig.Method)
	}
	if webhookConfig.Headers["Authorization"] != "Bearer token123" {
		t.Errorf("unexpected Authorization header: %q", webhookConfig.Headers["Authorization"])
	}
}

func TestBuildWizardChannelConfig_WebhookNoHeaders(t *testing.T) {
	wiz := interactive.NewWizard(nil)
	wiz.SetData("webhook_endpoint_url", "https://api.example.com/notify")
	wiz.SetData("webhook_method", "GET")
	// No headers set

	config, err := buildWizardChannelConfig(wiz, "webhook")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var webhookConfig struct {
		URL    string `json:"url"`
		Method string `json:"method"`
	}
	if err := json.Unmarshal(config, &webhookConfig); err != nil {
		t.Fatalf("failed to unmarshal config: %v", err)
	}

	if webhookConfig.URL != "https://api.example.com/notify" {
		t.Errorf("unexpected URL: %q", webhookConfig.URL)
	}
	if webhookConfig.Method != "GET" {
		t.Errorf("expected method='GET', got %q", webhookConfig.Method)
	}
}

func TestBuildWizardChannelConfig_PagerDuty(t *testing.T) {
	wiz := interactive.NewWizard(nil)
	wiz.SetData("pagerduty_routing_key", "R0123456789ABCDEF")
	wiz.SetData("pagerduty_severity", "critical")

	config, err := buildWizardChannelConfig(wiz, "pagerduty")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var pdConfig struct {
		RoutingKey string `json:"routing_key"`
		Severity   string `json:"severity"`
	}
	if err := json.Unmarshal(config, &pdConfig); err != nil {
		t.Fatalf("failed to unmarshal config: %v", err)
	}

	if pdConfig.RoutingKey != "R0123456789ABCDEF" {
		t.Errorf("unexpected routing key: %q", pdConfig.RoutingKey)
	}
	if pdConfig.Severity != "critical" {
		t.Errorf("expected severity='critical', got %q", pdConfig.Severity)
	}
}

func TestBuildWizardChannelConfig_SMS(t *testing.T) {
	wiz := interactive.NewWizard(nil)
	wiz.SetData("phone_number", "+15551234567")

	config, err := buildWizardChannelConfig(wiz, "sms")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var smsConfig struct {
		PhoneNumber string `json:"phone_number"`
	}
	if err := json.Unmarshal(config, &smsConfig); err != nil {
		t.Fatalf("failed to unmarshal config: %v", err)
	}

	if smsConfig.PhoneNumber != "+15551234567" {
		t.Errorf("unexpected phone number: %q", smsConfig.PhoneNumber)
	}
}

func TestBuildWizardChannelConfig_UnsupportedType(t *testing.T) {
	wiz := interactive.NewWizard(nil)

	_, err := buildWizardChannelConfig(wiz, "invalid-type")
	if err == nil {
		t.Error("expected error for unsupported channel type, got nil")
	}

	if !strings.Contains(err.Error(), "unsupported channel type") {
		t.Errorf("expected error about unsupported type, got: %v", err)
	}
}

func TestBuildWizardChannelConfig_WebhookInvalidHeaders(t *testing.T) {
	wiz := interactive.NewWizard(nil)
	wiz.SetData("webhook_endpoint_url", "https://api.example.com/notify")
	wiz.SetData("webhook_method", "POST")
	wiz.SetData("webhook_headers", "not-valid-json")

	_, err := buildWizardChannelConfig(wiz, "webhook")
	if err == nil {
		t.Error("expected error for invalid headers JSON, got nil")
	}

	if !strings.Contains(err.Error(), "invalid headers JSON") {
		t.Errorf("expected error about invalid JSON, got: %v", err)
	}
}

func TestChannelWizardRegistered(t *testing.T) {
	// Verify the wizard command is registered in the parent channel command
	channelCmd := NewChannelCmd()

	var wizardFound bool
	for _, sub := range channelCmd.Commands() {
		if sub.Use == "wizard" {
			wizardFound = true
			break
		}
	}

	if !wizardFound {
		t.Error("expected wizard command to be registered as subcommand of channel")
	}
}

func TestRunChannelWizardNonInteractive(t *testing.T) {
	// Test that non-interactive mode returns nil (success) and provides guidance
	err := runChannelWizardNonInteractive()
	if err != nil {
		t.Errorf("expected nil error from non-interactive mode, got: %v", err)
	}
}
