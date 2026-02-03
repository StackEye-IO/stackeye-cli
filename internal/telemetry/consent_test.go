package telemetry

import (
	"bytes"
	"strings"
	"testing"
)

func TestPromptConsent_EnablesOnYes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"lowercase y", "y\n", true},
		{"lowercase yes", "yes\n", true},
		{"uppercase Y", "Y\n", true},
		{"uppercase YES", "YES\n", true},
		{"mixed case Yes", "Yes\n", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdin := strings.NewReader(tt.input)
			stdout := &bytes.Buffer{}

			enabled, err := promptConsent(stdin, stdout)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if enabled != tt.expected {
				t.Errorf("promptConsent(%q) = %v, want %v", tt.input, enabled, tt.expected)
			}
		})
	}
}

func TestPromptConsent_DisablesOnNo(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"lowercase n", "n\n"},
		{"lowercase no", "no\n"},
		{"uppercase N", "N\n"},
		{"uppercase NO", "NO\n"},
		{"empty input", "\n"},
		{"random text", "maybe\n"},
		{"spaces", "   \n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdin := strings.NewReader(tt.input)
			stdout := &bytes.Buffer{}

			enabled, err := promptConsent(stdin, stdout)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if enabled {
				t.Errorf("promptConsent(%q) = true, want false", tt.input)
			}
		})
	}
}

func TestPromptConsent_OutputsConsentMessage(t *testing.T) {
	stdin := strings.NewReader("n\n")
	stdout := &bytes.Buffer{}

	_, err := promptConsent(stdin, stdout)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := stdout.String()

	// Should contain the consent message
	if !strings.Contains(output, "anonymous usage data") {
		t.Error("output should contain consent message about anonymous usage data")
	}

	// Should contain the prompt
	if !strings.Contains(output, "Enable telemetry?") {
		t.Error("output should contain the enable telemetry prompt")
	}
}

func TestPromptConsent_OutputsEnabledConfirmation(t *testing.T) {
	stdin := strings.NewReader("y\n")
	stdout := &bytes.Buffer{}

	_, err := promptConsent(stdin, stdout)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Telemetry enabled") {
		t.Error("output should confirm telemetry was enabled")
	}
}

func TestPromptConsent_OutputsDisabledConfirmation(t *testing.T) {
	stdin := strings.NewReader("n\n")
	stdout := &bytes.Buffer{}

	_, err := promptConsent(stdin, stdout)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Telemetry disabled") {
		t.Error("output should confirm telemetry was disabled")
	}
}

func TestPromptConsent_ErrorOnReadFailure(t *testing.T) {
	// Empty reader will cause EOF on ReadString
	stdin := strings.NewReader("")
	stdout := &bytes.Buffer{}

	_, err := promptConsent(stdin, stdout)
	if err == nil {
		t.Error("expected error when stdin is empty, got nil")
	}
}

func TestConsentMessage_ContainsPrivacyInfo(t *testing.T) {
	// Verify the consent message contains key privacy information
	message := ConsentMessage

	checks := []struct {
		name     string
		contains string
	}{
		{"anonymous", "anonymous"},
		{"usage data", "usage data"},
		{"command usage", "command"},
		{"no personal data", "personal data"},
		{"no API keys", "API keys"},
	}

	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			if !strings.Contains(strings.ToLower(message), strings.ToLower(check.contains)) {
				t.Errorf("ConsentMessage should contain %q", check.contains)
			}
		})
	}
}

func TestConsentPrompt_Format(t *testing.T) {
	// Verify the prompt has the expected format
	if !strings.Contains(ConsentPrompt, "[y/N]") {
		t.Error("ConsentPrompt should indicate default is N (no)")
	}

	if !strings.HasSuffix(ConsentPrompt, ": ") {
		t.Error("ConsentPrompt should end with ': '")
	}
}

func TestIsInteractive_ReturnsBoolWithoutCrash(t *testing.T) {
	// We can't easily mock os.Stdin in tests, but we can verify
	// the function doesn't panic and returns a bool
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("isInteractive panicked: %v", r)
		}
	}()

	// In test environment, stdin is typically not a terminal
	result := isInteractive()
	_ = result // Just verify it returns without error
}
