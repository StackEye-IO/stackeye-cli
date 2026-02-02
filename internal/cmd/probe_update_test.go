// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"
)

func TestNewProbeUpdateCmd(t *testing.T) {
	cmd := NewProbeUpdateCmd()

	if cmd.Use != "update <id>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "update <id>")
	}

	// Check that all optional flags exist
	optionalFlags := []string{
		"name", "url", "method", "interval", "timeout", "regions",
		"headers", "body", "expected-status-codes", "follow-redirects",
		"max-redirects", "keyword-check", "keyword-check-type",
		"json-path-check", "json-path-expected", "ssl-check-enabled",
		"ssl-expiry-threshold-days",
	}

	for _, flagName := range optionalFlags {
		if cmd.Flags().Lookup(flagName) == nil {
			t.Errorf("Expected --%s flag to exist", flagName)
		}
	}
}

func TestProbeUpdateCmd_NameResolution(t *testing.T) {
	// Since probe name resolution was added, non-UUID inputs are now treated as
	// potential probe names that need API resolution. Without a configured API
	// client, these will fail with an API client initialization error.
	cmd := NewProbeUpdateCmd()
	cmd.SetArgs([]string{"my-probe-name", "--name", "Test"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when API client not configured, got nil")
	}

	expectedMsg := "failed to initialize API client"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestProbeUpdateCmd_NoFlags(t *testing.T) {
	cmd := NewProbeUpdateCmd()
	cmd.SetArgs([]string{"550e8400-e29b-41d4-a716-446655440000"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when no flags specified, got nil")
	}

	expectedMsg := "no update flags specified"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestProbeUpdateCmd_EmptyName(t *testing.T) {
	cmd := NewProbeUpdateCmd()
	cmd.SetArgs([]string{"550e8400-e29b-41d4-a716-446655440000", "--name", ""})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for empty name, got nil")
	}

	expectedMsg := "--name cannot be empty"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestProbeUpdateCmd_EmptyURL(t *testing.T) {
	cmd := NewProbeUpdateCmd()
	cmd.SetArgs([]string{"550e8400-e29b-41d4-a716-446655440000", "--url", ""})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for empty URL, got nil")
	}

	expectedMsg := "--url cannot be empty"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestProbeUpdateCmd_InvalidURL(t *testing.T) {
	cmd := NewProbeUpdateCmd()
	cmd.SetArgs([]string{"550e8400-e29b-41d4-a716-446655440000", "--url", "not-a-url"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}

	expectedMsg := "URL must include scheme"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestProbeUpdateCmd_InvalidMethod(t *testing.T) {
	cmd := NewProbeUpdateCmd()
	cmd.SetArgs([]string{"550e8400-e29b-41d4-a716-446655440000", "--method", "INVALID"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid method, got nil")
	}

	expectedMsg := "invalid --method"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestProbeUpdateCmd_InvalidInterval(t *testing.T) {
	tests := []struct {
		name     string
		interval string
	}{
		{name: "too low", interval: "10"},
		{name: "too high", interval: "4000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewProbeUpdateCmd()
			cmd.SetArgs([]string{"550e8400-e29b-41d4-a716-446655440000", "--interval", tt.interval})

			err := cmd.Execute()
			if err == nil {
				t.Error("Expected error for invalid interval, got nil")
			}

			expectedMsg := "--interval must be between 30 and 3600"
			if err != nil && !strings.Contains(err.Error(), expectedMsg) {
				t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
			}
		})
	}
}

func TestProbeUpdateCmd_InvalidTimeout(t *testing.T) {
	tests := []struct {
		name    string
		timeout string
	}{
		{name: "too low", timeout: "0"},
		{name: "too high", timeout: "100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewProbeUpdateCmd()
			cmd.SetArgs([]string{"550e8400-e29b-41d4-a716-446655440000", "--timeout", tt.timeout})

			err := cmd.Execute()
			if err == nil {
				t.Error("Expected error for invalid timeout, got nil")
			}

			expectedMsg := "--timeout must be between 1 and 60"
			if err != nil && !strings.Contains(err.Error(), expectedMsg) {
				t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
			}
		})
	}
}

func TestProbeUpdateCmd_InvalidMaxRedirects(t *testing.T) {
	tests := []struct {
		name     string
		redirect string
	}{
		{name: "too low", redirect: "-1"},
		{name: "too high", redirect: "25"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewProbeUpdateCmd()
			cmd.SetArgs([]string{"550e8400-e29b-41d4-a716-446655440000", "--max-redirects", tt.redirect})

			err := cmd.Execute()
			if err == nil {
				t.Error("Expected error for invalid max-redirects, got nil")
			}

			expectedMsg := "--max-redirects must be between 0 and 20"
			if err != nil && !strings.Contains(err.Error(), expectedMsg) {
				t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
			}
		})
	}
}

func TestProbeUpdateCmd_InvalidKeywordCheckType(t *testing.T) {
	cmd := NewProbeUpdateCmd()
	cmd.SetArgs([]string{"550e8400-e29b-41d4-a716-446655440000", "--keyword-check-type", "invalid"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid keyword-check-type, got nil")
	}

	expectedMsg := "invalid --keyword-check-type"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestProbeUpdateCmd_InvalidSSLExpiryDays(t *testing.T) {
	tests := []struct {
		name string
		days string
	}{
		{name: "too low", days: "0"},
		{name: "too high", days: "400"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewProbeUpdateCmd()
			cmd.SetArgs([]string{"550e8400-e29b-41d4-a716-446655440000", "--ssl-expiry-threshold-days", tt.days})

			err := cmd.Execute()
			if err == nil {
				t.Error("Expected error for invalid ssl-expiry-threshold-days, got nil")
			}

			expectedMsg := "--ssl-expiry-threshold-days must be between 1 and 365"
			if err != nil && !strings.Contains(err.Error(), expectedMsg) {
				t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
			}
		})
	}
}

func TestProbeUpdateCmd_InvalidExpectedStatusCodes(t *testing.T) {
	cmd := NewProbeUpdateCmd()
	cmd.SetArgs([]string{"550e8400-e29b-41d4-a716-446655440000", "--expected-status-codes", "abc"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid expected-status-codes, got nil")
	}

	expectedMsg := "invalid --expected-status-codes"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}
