// Package cmd implements the CLI commands for StackEye.
// Task #8065
package cmd

import (
	"testing"
)

func TestNewLabelListCmd(t *testing.T) {
	cmd := NewLabelListCmd()

	if cmd == nil {
		t.Fatal("expected non-nil command")
	}

	// Verify command structure
	if cmd.Use != "list" {
		t.Errorf("expected Use='list', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected non-empty Short description")
	}

	if cmd.Long == "" {
		t.Error("expected non-empty Long description")
	}

	// Verify alias
	aliases := cmd.Aliases
	if len(aliases) != 1 || aliases[0] != "ls" {
		t.Errorf("expected aliases=['ls'], got %v", aliases)
	}

	// Verify RunE is set
	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}
}
