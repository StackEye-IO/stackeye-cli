// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"
)

func TestNewProbeDeleteCmd(t *testing.T) {
	cmd := NewProbeDeleteCmd()

	if cmd.Use != "delete <id> [id...]" {
		t.Errorf("Use = %q, want %q", cmd.Use, "delete <id> [id...]")
	}

	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if cmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestProbeDeleteCmd_YesFlag(t *testing.T) {
	cmd := NewProbeDeleteCmd()

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

func TestProbeDeleteCmd_NoArgs(t *testing.T) {
	cmd := NewProbeDeleteCmd()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when no arguments provided, got nil")
	}

	// Cobra's MinimumNArgs(1) produces a specific error message
	expectedMsg := "requires at least 1 arg"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestProbeDeleteCmd_InvalidUUID(t *testing.T) {
	cmd := NewProbeDeleteCmd()
	// Use --yes to skip confirmation prompt
	cmd.SetArgs([]string{"not-a-uuid", "--yes"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid UUID, got nil")
	}

	expectedMsg := "invalid probe ID"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestProbeDeleteCmd_InvalidUUIDInBatch(t *testing.T) {
	cmd := NewProbeDeleteCmd()
	// First UUID is valid, second is invalid - should fail on validation before any deletes
	cmd.SetArgs([]string{
		"550e8400-e29b-41d4-a716-446655440000",
		"invalid-uuid",
		"--yes",
	})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid UUID in batch, got nil")
	}

	expectedMsg := "invalid probe ID"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestProbeDeleteCmd_MultipleValidUUIDs(t *testing.T) {
	cmd := NewProbeDeleteCmd()

	// Both UUIDs are syntactically valid - command should accept them
	// (actual deletion will fail without API, but parsing should succeed)
	cmd.SetArgs([]string{
		"550e8400-e29b-41d4-a716-446655440000",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		"--yes",
	})

	err := cmd.Execute()
	// Will fail at API call stage (no API client configured in tests),
	// but that's expected. We're testing argument parsing here.
	if err == nil {
		t.Log("Note: Command succeeded (expected in some test environments)")
	}

	// Should NOT fail with invalid UUID error
	if err != nil && strings.Contains(err.Error(), "invalid probe ID") {
		t.Errorf("Should not get invalid UUID error for valid UUIDs: %v", err)
	}
}

func TestProbeDeleteCmd_Aliases(t *testing.T) {
	cmd := NewProbeDeleteCmd()

	// delete command doesn't define aliases (direct subcommand of probe)
	if len(cmd.Aliases) != 0 {
		t.Errorf("Expected no aliases for delete command, got %v", cmd.Aliases)
	}
}
