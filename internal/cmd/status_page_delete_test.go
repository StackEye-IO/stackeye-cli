// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"
)

func TestNewStatusPageDeleteCmd(t *testing.T) {
	cmd := NewStatusPageDeleteCmd()

	if cmd.Use != "delete <id>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "delete <id>")
	}

	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if cmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestStatusPageDeleteCmd_YesFlag(t *testing.T) {
	cmd := NewStatusPageDeleteCmd()

	flag := cmd.Flags().Lookup("yes")
	if flag == nil {
		t.Error("Expected --yes flag to exist")
		return
	}

	if flag.Shorthand != "y" {
		t.Errorf("--yes shorthand = %q, want %q", flag.Shorthand, "y")
	}

	if flag.DefValue != "false" {
		t.Errorf("--yes default = %q, want %q", flag.DefValue, "false")
	}
}

func TestStatusPageDeleteCmd_NoArgs(t *testing.T) {
	cmd := NewStatusPageDeleteCmd()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when no arguments provided, got nil")
	}

	// Cobra's ExactArgs(1) produces a specific error message
	expectedMsg := "accepts 1 arg"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestStatusPageDeleteCmd_TooManyArgs(t *testing.T) {
	cmd := NewStatusPageDeleteCmd()
	cmd.SetArgs([]string{"123", "456"})

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

func TestStatusPageDeleteCmd_InvalidID(t *testing.T) {
	cmd := NewStatusPageDeleteCmd()
	// Use --yes to skip confirmation prompt
	cmd.SetArgs([]string{"not-a-number", "--yes"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid ID, got nil")
	}

	expectedMsg := "invalid status page ID"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestStatusPageDeleteCmd_NegativeID(t *testing.T) {
	cmd := NewStatusPageDeleteCmd()
	// Use "--yes" before "--" to set the flag, then use "--" to signal end of flags
	// so that "-1" is treated as a positional argument, not a flag
	cmd.SetArgs([]string{"--yes", "--", "-1"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for negative ID, got nil")
	}

	expectedMsg := "invalid status page ID"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestStatusPageDeleteCmd_ValidID(t *testing.T) {
	cmd := NewStatusPageDeleteCmd()

	// ID is syntactically valid - command should accept it
	// (actual deletion will fail without API, but parsing should succeed)
	cmd.SetArgs([]string{"123", "--yes"})

	err := cmd.Execute()
	// Will fail at API call stage (no API client configured in tests),
	// but that's expected. We're testing argument parsing here.
	if err == nil {
		t.Log("Note: Command succeeded (expected in some test environments)")
	}

	// Should NOT fail with invalid ID error
	if err != nil && strings.Contains(err.Error(), "invalid status page ID") {
		t.Errorf("Should not get invalid status page ID error for valid ID: %v", err)
	}
}

func TestStatusPageDeleteCmd_Aliases(t *testing.T) {
	cmd := NewStatusPageDeleteCmd()

	// delete command doesn't define aliases (direct subcommand of status-page)
	if len(cmd.Aliases) != 0 {
		t.Errorf("Expected no aliases for delete command, got %v", cmd.Aliases)
	}
}

func TestStatusPageDeleteCmd_EmptyID(t *testing.T) {
	cmd := NewStatusPageDeleteCmd()
	cmd.SetArgs([]string{"", "--yes"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for empty ID, got nil")
	}

	expectedMsg := "invalid status page ID"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestStatusPageDeleteCmd_ZeroID(t *testing.T) {
	cmd := NewStatusPageDeleteCmd()
	cmd.SetArgs([]string{"0", "--yes"})

	// Zero is a valid uint but likely invalid for an ID
	// The API will reject it, but the command should accept it syntactically
	err := cmd.Execute()

	// Should NOT fail with invalid ID parsing error
	if err != nil && strings.Contains(err.Error(), "invalid status page ID") {
		t.Errorf("Should not get parsing error for zero ID: %v", err)
	}
}
