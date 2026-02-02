// Package cmd implements the CLI commands for StackEye.
// Task #8067
package cmd

import (
	"strings"
	"testing"
)

func TestNewLabelDeleteCmd(t *testing.T) {
	cmd := NewLabelDeleteCmd()

	if cmd.Use != "delete <key>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "delete <key>")
	}

	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if cmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestLabelDeleteCmd_YesFlag(t *testing.T) {
	cmd := NewLabelDeleteCmd()

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

func TestLabelDeleteCmd_ForceFlag(t *testing.T) {
	cmd := NewLabelDeleteCmd()

	flag := cmd.Flags().Lookup("force")
	if flag == nil {
		t.Error("Expected --force flag to exist")
		return
	}

	if flag.Shorthand != "" {
		t.Errorf("--force should have no shorthand, got %q", flag.Shorthand)
	}

	if flag.DefValue != "false" {
		t.Errorf("--force default = %q, want %q", flag.DefValue, "false")
	}
}

func TestLabelDeleteCmd_NoArgs(t *testing.T) {
	cmd := NewLabelDeleteCmd()
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

func TestLabelDeleteCmd_TooManyArgs(t *testing.T) {
	cmd := NewLabelDeleteCmd()
	cmd.SetArgs([]string{"env", "tier"})

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

func TestLabelDeleteCmd_InvalidKey(t *testing.T) {
	cmd := NewLabelDeleteCmd()
	// Use --yes to skip confirmation prompt
	cmd.SetArgs([]string{"Invalid-KEY", "--yes"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid key format, got nil")
	}

	expectedMsg := "invalid key format"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestLabelDeleteCmd_KeyTooLong(t *testing.T) {
	cmd := NewLabelDeleteCmd()
	// Create a key longer than 63 characters
	longKey := strings.Repeat("a", 64)
	cmd.SetArgs([]string{longKey, "--yes"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for key too long, got nil")
	}

	expectedMsg := "at most 63 characters"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestLabelDeleteCmd_EmptyKey(t *testing.T) {
	cmd := NewLabelDeleteCmd()
	cmd.SetArgs([]string{"", "--yes"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for empty key, got nil")
	}

	expectedMsg := "label key is required"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestLabelDeleteCmd_ValidKey(t *testing.T) {
	cmd := NewLabelDeleteCmd()

	// Key is syntactically valid - command should accept it
	// (actual deletion will fail without API, but parsing should succeed)
	cmd.SetArgs([]string{"env", "--yes"})

	err := cmd.Execute()
	// Will fail at API call stage (no API client configured in tests),
	// but that's expected. We're testing argument parsing here.
	if err == nil {
		t.Log("Note: Command succeeded (expected in some test environments)")
	}

	// Should NOT fail with invalid key format error
	if err != nil && strings.Contains(err.Error(), "invalid key format") {
		t.Errorf("Should not get invalid key error for valid key: %v", err)
	}
}

func TestLabelDeleteCmd_Aliases(t *testing.T) {
	cmd := NewLabelDeleteCmd()

	// delete command doesn't define aliases (direct subcommand of label)
	if len(cmd.Aliases) != 0 {
		t.Errorf("Expected no aliases for delete command, got %v", cmd.Aliases)
	}
}

func TestLabelDeleteCmd_ValidKeyWithHyphens(t *testing.T) {
	cmd := NewLabelDeleteCmd()

	// Keys can contain hyphens
	cmd.SetArgs([]string{"service-tier", "--force"})

	err := cmd.Execute()
	// Will fail at API call stage, but parsing should succeed
	if err == nil {
		t.Log("Note: Command succeeded (expected in some test environments)")
	}

	// Should NOT fail with invalid key format error
	if err != nil && strings.Contains(err.Error(), "invalid key format") {
		t.Errorf("Should not get invalid key error for valid key with hyphens: %v", err)
	}
}
