// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"
)

func TestNewChannelTestCmd(t *testing.T) {
	cmd := NewChannelTestCmd()

	if cmd.Use != "test <id>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "test <id>")
	}

	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if cmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestChannelTestCmd_NoArgs(t *testing.T) {
	cmd := NewChannelTestCmd()
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

func TestChannelTestCmd_TooManyArgs(t *testing.T) {
	cmd := NewChannelTestCmd()
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

func TestChannelTestCmd_InvalidUUID(t *testing.T) {
	cmd := NewChannelTestCmd()
	cmd.SetArgs([]string{"not-a-uuid"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid UUID, got nil")
	}

	expectedMsg := "invalid channel ID"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestChannelTestCmd_ValidUUID(t *testing.T) {
	cmd := NewChannelTestCmd()

	// UUID is syntactically valid - command should accept it
	// (actual test will fail without API, but parsing should succeed)
	cmd.SetArgs([]string{"550e8400-e29b-41d4-a716-446655440000"})

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

func TestChannelTestCmd_EmptyUUID(t *testing.T) {
	cmd := NewChannelTestCmd()
	cmd.SetArgs([]string{""})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for empty UUID, got nil")
	}

	expectedMsg := "invalid channel ID"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestChannelTestCmd_LongDescriptionContent(t *testing.T) {
	cmd := NewChannelTestCmd()

	// Verify help content includes key information
	expectedContent := []string{
		"test notification",
		"channel",
		"webhook",
		"email",
		"Slack",
	}

	for _, content := range expectedContent {
		if !strings.Contains(cmd.Long, content) {
			t.Errorf("Long description should mention %q", content)
		}
	}
}

func TestChannelTestResult_FieldsExist(t *testing.T) {
	// Test that ChannelTestResult can be instantiated with expected fields
	errMsg := "test error"
	result := ChannelTestResult{
		ChannelID:      [16]byte{}, // Zero UUID
		ChannelName:    "Test Channel",
		ChannelType:    "email",
		Success:        true,
		Message:        "Test notification sent",
		Error:          &errMsg,
		ResponseTimeMs: 150,
	}

	if result.ChannelName != "Test Channel" {
		t.Errorf("ChannelName = %q, want %q", result.ChannelName, "Test Channel")
	}

	if result.ChannelType != "email" {
		t.Errorf("ChannelType = %q, want %q", result.ChannelType, "email")
	}

	if !result.Success {
		t.Error("Success should be true")
	}

	if result.ResponseTimeMs != 150 {
		t.Errorf("ResponseTimeMs = %d, want %d", result.ResponseTimeMs, 150)
	}

	if result.Error == nil || *result.Error != "test error" {
		t.Errorf("Error = %v, want %q", result.Error, "test error")
	}
}
