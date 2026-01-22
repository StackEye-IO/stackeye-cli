// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"
)

// testProbeUUID is a test fixture UUID for add-probe command tests.
const testProbeUUID = "550e8400-e29b-41d4-a716-446655440000"

func TestNewStatusPageAddProbeCmd(t *testing.T) {
	cmd := NewStatusPageAddProbeCmd()

	if cmd.Use != "add-probe <status-page-id>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "add-probe <status-page-id>")
	}

	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if cmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestStatusPageAddProbeCmd_ProbeIDFlag(t *testing.T) {
	cmd := NewStatusPageAddProbeCmd()

	flag := cmd.Flags().Lookup("probe-id")
	if flag == nil {
		t.Error("Expected --probe-id flag to exist")
		return
	}

	if flag.DefValue != "" {
		t.Errorf("--probe-id default = %q, want empty string", flag.DefValue)
	}
}

func TestStatusPageAddProbeCmd_DisplayNameFlag(t *testing.T) {
	cmd := NewStatusPageAddProbeCmd()

	flag := cmd.Flags().Lookup("display-name")
	if flag == nil {
		t.Error("Expected --display-name flag to exist")
		return
	}

	if flag.DefValue != "" {
		t.Errorf("--display-name default = %q, want empty string", flag.DefValue)
	}
}

func TestStatusPageAddProbeCmd_ShowResponseTimeFlag(t *testing.T) {
	cmd := NewStatusPageAddProbeCmd()

	flag := cmd.Flags().Lookup("show-response-time")
	if flag == nil {
		t.Error("Expected --show-response-time flag to exist")
		return
	}

	if flag.DefValue != "false" {
		t.Errorf("--show-response-time default = %q, want %q", flag.DefValue, "false")
	}
}

func TestStatusPageAddProbeCmd_NoArgs(t *testing.T) {
	cmd := NewStatusPageAddProbeCmd()
	cmd.SetArgs([]string{"--probe-id", testProbeUUID})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when no positional arguments provided, got nil")
	}

	// Cobra's ExactArgs(1) produces a specific error message
	expectedMsg := "accepts 1 arg"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestStatusPageAddProbeCmd_TooManyArgs(t *testing.T) {
	cmd := NewStatusPageAddProbeCmd()
	cmd.SetArgs([]string{"123", "456", "--probe-id", testProbeUUID})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when too many arguments provided, got nil")
	}

	// Cobra's ExactArgs(1) produces a specific error message
	expectedMsg := "accepts 1 arg"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestStatusPageAddProbeCmd_MissingProbeID(t *testing.T) {
	cmd := NewStatusPageAddProbeCmd()
	cmd.SetArgs([]string{"123"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when --probe-id not provided, got nil")
	}

	expectedMsg := "required flag"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestStatusPageAddProbeCmd_InvalidID(t *testing.T) {
	cmd := NewStatusPageAddProbeCmd()
	cmd.SetArgs([]string{"not-a-number", "--probe-id", testProbeUUID})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid ID, got nil")
	}

	expectedMsg := "invalid status page ID"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestStatusPageAddProbeCmd_NegativeID(t *testing.T) {
	cmd := NewStatusPageAddProbeCmd()
	// Use "--" to signal end of flags so "-1" is treated as positional argument
	cmd.SetArgs([]string{"--probe-id", testProbeUUID, "--", "-1"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for negative ID, got nil")
	}

	expectedMsg := "invalid status page ID"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestStatusPageAddProbeCmd_ZeroID(t *testing.T) {
	cmd := NewStatusPageAddProbeCmd()
	cmd.SetArgs([]string{"0", "--probe-id", testProbeUUID})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for zero ID, got nil")
	}

	expectedMsg := "invalid status page ID"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestStatusPageAddProbeCmd_EmptyID(t *testing.T) {
	cmd := NewStatusPageAddProbeCmd()
	cmd.SetArgs([]string{"", "--probe-id", testProbeUUID})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for empty ID, got nil")
	}

	expectedMsg := "invalid status page ID"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestStatusPageAddProbeCmd_ValidID(t *testing.T) {
	cmd := NewStatusPageAddProbeCmd()
	cmd.SetArgs([]string{"123", "--probe-id", testProbeUUID})

	err := cmd.Execute()
	// Will fail at API call stage (no API client configured in tests),
	// but that's expected. We're testing argument parsing here.
	if err == nil {
		t.Log("Note: Command succeeded (expected in some test environments)")
	}

	// Should NOT fail with invalid status page ID error
	if err != nil && strings.Contains(err.Error(), "invalid status page ID") {
		t.Errorf("Should not get invalid status page ID error for valid ID: %v", err)
	}
}

func TestStatusPageAddProbeCmd_Aliases(t *testing.T) {
	cmd := NewStatusPageAddProbeCmd()

	// add-probe command doesn't define aliases
	if len(cmd.Aliases) != 0 {
		t.Errorf("Expected no aliases for add-probe command, got %v", cmd.Aliases)
	}
}

func TestStatusPageAddProbeCmd_WithDisplayName(t *testing.T) {
	cmd := NewStatusPageAddProbeCmd()
	cmd.SetArgs([]string{"123", "--probe-id", testProbeUUID, "--display-name", "API Server"})

	err := cmd.Execute()
	// Will fail at API call stage, but parsing should succeed
	if err != nil && strings.Contains(err.Error(), "invalid") {
		t.Errorf("Should not get parsing error with valid arguments: %v", err)
	}
}

func TestStatusPageAddProbeCmd_WithShowResponseTime(t *testing.T) {
	cmd := NewStatusPageAddProbeCmd()
	cmd.SetArgs([]string{"123", "--probe-id", testProbeUUID, "--show-response-time"})

	err := cmd.Execute()
	// Will fail at API call stage, but parsing should succeed
	if err != nil && strings.Contains(err.Error(), "invalid") {
		t.Errorf("Should not get parsing error with valid arguments: %v", err)
	}
}

func TestStatusPageAddProbeCmd_AllFlags(t *testing.T) {
	cmd := NewStatusPageAddProbeCmd()
	cmd.SetArgs([]string{
		"123",
		"--probe-id", testProbeUUID,
		"--display-name", "API Server",
		"--show-response-time",
	})

	err := cmd.Execute()
	// Will fail at API call stage, but parsing should succeed
	if err != nil && strings.Contains(err.Error(), "invalid") {
		t.Errorf("Should not get parsing error with all valid flags: %v", err)
	}
}
