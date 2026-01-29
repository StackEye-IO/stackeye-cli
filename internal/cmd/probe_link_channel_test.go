// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"
)

func TestNewProbeLinkChannelCmd(t *testing.T) {
	cmd := NewProbeLinkChannelCmd()

	if cmd.Use != "link-channel <probe-id> <channel-id>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "link-channel <probe-id> <channel-id>")
	}

	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if cmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestProbeLinkChannelCmd_NoArgs(t *testing.T) {
	cmd := NewProbeLinkChannelCmd()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when no arguments provided, got nil")
	}

	// Cobra's ExactArgs(2) produces a specific error message
	expectedMsg := "accepts 2 arg(s)"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestProbeLinkChannelCmd_OneArg(t *testing.T) {
	cmd := NewProbeLinkChannelCmd()
	cmd.SetArgs([]string{"550e8400-e29b-41d4-a716-446655440000"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when only one argument provided, got nil")
	}

	expectedMsg := "accepts 2 arg(s)"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestProbeLinkChannelCmd_InvalidProbeUUID(t *testing.T) {
	cmd := NewProbeLinkChannelCmd()
	cmd.SetArgs([]string{"not-a-uuid", "550e8400-e29b-41d4-a716-446655440000"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid probe UUID, got nil")
	}

	expectedMsg := "invalid probe ID"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestProbeLinkChannelCmd_InvalidChannelUUID(t *testing.T) {
	cmd := NewProbeLinkChannelCmd()
	cmd.SetArgs([]string{"550e8400-e29b-41d4-a716-446655440000", "not-a-uuid"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid channel UUID, got nil")
	}

	expectedMsg := "invalid channel ID"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestProbeLinkChannelCmd_ValidUUIDs(t *testing.T) {
	cmd := NewProbeLinkChannelCmd()
	cmd.SetArgs([]string{
		"550e8400-e29b-41d4-a716-446655440000",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c8",
	})

	err := cmd.Execute()
	// Will fail at API call stage (no API client configured in tests),
	// but that's expected. We're testing argument parsing here.
	if err == nil {
		t.Log("Note: Command succeeded (expected in some test environments)")
	}

	// Should NOT fail with invalid UUID error
	if err != nil && strings.Contains(err.Error(), "invalid probe ID") {
		t.Errorf("Should not get invalid probe UUID error for valid UUIDs: %v", err)
	}
	if err != nil && strings.Contains(err.Error(), "invalid channel ID") {
		t.Errorf("Should not get invalid channel UUID error for valid UUIDs: %v", err)
	}
}

func TestProbeLinkChannelCmd_LongDescription(t *testing.T) {
	cmd := NewProbeLinkChannelCmd()

	// Verify long description contains key information
	longDesc := cmd.Long

	expectedPhrases := []string{
		"Link a notification channel",
		"alerts",
		"probe get",
		"channel list",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(longDesc, phrase) {
			t.Errorf("Long description should contain %q", phrase)
		}
	}
}

func TestProbeLinkChannelCmd_Aliases(t *testing.T) {
	cmd := NewProbeLinkChannelCmd()

	// link-channel command doesn't define aliases
	if len(cmd.Aliases) != 0 {
		t.Errorf("Expected no aliases for link-channel command, got %v", cmd.Aliases)
	}
}
