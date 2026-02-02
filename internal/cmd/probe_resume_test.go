// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"
)

func TestNewProbeResumeCmd(t *testing.T) {
	cmd := NewProbeResumeCmd()

	if cmd.Use != "resume <id> [id...]" {
		t.Errorf("Use = %q, want %q", cmd.Use, "resume <id> [id...]")
	}

	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if cmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestProbeResumeCmd_YesFlag(t *testing.T) {
	cmd := NewProbeResumeCmd()

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

func TestProbeResumeCmd_NoArgs(t *testing.T) {
	cmd := NewProbeResumeCmd()
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

func TestProbeResumeCmd_NameResolution(t *testing.T) {
	// Since probe name resolution was added, non-UUID inputs are now treated as
	// potential probe names that need API resolution. Without a configured API
	// client, these will fail with an API client initialization error.
	cmd := NewProbeResumeCmd()
	// Use --yes to skip confirmation prompt
	cmd.SetArgs([]string{"my-probe-name", "--yes"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when API client not configured, got nil")
	}

	expectedMsg := "failed to initialize API client"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestProbeResumeCmd_MixedUUIDsAndNames(t *testing.T) {
	// With name resolution, a mix of UUIDs and names is valid input.
	// All names will be resolved via API. Without API client, this fails.
	cmd := NewProbeResumeCmd()
	cmd.SetArgs([]string{
		"550e8400-e29b-41d4-a716-446655440000",
		"my-probe-name",
		"--yes",
	})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when API client not configured, got nil")
	}

	expectedMsg := "failed to initialize API client"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestProbeResumeCmd_MultipleValidUUIDs(t *testing.T) {
	cmd := NewProbeResumeCmd()

	// Both UUIDs are syntactically valid - command should accept them
	// (actual resume will fail without API, but parsing should succeed)
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

func TestProbeResumeCmd_Aliases(t *testing.T) {
	cmd := NewProbeResumeCmd()

	// resume command doesn't define aliases (direct subcommand of probe)
	if len(cmd.Aliases) != 0 {
		t.Errorf("Expected no aliases for resume command, got %v", cmd.Aliases)
	}
}

func TestProbeResumeCmd_LongDescription(t *testing.T) {
	cmd := NewProbeResumeCmd()

	// Verify long description contains key information
	longDesc := cmd.Long

	expectedPhrases := []string{
		"Resume monitoring",
		"paused",
		"--yes",
		"checks",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(longDesc, phrase) {
			t.Errorf("Long description should contain %q", phrase)
		}
	}
}
