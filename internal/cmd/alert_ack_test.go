package cmd

import (
	"context"
	"slices"
	"strings"
	"testing"
)

func TestNewAlertAckCmd(t *testing.T) {
	cmd := NewAlertAckCmd()

	if cmd.Use != "ack <id> [id...]" {
		t.Errorf("expected Use='ack <id> [id...]', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Short != "Acknowledge a monitoring alert" {
		t.Errorf("expected Short='Acknowledge a monitoring alert', got %q", cmd.Short)
	}
}

func TestNewAlertAckCmd_Args(t *testing.T) {
	cmd := NewAlertAckCmd()

	// Should require at least 1 argument
	if cmd.Args == nil {
		t.Error("expected Args validator to be set")
	}

	// Test with no args via Execute
	cmd.SetArgs([]string{}) // No args
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when no arguments provided")
	}
}

func TestNewAlertAckCmd_Long(t *testing.T) {
	cmd := NewAlertAckCmd()

	long := cmd.Long

	// Should contain state documentation
	if !strings.Contains(long, "acknowledged") {
		t.Error("expected Long description to mention acknowledged state")
	}

	// Should have usage examples
	if !strings.Contains(long, "stackeye alert ack") {
		t.Error("expected Long description to contain example commands")
	}

	// Should mention message flag
	if !strings.Contains(long, "-m") {
		t.Error("expected Long description to mention -m flag")
	}

	// Should document batch operation
	if !strings.Contains(long, "multiple alerts") {
		t.Error("expected Long description to mention multiple alerts capability")
	}
}

func TestNewAlertAckCmd_Flags(t *testing.T) {
	cmd := NewAlertAckCmd()

	// Verify expected flags exist
	flags := []struct {
		name         string
		shorthand    string
		defaultValue string
	}{
		{"message", "m", ""},
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
		if flag.Shorthand != f.shorthand {
			t.Errorf("flag %q: expected shorthand %q, got %q", f.name, f.shorthand, flag.Shorthand)
		}
	}
}

func TestNewAlertAckCmd_Aliases(t *testing.T) {
	cmd := NewAlertAckCmd()

	if len(cmd.Aliases) == 0 {
		t.Error("expected command to have aliases")
		return
	}

	if !slices.Contains(cmd.Aliases, "acknowledge") {
		t.Error("expected 'acknowledge' to be an alias")
	}
}

func TestRunAlertAck_Validation(t *testing.T) {
	// Test that validation errors are returned for invalid inputs.
	// Since runAlertAck requires an API client, validation happens first
	// and will fail before making API calls for invalid inputs.
	tests := []struct {
		name         string
		alertIDs     []string
		wantErrorMsg string
	}{
		{
			name:         "empty ID",
			alertIDs:     []string{""},
			wantErrorMsg: `invalid alert ID ""`,
		},
		{
			name:         "invalid UUID format",
			alertIDs:     []string{"not-a-uuid"},
			wantErrorMsg: `invalid alert ID "not-a-uuid": must be a valid UUID`,
		},
		{
			name:         "partial UUID",
			alertIDs:     []string{"550e8400-e29b-41d4"},
			wantErrorMsg: `invalid alert ID "550e8400-e29b-41d4": must be a valid UUID`,
		},
		{
			name:         "UUID with extra characters",
			alertIDs:     []string{"550e8400-e29b-41d4-a716-446655440000-extra"},
			wantErrorMsg: `invalid alert ID`,
		},
		{
			name:         "second ID invalid in batch",
			alertIDs:     []string{"550e8400-e29b-41d4-a716-446655440000", "invalid"},
			wantErrorMsg: `invalid alert ID "invalid": must be a valid UUID`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := &alertAckFlags{
				message: "",
			}

			// Call runAlertAck with a background context.
			// It should fail on validation before needing API client.
			err := runAlertAck(context.Background(), tt.alertIDs, flags)

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

func TestRunAlertAck_ValidUUID(t *testing.T) {
	// Test that a valid UUID passes validation (will fail later on API client)
	flags := &alertAckFlags{
		message: "",
	}

	validUUID := "550e8400-e29b-41d4-a716-446655440000"
	err := runAlertAck(context.Background(), []string{validUUID}, flags)

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

func TestRunAlertAck_ValidUUIDWithMessage(t *testing.T) {
	// Test that a valid UUID with message flag passes validation
	flags := &alertAckFlags{
		message: "Investigating the issue",
	}

	validUUID := "550e8400-e29b-41d4-a716-446655440000"
	err := runAlertAck(context.Background(), []string{validUUID}, flags)

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

func TestRunAlertAck_MultipleValidUUIDs(t *testing.T) {
	// Test that multiple valid UUIDs pass validation
	flags := &alertAckFlags{
		message: "",
	}

	validUUIDs := []string{
		"550e8400-e29b-41d4-a716-446655440000",
		"660e8400-e29b-41d4-a716-446655440001",
		"770e8400-e29b-41d4-a716-446655440002",
	}
	err := runAlertAck(context.Background(), validUUIDs, flags)

	// Should fail on API client initialization, not validation
	if err == nil {
		t.Error("expected error (no API client configured), got nil")
		return
	}

	// Error should NOT be a validation error
	if strings.Contains(err.Error(), "invalid alert ID") {
		t.Errorf("got unexpected validation error for valid UUIDs: %s", err.Error())
	}
}

func TestAlertAckResponse_Structure(t *testing.T) {
	// Test that AlertAckResponse has expected fields and can be created
	response := &AlertAckResponse{
		Alert:   nil,
		Message: "test message",
	}

	// Verify the Alert field can hold nil
	if response.Alert != nil {
		t.Error("expected Alert field to be nil when set to nil")
	}

	// Verify Message field
	if response.Message != "test message" {
		t.Errorf("expected Message='test message', got %q", response.Message)
	}
}

func TestAlertAckBatchResponse_Structure(t *testing.T) {
	// Test that AlertAckBatchResponse has expected fields and can be created
	response := &AlertAckBatchResponse{
		Acknowledged: nil,
		Failed:       nil,
		Total:        5,
		SuccessCount: 3,
		FailedCount:  2,
	}

	// Verify counts
	if response.Total != 5 {
		t.Errorf("expected Total=5, got %d", response.Total)
	}
	if response.SuccessCount != 3 {
		t.Errorf("expected SuccessCount=3, got %d", response.SuccessCount)
	}
	if response.FailedCount != 2 {
		t.Errorf("expected FailedCount=2, got %d", response.FailedCount)
	}
}

func TestAlertAckError_Structure(t *testing.T) {
	// Test that AlertAckError has expected fields
	errEntry := AlertAckError{
		ID:    "550e8400-e29b-41d4-a716-446655440000",
		Error: "not found",
	}

	if errEntry.ID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("expected ID to be set correctly, got %q", errEntry.ID)
	}
	if errEntry.Error != "not found" {
		t.Errorf("expected Error='not found', got %q", errEntry.Error)
	}
}
