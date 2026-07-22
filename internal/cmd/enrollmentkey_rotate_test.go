// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"
)

func TestNewEnrollmentKeyRotateCmd(t *testing.T) {
	cmd := NewEnrollmentKeyRotateCmd()

	if cmd.Use != "rotate <id>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "rotate <id>")
	}
	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestEnrollmentKeyRotateCmd_NoArgs(t *testing.T) {
	cmd := NewEnrollmentKeyRotateCmd()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when no arguments provided, got nil")
	}
	if !strings.Contains(err.Error(), "accepts 1 arg") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "accepts 1 arg")
	}
}

func TestEnrollmentKeyRotateCmd_RequiresClient(t *testing.T) {
	// No real backend/config exists in the test environment: this fails either
	// at client-init or, if a sibling test left a fake authenticated config in
	// place, at the HTTP call itself — either way, no error is unexpected.
	cmd := NewEnrollmentKeyRotateCmd()
	cmd.SetArgs([]string{"a1b2c3d4-e5f6-7890-abcd-ef1234567890"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected an error with no real backend configured, got nil")
	}
}
