// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"
)

// testRemoveProbeUUID is a test fixture UUID for remove-probe command tests.
const testRemoveProbeUUID = "550e8400-e29b-41d4-a716-446655440000"

func TestNewStatusPageRemoveProbeCmd(t *testing.T) {
	cmd := NewStatusPageRemoveProbeCmd()

	if cmd.Use != "remove-probe <status-page-id>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "remove-probe <status-page-id>")
	}

	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if cmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestStatusPageRemoveProbeCmd_ProbeIDFlag(t *testing.T) {
	cmd := NewStatusPageRemoveProbeCmd()

	flag := cmd.Flags().Lookup("probe-id")
	if flag == nil {
		t.Error("Expected --probe-id flag to exist")
		return
	}

	if flag.DefValue != "" {
		t.Errorf("--probe-id default = %q, want empty string", flag.DefValue)
	}
}

func TestStatusPageRemoveProbeCmd_NoArgs(t *testing.T) {
	cmd := NewStatusPageRemoveProbeCmd()
	cmd.SetArgs([]string{"--probe-id", testRemoveProbeUUID})

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

func TestStatusPageRemoveProbeCmd_TooManyArgs(t *testing.T) {
	cmd := NewStatusPageRemoveProbeCmd()
	cmd.SetArgs([]string{"123", "456", "--probe-id", testRemoveProbeUUID})

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

func TestStatusPageRemoveProbeCmd_MissingProbeID(t *testing.T) {
	cmd := NewStatusPageRemoveProbeCmd()
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

func TestStatusPageRemoveProbeCmd_InvalidID(t *testing.T) {
	cmd := NewStatusPageRemoveProbeCmd()
	cmd.SetArgs([]string{"not-a-number", "--probe-id", testRemoveProbeUUID})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid ID, got nil")
	}

	expectedMsg := "invalid status page ID"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestStatusPageRemoveProbeCmd_InvalidProbeUUID(t *testing.T) {
	cmd := NewStatusPageRemoveProbeCmd()
	cmd.SetArgs([]string{"123", "--probe-id", "not-a-valid-uuid"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid probe UUID, got nil")
	}

	expectedMsg := "invalid probe ID"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestStatusPageRemoveProbeCmd_NegativeID(t *testing.T) {
	cmd := NewStatusPageRemoveProbeCmd()
	// Use "--" to signal end of flags so "-1" is treated as positional argument
	cmd.SetArgs([]string{"--probe-id", testRemoveProbeUUID, "--", "-1"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for negative ID, got nil")
	}

	expectedMsg := "invalid status page ID"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestStatusPageRemoveProbeCmd_ZeroID(t *testing.T) {
	cmd := NewStatusPageRemoveProbeCmd()
	cmd.SetArgs([]string{"0", "--probe-id", testRemoveProbeUUID})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for zero ID, got nil")
	}

	expectedMsg := "invalid status page ID"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestStatusPageRemoveProbeCmd_EmptyID(t *testing.T) {
	cmd := NewStatusPageRemoveProbeCmd()
	cmd.SetArgs([]string{"", "--probe-id", testRemoveProbeUUID})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for empty ID, got nil")
	}

	expectedMsg := "invalid status page ID"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestStatusPageRemoveProbeCmd_ValidID(t *testing.T) {
	cmd := NewStatusPageRemoveProbeCmd()
	cmd.SetArgs([]string{"123", "--probe-id", testRemoveProbeUUID})

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

func TestStatusPageRemoveProbeCmd_Aliases(t *testing.T) {
	cmd := NewStatusPageRemoveProbeCmd()

	// remove-probe command doesn't define aliases
	if len(cmd.Aliases) != 0 {
		t.Errorf("Expected no aliases for remove-probe command, got %v", cmd.Aliases)
	}
}

func TestStatusPageRemoveProbeCmd_EmptyProbeID(t *testing.T) {
	cmd := NewStatusPageRemoveProbeCmd()
	cmd.SetArgs([]string{"123", "--probe-id", ""})

	err := cmd.Execute()
	// Empty string is valid for flag parsing but will fail validation
	// Note: Cobra may accept empty string as flag value
	if err == nil {
		t.Log("Note: Empty probe-id may pass flag parsing but should fail in execution")
	}
}
