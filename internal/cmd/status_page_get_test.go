package cmd

import (
	"context"
	"strings"
	"testing"
)

func TestNewStatusPageGetCmd(t *testing.T) {
	cmd := NewStatusPageGetCmd()

	if cmd.Use != "get <id>" {
		t.Errorf("expected Use='get <id>', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Short != "Get details of a status page" {
		t.Errorf("expected Short='Get details of a status page', got %q", cmd.Short)
	}
}

func TestNewStatusPageGetCmd_Args(t *testing.T) {
	cmd := NewStatusPageGetCmd()

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

func TestNewStatusPageGetCmd_Long(t *testing.T) {
	cmd := NewStatusPageGetCmd()

	long := cmd.Long

	// Should contain field documentation
	fields := []string{"NAME", "SLUG", "THEME", "PUBLIC", "ENABLED", "PROBES"}
	for _, field := range fields {
		if !strings.Contains(long, field) {
			t.Errorf("expected Long description to mention field %q", field)
		}
	}

	// Should contain wide mode fields
	wideFields := []string{"DOMAIN", "UPTIME%", "ID", "CREATED"}
	for _, field := range wideFields {
		if !strings.Contains(long, field) {
			t.Errorf("expected Long description to mention wide field %q", field)
		}
	}

	// Should have usage examples
	if !strings.Contains(long, "stackeye status-page get") {
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

func TestRunStatusPageGet_Validation(t *testing.T) {
	// Test that validation errors are returned for invalid inputs.
	// Since runStatusPageGet requires an API client, validation happens first
	// and will fail before making API calls for invalid inputs.
	tests := []struct {
		name         string
		statusPageID string
		wantErrorMsg string
	}{
		{
			name:         "empty ID",
			statusPageID: "",
			wantErrorMsg: `invalid status page ID ""`,
		},
		{
			name:         "non-numeric ID",
			statusPageID: "not-a-number",
			wantErrorMsg: `invalid status page ID "not-a-number": must be a positive integer`,
		},
		{
			name:         "negative number",
			statusPageID: "-1",
			wantErrorMsg: `invalid status page ID "-1": must be a positive integer`,
		},
		{
			name:         "zero ID",
			statusPageID: "0",
			wantErrorMsg: "invalid status page ID: must be greater than 0",
		},
		{
			name:         "float number",
			statusPageID: "1.5",
			wantErrorMsg: `invalid status page ID "1.5": must be a positive integer`,
		},
		{
			name:         "UUID instead of integer",
			statusPageID: "550e8400-e29b-41d4-a716-446655440000",
			wantErrorMsg: `invalid status page ID`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call runStatusPageGet with a background context.
			// It should fail on validation before needing API client.
			err := runStatusPageGet(context.Background(), tt.statusPageID)

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

func TestRunStatusPageGet_ValidID(t *testing.T) {
	// Test that a valid ID passes validation (will fail later on API client)
	validID := "123"
	err := runStatusPageGet(context.Background(), validID)

	// Should fail on API client initialization, not validation
	if err == nil {
		t.Error("expected error (no API client configured), got nil")
		return
	}

	// Error should NOT be a validation error
	if strings.Contains(err.Error(), "invalid status page ID") {
		t.Errorf("got unexpected validation error for valid ID: %s", err.Error())
	}
}

func TestRunStatusPageGet_LargeValidID(t *testing.T) {
	// Test that large valid IDs work (within uint64 range)
	largeID := "999999999"
	err := runStatusPageGet(context.Background(), largeID)

	// Should fail on API client initialization, not validation
	if err == nil {
		t.Error("expected error (no API client configured), got nil")
		return
	}

	// Error should NOT be a validation error
	if strings.Contains(err.Error(), "invalid status page ID") {
		t.Errorf("got unexpected validation error for large valid ID: %s", err.Error())
	}
}
