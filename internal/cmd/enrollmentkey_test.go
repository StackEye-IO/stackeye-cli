// Package cmd implements the CLI commands for StackEye.
package cmd

import "testing"

func TestNewEnrollmentKeyCmd(t *testing.T) {
	cmd := NewEnrollmentKeyCmd()

	if cmd.Use != "enrollment-key" {
		t.Errorf("Use = %q, want %q", cmd.Use, "enrollment-key")
	}
	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}
	if cmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestEnrollmentKeyCmd_RegistersSubcommands(t *testing.T) {
	cmd := NewEnrollmentKeyCmd()

	expected := map[string]bool{"list": false, "create": false, "revoke <id>": false, "rotate <id>": false}
	for _, sub := range cmd.Commands() {
		if _, ok := expected[sub.Use]; ok {
			expected[sub.Use] = true
		}
	}
	for use, found := range expected {
		if !found {
			t.Errorf("expected subcommand %q to be registered under enrollment-key", use)
		}
	}
}

func TestEnrollmentKeyCmd_Aliases(t *testing.T) {
	cmd := NewEnrollmentKeyCmd()

	expected := []string{"enrollmentkey", "enrollment-keys", "ek"}
	if len(cmd.Aliases) != len(expected) {
		t.Fatalf("expected %d aliases, got %d: %v", len(expected), len(cmd.Aliases), cmd.Aliases)
	}
	for i, alias := range expected {
		if cmd.Aliases[i] != alias {
			t.Errorf("Alias[%d] = %q, want %q", i, cmd.Aliases[i], alias)
		}
	}
}
