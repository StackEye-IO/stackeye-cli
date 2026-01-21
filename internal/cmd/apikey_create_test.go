// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"strings"
	"testing"
	"time"
)

// TestNewAPIKeyCreateCmd verifies that the api-key create command is properly constructed.
func TestNewAPIKeyCreateCmd(t *testing.T) {
	cmd := NewAPIKeyCreateCmd()

	if cmd.Use != "create" {
		t.Errorf("expected Use to be 'create', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	// Verify required flags exist
	nameFlag := cmd.Flags().Lookup("name")
	if nameFlag == nil {
		t.Error("expected --name flag to be defined")
	}

	// Verify optional flags exist
	permsFlag := cmd.Flags().Lookup("permissions")
	if permsFlag == nil {
		t.Error("expected --permissions flag to be defined")
	}

	expiresFlag := cmd.Flags().Lookup("expires-in")
	if expiresFlag == nil {
		t.Error("expected --expires-in flag to be defined")
	}
}

// TestAPIKeyCreateCmd_Parent verifies the parent api-key command registers the create subcommand.
func TestAPIKeyCreateCmd_Parent(t *testing.T) {
	parentCmd := NewAPIKeyCmd()

	// Check that create subcommand is registered
	var foundCreate bool
	for _, sub := range parentCmd.Commands() {
		if sub.Use == "create" {
			foundCreate = true
			break
		}
	}

	if !foundCreate {
		t.Error("expected 'create' subcommand to be registered under api-key command")
	}
}

// TestRunAPIKeyCreate_RequiresClient tests that running create without a configured client fails.
func TestRunAPIKeyCreate_RequiresClient(t *testing.T) {
	flags := &apiKeyCreateFlags{
		name: "Test Key",
	}

	// Call runAPIKeyCreate with a background context.
	// It should fail when trying to get the API client since none is configured.
	err := runAPIKeyCreate(context.Background(), flags)

	if err == nil {
		t.Error("expected error (no API client configured), got nil")
		return
	}

	// Error should be related to API client initialization
	errStr := err.Error()
	if strings.Contains(errStr, "API client") || strings.Contains(errStr, "client") || strings.Contains(errStr, "config") {
		t.Logf("got expected error: %v", err)
	} else {
		t.Errorf("error should mention API client or config, got: %v", err)
	}
}

// TestAPIKeyCreateCmd_LongDescription verifies the long description contains expected content.
func TestAPIKeyCreateCmd_LongDescription(t *testing.T) {
	cmd := NewAPIKeyCreateCmd()

	expectedPhrases := []string{
		"API key",
		"ONLY ONCE",
		"--name",
		"Security",
		"stackeye api-key create",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(cmd.Long, phrase) {
			t.Errorf("expected Long description to contain %q", phrase)
		}
	}
}

// TestParseDuration tests the duration parsing function.
func TestParseDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		wantErr  bool
		errMsg   string
	}{
		// Valid inputs
		{"1d", 24 * time.Hour, false, ""},
		{"7d", 7 * 24 * time.Hour, false, ""},
		{"30d", 30 * 24 * time.Hour, false, ""},
		{"1w", 7 * 24 * time.Hour, false, ""},
		{"2w", 14 * 24 * time.Hour, false, ""},
		{"1m", 30 * 24 * time.Hour, false, ""},
		{"3m", 90 * 24 * time.Hour, false, ""},
		{"1y", 365 * 24 * time.Hour, false, ""},

		// Invalid inputs
		{"", 0, true, "too short"},
		{"d", 0, true, ""},
		{"0d", 0, true, "positive"},
		{"-1d", 0, true, ""},
		{"1x", 0, true, "unknown unit"},
		{"abc", 0, true, ""},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result, err := parseDuration(tc.input)

			if tc.wantErr {
				if err == nil {
					t.Errorf("parseDuration(%q) expected error, got nil", tc.input)
				} else if tc.errMsg != "" && !strings.Contains(err.Error(), tc.errMsg) {
					t.Errorf("parseDuration(%q) error %q should contain %q", tc.input, err.Error(), tc.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("parseDuration(%q) unexpected error: %v", tc.input, err)
				}
				if result != tc.expected {
					t.Errorf("parseDuration(%q) = %v, expected %v", tc.input, result, tc.expected)
				}
			}
		})
	}
}

// TestAPIKeyCreateCmd_RequiresNameFlag tests that the name flag is required.
func TestAPIKeyCreateCmd_RequiresNameFlag(t *testing.T) {
	cmd := NewAPIKeyCreateCmd()

	// Attempt to execute without setting --name
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	if err == nil {
		t.Error("expected error for missing required --name flag")
	}

	if !strings.Contains(err.Error(), "name") {
		t.Errorf("error should mention 'name' flag: %v", err)
	}
}
