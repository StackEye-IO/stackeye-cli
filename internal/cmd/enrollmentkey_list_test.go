// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"testing"
)

func TestNewEnrollmentKeyListCmd(t *testing.T) {
	cmd := NewEnrollmentKeyListCmd()

	if cmd.Use != "list" {
		t.Errorf("Use = %q, want %q", cmd.Use, "list")
	}
	for _, name := range []string{"limit", "offset"} {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected --%s flag to be defined", name)
		}
	}
}

func TestEnrollmentKeyListCmd_Aliases(t *testing.T) {
	cmd := NewEnrollmentKeyListCmd()
	if len(cmd.Aliases) != 1 || cmd.Aliases[0] != "ls" {
		t.Errorf("expected aliases [ls], got %v", cmd.Aliases)
	}
}

func TestRunEnrollmentKeyList_RequiresClient(t *testing.T) {
	// No real backend/config exists in the test environment: this fails either
	// at client-init or, if a sibling test left a fake authenticated config in
	// place, at the HTTP call itself — either way, no error is unexpected.
	err := runEnrollmentKeyList(context.Background(), 0, 0)
	if err == nil {
		t.Fatal("expected an error with no real backend configured, got nil")
	}
}
