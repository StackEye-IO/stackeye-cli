// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"
)

func TestNewProbeUnlinkChannelCmd(t *testing.T) {
	cmd := NewProbeUnlinkChannelCmd()

	if cmd.Use != "unlink-channel <probe-id> <channel-id>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "unlink-channel <probe-id> <channel-id>")
	}

	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if cmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestProbeUnlinkChannelCmd_NoArgs(t *testing.T) {
	cmd := NewProbeUnlinkChannelCmd()
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

func TestProbeUnlinkChannelCmd_OneArg(t *testing.T) {
	cmd := NewProbeUnlinkChannelCmd()
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

func TestProbeUnlinkChannelCmd_ProbeNameResolution(t *testing.T) {
	// Since probe name resolution was added, non-UUID inputs are now treated as
	// potential probe names that need API resolution. Without a configured API
	// client, these will fail with an API client initialization error.
	cmd := NewProbeUnlinkChannelCmd()
	cmd.SetArgs([]string{"my-probe-name", "550e8400-e29b-41d4-a716-446655440000"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when API client not configured, got nil")
	}

	expectedMsg := "failed to initialize API client"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestProbeUnlinkChannelCmd_InvalidChannelUUID(t *testing.T) {
	cmd := NewProbeUnlinkChannelCmd()
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

func TestProbeUnlinkChannelCmd_ValidUUIDs(t *testing.T) {
	cmd := NewProbeUnlinkChannelCmd()
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

func TestProbeUnlinkChannelCmd_LongDescription(t *testing.T) {
	cmd := NewProbeUnlinkChannelCmd()

	// Verify long description contains key information
	longDesc := cmd.Long

	expectedPhrases := []string{
		"Unlink a notification channel",
		"stop receiving alert notifications",
		"probe get",
		"channel list",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(longDesc, phrase) {
			t.Errorf("Long description should contain %q", phrase)
		}
	}
}

func TestProbeUnlinkChannelCmd_Aliases(t *testing.T) {
	cmd := NewProbeUnlinkChannelCmd()

	// unlink-channel command doesn't define aliases
	if len(cmd.Aliases) != 0 {
		t.Errorf("Expected no aliases for unlink-channel command, got %v", cmd.Aliases)
	}
}

func TestProbeUnlinkChannelCmd_Registered(t *testing.T) {
	// Verify the unlink-channel command is registered in the parent probe command
	probeCmd := NewProbeCmd()

	var unlinkFound bool
	for _, sub := range probeCmd.Commands() {
		if sub.Use == "unlink-channel <probe-id> <channel-id>" {
			unlinkFound = true
			break
		}
	}

	if !unlinkFound {
		t.Error("Expected unlink-channel command to be registered as subcommand of probe")
	}
}
