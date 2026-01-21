package cmd

import (
	"context"
	"strings"
	"testing"
)

func TestNewChannelGetCmd(t *testing.T) {
	cmd := NewChannelGetCmd()

	if cmd.Use != "get <id>" {
		t.Errorf("expected Use='get <id>', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Short != "Get details of a notification channel" {
		t.Errorf("expected Short='Get details of a notification channel', got %q", cmd.Short)
	}
}

func TestNewChannelGetCmd_Args(t *testing.T) {
	cmd := NewChannelGetCmd()

	// Should require exactly 1 argument
	if cmd.Args == nil {
		t.Error("expected Args validator to be set")
	}

	// Test with wrong number of args via Execute
	cmd.SetArgs([]string{}) // No args
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when no arguments provided")
	}
}

func TestNewChannelGetCmd_Long(t *testing.T) {
	cmd := NewChannelGetCmd()

	long := cmd.Long

	// Should contain channel type documentation
	channelTypes := []string{"email", "slack", "webhook", "pagerduty", "discord", "teams", "sms"}
	for _, channelType := range channelTypes {
		if !strings.Contains(long, channelType) {
			t.Errorf("expected Long description to mention channel type %q", channelType)
		}
	}

	// Should have usage examples
	if !strings.Contains(long, "stackeye channel get") {
		t.Error("expected Long description to contain example commands")
	}

	// Should mention output formats
	if !strings.Contains(long, "-o json") {
		t.Error("expected Long description to mention JSON output format")
	}

	if !strings.Contains(long, "-o yaml") {
		t.Error("expected Long description to mention YAML output format")
	}
}

func TestRunChannelGet_Validation(t *testing.T) {
	// Test that validation errors are returned for invalid inputs.
	// Since runChannelGet requires an API client, validation happens first
	// and will fail before making API calls for invalid inputs.
	tests := []struct {
		name         string
		channelID    string
		wantErrorMsg string
	}{
		{
			name:         "empty ID",
			channelID:    "",
			wantErrorMsg: `invalid channel ID ""`,
		},
		{
			name:         "invalid UUID format",
			channelID:    "not-a-uuid",
			wantErrorMsg: `invalid channel ID "not-a-uuid": must be a valid UUID`,
		},
		{
			name:         "partial UUID",
			channelID:    "550e8400-e29b-41d4",
			wantErrorMsg: `invalid channel ID "550e8400-e29b-41d4": must be a valid UUID`,
		},
		{
			name:         "UUID with extra characters",
			channelID:    "550e8400-e29b-41d4-a716-446655440000-extra",
			wantErrorMsg: `invalid channel ID`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call runChannelGet with a background context.
			// It should fail on validation before needing API client.
			err := runChannelGet(context.Background(), tt.channelID)

			if err == nil {
				t.Errorf("expected error containing %q, got nil", tt.wantErrorMsg)
				return
			}

			if !strings.Contains(err.Error(), tt.wantErrorMsg) {
				t.Errorf("expected error containing %q, got %q", tt.wantErrorMsg, err.Error())
			}
		})
	}
}

func TestRunChannelGet_ValidUUID(t *testing.T) {
	// Test that a valid UUID passes validation (will fail later on API client)
	validUUID := "550e8400-e29b-41d4-a716-446655440000"
	err := runChannelGet(context.Background(), validUUID)

	// Should fail on API client initialization, not validation
	if err == nil {
		t.Error("expected error (no API client configured), got nil")
		return
	}

	// Error should NOT be a validation error
	if strings.Contains(err.Error(), "invalid channel ID") {
		t.Errorf("got unexpected validation error for valid UUID: %s", err.Error())
	}
}
