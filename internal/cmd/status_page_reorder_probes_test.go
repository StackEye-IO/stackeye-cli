// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"
)

// testReorderProbeUUID1 is a test fixture UUID for reorder-probes command tests.
const testReorderProbeUUID1 = "550e8400-e29b-41d4-a716-446655440001"

// testReorderProbeUUID2 is a second test fixture UUID for reorder-probes command tests.
const testReorderProbeUUID2 = "550e8400-e29b-41d4-a716-446655440002"

// testReorderProbeUUID3 is a third test fixture UUID for reorder-probes command tests.
const testReorderProbeUUID3 = "550e8400-e29b-41d4-a716-446655440003"

func TestNewStatusPageReorderProbesCmd(t *testing.T) {
	cmd := NewStatusPageReorderProbesCmd()

	if cmd.Use != "reorder-probes <status-page-id>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "reorder-probes <status-page-id>")
	}

	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if cmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestStatusPageReorderProbesCmd_ProbeIDsFlag(t *testing.T) {
	cmd := NewStatusPageReorderProbesCmd()

	flag := cmd.Flags().Lookup("probe-ids")
	if flag == nil {
		t.Error("Expected --probe-ids flag to exist")
		return
	}

	if flag.DefValue != "" {
		t.Errorf("--probe-ids default = %q, want empty string", flag.DefValue)
	}
}

func TestStatusPageReorderProbesCmd_NoArgs(t *testing.T) {
	cmd := NewStatusPageReorderProbesCmd()
	cmd.SetArgs([]string{"--probe-ids", testReorderProbeUUID1})

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

func TestStatusPageReorderProbesCmd_TooManyArgs(t *testing.T) {
	cmd := NewStatusPageReorderProbesCmd()
	cmd.SetArgs([]string{"123", "456", "--probe-ids", testReorderProbeUUID1})

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

func TestStatusPageReorderProbesCmd_MissingProbeIDs(t *testing.T) {
	cmd := NewStatusPageReorderProbesCmd()
	cmd.SetArgs([]string{"123"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when --probe-ids not provided, got nil")
	}

	expectedMsg := "required flag"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestStatusPageReorderProbesCmd_InvalidID(t *testing.T) {
	cmd := NewStatusPageReorderProbesCmd()
	cmd.SetArgs([]string{"not-a-number", "--probe-ids", testReorderProbeUUID1})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid ID, got nil")
	}

	expectedMsg := "invalid status page ID"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestStatusPageReorderProbesCmd_InvalidProbeUUID(t *testing.T) {
	cmd := NewStatusPageReorderProbesCmd()
	cmd.SetArgs([]string{"123", "--probe-ids", "not-a-valid-uuid"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid probe UUID, got nil")
	}

	expectedMsg := "invalid probe ID"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestStatusPageReorderProbesCmd_InvalidUUIDInList(t *testing.T) {
	cmd := NewStatusPageReorderProbesCmd()
	// Second UUID in the list is invalid
	cmd.SetArgs([]string{"123", "--probe-ids", testReorderProbeUUID1 + ",invalid-uuid," + testReorderProbeUUID3})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid UUID in list, got nil")
	}

	expectedMsg := "invalid probe ID"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}

	// Should indicate position 2 (1-indexed)
	if err != nil && !strings.Contains(err.Error(), "position 2") {
		t.Errorf("Error = %q, want to contain position info", err.Error())
	}
}

func TestStatusPageReorderProbesCmd_NegativeID(t *testing.T) {
	cmd := NewStatusPageReorderProbesCmd()
	// Use "--" to signal end of flags so "-1" is treated as positional argument
	cmd.SetArgs([]string{"--probe-ids", testReorderProbeUUID1, "--", "-1"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for negative ID, got nil")
	}

	expectedMsg := "invalid status page ID"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestStatusPageReorderProbesCmd_ZeroID(t *testing.T) {
	cmd := NewStatusPageReorderProbesCmd()
	cmd.SetArgs([]string{"0", "--probe-ids", testReorderProbeUUID1})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for zero ID, got nil")
	}

	expectedMsg := "invalid status page ID"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestStatusPageReorderProbesCmd_EmptyID(t *testing.T) {
	cmd := NewStatusPageReorderProbesCmd()
	cmd.SetArgs([]string{"", "--probe-ids", testReorderProbeUUID1})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for empty ID, got nil")
	}

	expectedMsg := "invalid status page ID"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestStatusPageReorderProbesCmd_ValidSingleProbe(t *testing.T) {
	cmd := NewStatusPageReorderProbesCmd()
	cmd.SetArgs([]string{"123", "--probe-ids", testReorderProbeUUID1})

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

	// Should NOT fail with invalid probe ID error
	if err != nil && strings.Contains(err.Error(), "invalid probe ID") {
		t.Errorf("Should not get invalid probe ID error for valid UUID: %v", err)
	}
}

func TestStatusPageReorderProbesCmd_ValidMultipleProbes(t *testing.T) {
	cmd := NewStatusPageReorderProbesCmd()
	probeList := testReorderProbeUUID1 + "," + testReorderProbeUUID2 + "," + testReorderProbeUUID3
	cmd.SetArgs([]string{"123", "--probe-ids", probeList})

	err := cmd.Execute()
	// Will fail at API call stage (no API client configured in tests),
	// but that's expected. We're testing argument parsing here.
	if err == nil {
		t.Log("Note: Command succeeded (expected in some test environments)")
	}

	// Should NOT fail with validation errors
	if err != nil && strings.Contains(err.Error(), "invalid status page ID") {
		t.Errorf("Should not get invalid status page ID error for valid ID: %v", err)
	}
	if err != nil && strings.Contains(err.Error(), "invalid probe ID") {
		t.Errorf("Should not get invalid probe ID error for valid UUIDs: %v", err)
	}
}

func TestStatusPageReorderProbesCmd_Aliases(t *testing.T) {
	cmd := NewStatusPageReorderProbesCmd()

	// reorder-probes command doesn't define aliases
	if len(cmd.Aliases) != 0 {
		t.Errorf("Expected no aliases for reorder-probes command, got %v", cmd.Aliases)
	}
}

func TestStatusPageReorderProbesCmd_EmptyProbeIDs(t *testing.T) {
	cmd := NewStatusPageReorderProbesCmd()
	cmd.SetArgs([]string{"123", "--probe-ids", ""})

	err := cmd.Execute()
	// Empty string should fail validation
	if err == nil {
		t.Error("Expected error for empty --probe-ids, got nil")
	}

	expectedMsg := "--probe-ids is required"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestStatusPageReorderProbesCmd_TrailingComma(t *testing.T) {
	cmd := NewStatusPageReorderProbesCmd()
	// Trailing comma should be handled gracefully
	cmd.SetArgs([]string{"123", "--probe-ids", testReorderProbeUUID1 + ","})

	err := cmd.Execute()
	// Will fail at API call stage, but should NOT fail with validation errors
	if err == nil {
		t.Log("Note: Command succeeded (expected in some test environments)")
	}

	// Should NOT fail with invalid probe ID error
	if err != nil && strings.Contains(err.Error(), "invalid probe ID") {
		t.Errorf("Trailing comma should be handled gracefully: %v", err)
	}
}

func TestStatusPageReorderProbesCmd_OnlyCommas(t *testing.T) {
	cmd := NewStatusPageReorderProbesCmd()
	cmd.SetArgs([]string{"123", "--probe-ids", ",,,"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for only commas in --probe-ids, got nil")
	}

	expectedMsg := "at least one valid probe UUID"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}
