// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"
)

func TestNewChannelDeleteCmd(t *testing.T) {
	cmd := NewChannelDeleteCmd()

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

func TestChannelDeleteCmd_YesFlag(t *testing.T) {
	cmd := NewChannelDeleteCmd()

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

func TestChannelDeleteCmd_NoArgs(t *testing.T) {
	cmd := NewChannelDeleteCmd()
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

func TestChannelDeleteCmd_TooManyArgs(t *testing.T) {
	cmd := NewChannelDeleteCmd()
	cmd.SetArgs([]string{
		"550e8400-e29b-41d4-a716-446655440000",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c8",
	})

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

func TestChannelDeleteCmd_InvalidUUID(t *testing.T) {
	cmd := NewChannelDeleteCmd()
	// Use --yes to skip confirmation prompt
	cmd.SetArgs([]string{"not-a-uuid", "--yes"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid UUID, got nil")
	}

	expectedMsg := "invalid channel ID"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestChannelDeleteCmd_ValidUUID(t *testing.T) {
	cmd := NewChannelDeleteCmd()

	// UUID is syntactically valid - command should accept it
	// (actual deletion will fail without API, but parsing should succeed)
	cmd.SetArgs([]string{
		"550e8400-e29b-41d4-a716-446655440000",
		"--yes",
	})

	err := cmd.Execute()
	// Will fail at API call stage (no API client configured in tests),
	// but that's expected. We're testing argument parsing here.
	if err == nil {
		t.Log("Note: Command succeeded (expected in some test environments)")
	}

	// Should NOT fail with invalid UUID error
	if err != nil && strings.Contains(err.Error(), "invalid channel ID") {
		t.Errorf("Should not get invalid UUID error for valid UUID: %v", err)
	}
}

func TestChannelDeleteCmd_Aliases(t *testing.T) {
	cmd := NewChannelDeleteCmd()

	// delete command doesn't define aliases (direct subcommand of channel)
	if len(cmd.Aliases) != 0 {
		t.Errorf("Expected no aliases for delete command, got %v", cmd.Aliases)
	}
}

func TestChannelDeleteCmd_EmptyUUID(t *testing.T) {
	cmd := NewChannelDeleteCmd()
	cmd.SetArgs([]string{"", "--yes"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for empty UUID, got nil")
	}

	expectedMsg := "invalid channel ID"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}
