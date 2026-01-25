// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"
)

// TestNewAdminWorkerKeyDeleteCmd verifies that the worker-key delete command is properly constructed.
func TestNewAdminWorkerKeyDeleteCmd(t *testing.T) {
	cmd := NewAdminWorkerKeyDeleteCmd()

	if cmd.Use != "delete <key-id>" {
		t.Errorf("expected Use='delete <key-id>', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Short != "Delete a worker key" {
		t.Errorf("expected Short='Delete a worker key', got %q", cmd.Short)
	}
}

// TestNewAdminWorkerKeyDeleteCmd_Long verifies the Long description contains key information.
func TestNewAdminWorkerKeyDeleteCmd_Long(t *testing.T) {
	cmd := NewAdminWorkerKeyDeleteCmd()

	long := cmd.Long

	// Should contain key features
	features := []string{
		"permanently delete",
		"worker key",
		"authentication",
		"irreversible",
		"deactivate",
	}
	for _, feature := range features {
		if !strings.Contains(strings.ToLower(long), strings.ToLower(feature)) {
			t.Errorf("expected Long description to mention %q", feature)
		}
	}

	// Should have usage examples
	if !strings.Contains(long, "stackeye admin worker-key delete") {
		t.Error("expected Long description to contain example commands")
	}

	// Should mention force flag
	if !strings.Contains(long, "--force") {
		t.Error("expected Long description to mention --force flag")
	}
}

// TestNewAdminWorkerKeyDeleteCmd_RunEIsSet verifies that RunE is properly configured.
func TestNewAdminWorkerKeyDeleteCmd_RunEIsSet(t *testing.T) {
	cmd := NewAdminWorkerKeyDeleteCmd()

	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}
}

// TestNewAdminWorkerKeyDeleteCmd_ArgsValidation verifies that exactly one argument is required.
func TestNewAdminWorkerKeyDeleteCmd_ArgsValidation(t *testing.T) {
	cmd := NewAdminWorkerKeyDeleteCmd()

	// Test with no arguments
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("expected error with no arguments")
	}

	// Test with one argument
	err = cmd.Args(cmd, []string{"550e8400-e29b-41d4-a716-446655440000"})
	if err != nil {
		t.Errorf("expected no error with one argument, got %v", err)
	}

	// Test with too many arguments
	err = cmd.Args(cmd, []string{"key1", "key2"})
	if err == nil {
		t.Error("expected error with too many arguments")
	}
}

// TestNewAdminWorkerKeyDeleteCmd_ForceFlag verifies that the force flag is registered.
func TestNewAdminWorkerKeyDeleteCmd_ForceFlag(t *testing.T) {
	cmd := NewAdminWorkerKeyDeleteCmd()

	flag := cmd.Flags().Lookup("force")
	if flag == nil {
		t.Fatal("expected --force flag to be registered")
	}

	if flag.Shorthand != "f" {
		t.Errorf("expected force flag shorthand to be 'f', got %q", flag.Shorthand)
	}

	if flag.DefValue != "false" {
		t.Errorf("expected force flag default value to be 'false', got %q", flag.DefValue)
	}

	// Check flag usage contains description
	if flag.Usage == "" {
		t.Error("expected force flag to have a usage description")
	}
}

// TestPrintWorkerKeyDeleted_DoesNotPanic verifies that printWorkerKeyDeleted doesn't panic.
func TestPrintWorkerKeyDeleted_DoesNotPanic(t *testing.T) {
	tests := []struct {
		name  string
		keyID string
	}{
		{"standard uuid", "550e8400-e29b-41d4-a716-446655440000"},
		{"short id", "abc123"},
		{"empty id", ""},
		{"very long id", strings.Repeat("a", 100)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("printWorkerKeyDeleted panicked: %v", r)
				}
			}()
			printWorkerKeyDeleted(tt.keyID)
		})
	}
}

// TestNewAdminWorkerKeyCmd_HasDeleteSubcommand verifies that delete subcommand is registered.
func TestNewAdminWorkerKeyCmd_HasDeleteSubcommand(t *testing.T) {
	cmd := NewAdminWorkerKeyCmd()

	subcommands := cmd.Commands()
	if len(subcommands) < 2 {
		t.Errorf("expected worker-key command to have at least 2 subcommands (create, delete), got %d", len(subcommands))
	}

	// Verify delete subcommand is registered
	found := false
	for _, sub := range subcommands {
		if sub.Use == "delete <key-id>" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'delete <key-id>' subcommand to be registered")
	}
}
