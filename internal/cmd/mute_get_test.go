package cmd

import (
	"context"
	"strings"
	"testing"
)

func TestNewMuteGetCmd(t *testing.T) {
	cmd := NewMuteGetCmd()

	if cmd.Use != "get <id>" {
		t.Errorf("expected Use='get <id>', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Short != "Get details of an alert mute period" {
		t.Errorf("expected Short='Get details of an alert mute period', got %q", cmd.Short)
	}
}

func TestNewMuteGetCmd_Args(t *testing.T) {
	cmd := NewMuteGetCmd()

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

func TestNewMuteGetCmd_Long(t *testing.T) {
	cmd := NewMuteGetCmd()

	long := cmd.Long

	// Should contain scope documentation
	scopes := []string{"organization", "probe", "channel", "alert_type"}
	for _, scope := range scopes {
		if !strings.Contains(long, scope) {
			t.Errorf("expected Long description to mention scope %q", scope)
		}
	}

	// Should contain status documentation
	statuses := []string{"ACTIVE", "EXPIRED"}
	for _, status := range statuses {
		if !strings.Contains(long, status) {
			t.Errorf("expected Long description to mention status %q", status)
		}
	}

	// Should have usage examples
	if !strings.Contains(long, "stackeye mute get") {
		t.Error("expected Long description to contain example commands")
	}

	// Should mention output formats
	formats := []string{"-o json", "-o yaml", "-o wide"}
	for _, format := range formats {
		if !strings.Contains(long, format) {
			t.Errorf("expected Long description to mention output format %q", format)
		}
	}
}

func TestRunMuteGet_Validation(t *testing.T) {
	// Test that validation errors are returned for invalid inputs.
	// Since runMuteGet requires an API client, validation happens first
	// and will fail before making API calls for invalid inputs.
	tests := []struct {
		name         string
		muteID       string
		wantErrorMsg string
	}{
		{
			name:         "empty ID",
			muteID:       "",
			wantErrorMsg: `invalid mute ID ""`,
		},
		{
			name:         "invalid UUID format",
			muteID:       "not-a-uuid",
			wantErrorMsg: `invalid mute ID "not-a-uuid": must be a valid UUID`,
		},
		{
			name:         "partial UUID",
			muteID:       "550e8400-e29b-41d4",
			wantErrorMsg: `invalid mute ID "550e8400-e29b-41d4": must be a valid UUID`,
		},
		{
			name:         "UUID with extra characters",
			muteID:       "550e8400-e29b-41d4-a716-446655440000-extra",
			wantErrorMsg: `invalid mute ID`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call runMuteGet with a background context.
			// It should fail on validation before needing API client.
			err := runMuteGet(context.Background(), tt.muteID)

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

func TestRunMuteGet_ValidUUID(t *testing.T) {
	// Test that a valid UUID passes validation (will fail later on API client)
	validUUID := "550e8400-e29b-41d4-a716-446655440000"
	err := runMuteGet(context.Background(), validUUID)

	// Should fail on API client initialization, not validation
	if err == nil {
		t.Error("expected error (no API client configured), got nil")
		return
	}

	// Error should NOT be a validation error
	if strings.Contains(err.Error(), "invalid mute ID") {
		t.Errorf("got unexpected validation error for valid UUID: %s", err.Error())
	}
}
