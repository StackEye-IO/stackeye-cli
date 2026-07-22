// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"
)

func TestNewEnrollmentKeyRevokeCmd(t *testing.T) {
	cmd := NewEnrollmentKeyRevokeCmd()

	if cmd.Use != "revoke <id>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "revoke <id>")
	}
	if flag := cmd.Flags().Lookup("yes"); flag == nil {
		t.Error("expected --yes flag to be defined")
	} else if flag.Shorthand != "y" {
		t.Errorf("--yes shorthand = %q, want %q", flag.Shorthand, "y")
	}
}

func TestEnrollmentKeyRevokeCmd_Aliases(t *testing.T) {
	cmd := NewEnrollmentKeyRevokeCmd()
	expected := []string{"rm", "remove"}
	if len(cmd.Aliases) != len(expected) {
		t.Fatalf("expected %d aliases, got %d: %v", len(expected), len(cmd.Aliases), cmd.Aliases)
	}
	for i, alias := range expected {
		if cmd.Aliases[i] != alias {
			t.Errorf("Alias[%d] = %q, want %q", i, cmd.Aliases[i], alias)
		}
	}
}

func TestEnrollmentKeyRevokeCmd_NoArgs(t *testing.T) {
	cmd := NewEnrollmentKeyRevokeCmd()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when no arguments provided, got nil")
	}
	if !strings.Contains(err.Error(), "accepts 1 arg") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "accepts 1 arg")
	}
}

func TestEnrollmentKeyRevokeCmd_TooManyArgs(t *testing.T) {
	cmd := NewEnrollmentKeyRevokeCmd()
	cmd.SetArgs([]string{"id-one", "id-two"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when too many arguments provided, got nil")
	}
	if !strings.Contains(err.Error(), "accepts 1 arg") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "accepts 1 arg")
	}
}

func TestEnrollmentKeyRevokeCmd_RequiresClient(t *testing.T) {
	cmd := NewEnrollmentKeyRevokeCmd()
	// --yes skips the confirmation prompt so this exercises the API-client path.
	// No real backend/config exists in the test environment: this fails either
	// at client-init or, if a sibling test left a fake authenticated config in
	// place, at the HTTP call itself — either way, no error is unexpected.
	cmd.SetArgs([]string{"a1b2c3d4-e5f6-7890-abcd-ef1234567890", "--yes"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected an error with no real backend configured, got nil")
	}
}
