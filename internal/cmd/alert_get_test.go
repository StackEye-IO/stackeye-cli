package cmd

import (
	"context"
	"strings"
	"testing"
)

func TestNewAlertGetCmd(t *testing.T) {
	cmd := NewAlertGetCmd()

	if cmd.Use != "get <id>" {
		t.Errorf("expected Use='get <id>', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Short != "Get details of a monitoring alert" {
		t.Errorf("expected Short='Get details of a monitoring alert', got %q", cmd.Short)
	}
}

func TestNewAlertGetCmd_Args(t *testing.T) {
	cmd := NewAlertGetCmd()

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

func TestNewAlertGetCmd_Long(t *testing.T) {
	cmd := NewAlertGetCmd()

	long := cmd.Long

	// Should contain status documentation
	statuses := []string{"active", "acknowledged", "resolved"}
	for _, status := range statuses {
		if !strings.Contains(long, status) {
			t.Errorf("expected Long description to mention status %q", status)
		}
	}

	// Should contain severity documentation
	severities := []string{"critical", "warning", "info"}
	for _, severity := range severities {
		if !strings.Contains(long, severity) {
			t.Errorf("expected Long description to mention severity %q", severity)
		}
	}

	// Should have usage examples
	if !strings.Contains(long, "stackeye alert get") {
		t.Error("expected Long description to contain example commands")
	}

	// Should mention timeline flag
	if !strings.Contains(long, "--timeline") {
		t.Error("expected Long description to mention --timeline flag")
	}
}

func TestNewAlertGetCmd_Flags(t *testing.T) {
	cmd := NewAlertGetCmd()

	// Verify expected flags exist
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"timeline", "false"},
	}

	for _, f := range flags {
		flag := cmd.Flags().Lookup(f.name)
		if flag == nil {
			t.Errorf("expected flag %q to exist", f.name)
			continue
		}
		if flag.DefValue != f.defaultValue {
			t.Errorf("flag %q: expected default %q, got %q", f.name, f.defaultValue, flag.DefValue)
		}
	}
}

func TestRunAlertGet_Validation(t *testing.T) {
	// Test that validation errors are returned for invalid inputs.
	// Since runAlertGet requires an API client, validation happens first
	// and will fail before making API calls for invalid inputs.
	tests := []struct {
		name         string
		alertID      string
		wantErrorMsg string
	}{
		{
			name:         "empty ID",
			alertID:      "",
			wantErrorMsg: `invalid alert ID ""`,
		},
		{
			name:         "invalid UUID format",
			alertID:      "not-a-uuid",
			wantErrorMsg: `invalid alert ID "not-a-uuid": must be a valid UUID`,
		},
		{
			name:         "partial UUID",
			alertID:      "550e8400-e29b-41d4",
			wantErrorMsg: `invalid alert ID "550e8400-e29b-41d4": must be a valid UUID`,
		},
		{
			name:         "UUID with extra characters",
			alertID:      "550e8400-e29b-41d4-a716-446655440000-extra",
			wantErrorMsg: `invalid alert ID`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := &alertGetFlags{
				timeline: false,
			}

			// Call runAlertGet with a background context.
			// It should fail on validation before needing API client.
			err := runAlertGet(context.Background(), tt.alertID, flags)

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

func TestRunAlertGet_ValidUUID(t *testing.T) {
	// Test that a valid UUID passes validation (will fail later on API client)
	flags := &alertGetFlags{
		timeline: false,
	}

	validUUID := "550e8400-e29b-41d4-a716-446655440000"
	err := runAlertGet(context.Background(), validUUID, flags)

	// Should fail on API client initialization, not validation
	if err == nil {
		t.Error("expected error (no API client configured), got nil")
		return
	}

	// Error should NOT be a validation error
	if strings.Contains(err.Error(), "invalid alert ID") {
		t.Errorf("got unexpected validation error for valid UUID: %s", err.Error())
	}
}

func TestRunAlertGet_ValidUUIDWithTimeline(t *testing.T) {
	// Test that a valid UUID with timeline flag passes validation
	flags := &alertGetFlags{
		timeline: true,
	}

	validUUID := "550e8400-e29b-41d4-a716-446655440000"
	err := runAlertGet(context.Background(), validUUID, flags)

	// Should fail on API client initialization, not validation
	if err == nil {
		t.Error("expected error (no API client configured), got nil")
		return
	}

	// Error should NOT be a validation error
	if strings.Contains(err.Error(), "invalid alert ID") {
		t.Errorf("got unexpected validation error for valid UUID: %s", err.Error())
	}
}

func TestAlertGetResponse_Structure(t *testing.T) {
	// Test that AlertGetResponse has expected fields and can be created
	response := &AlertGetResponse{
		Alert:    nil,
		Timeline: nil,
	}

	// Verify the Alert field can hold nil
	if response.Alert != nil {
		t.Error("expected Alert field to be nil when set to nil")
	}

	// Verify Timeline field can hold nil
	if response.Timeline != nil {
		t.Error("expected Timeline field to be nil when set to nil")
	}
}
