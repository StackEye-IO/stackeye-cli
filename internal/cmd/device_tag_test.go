// Package cmd implements the CLI commands for StackEye.
// Task stackeye-5859.
package cmd

import (
	"context"
	"strings"
	"testing"
)

func TestNewDeviceTagCmd(t *testing.T) {
	cmd := NewDeviceTagCmd()

	if cmd.Use != "tag <device-id> <labels...>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "tag <device-id> <labels...>")
	}
	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}
	if cmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestDeviceTagCmd_HasListSubcommand(t *testing.T) {
	cmd := NewDeviceTagCmd()
	found := false
	for _, sub := range cmd.Commands() {
		if sub.Name() == "list" {
			found = true
		}
	}
	if !found {
		t.Error("expected device tag to register a list subcommand")
	}
}

func TestDeviceTagCmd_NoArgs(t *testing.T) {
	cmd := NewDeviceTagCmd()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when no arguments provided, got nil")
	}
	expectedMsg := "requires at least 2 arg"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestDeviceTagCmd_OnlyDeviceID(t *testing.T) {
	cmd := NewDeviceTagCmd()
	cmd.SetArgs([]string{"web-01"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when only device ID provided, got nil")
	}
	expectedMsg := "requires at least 2 arg"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestRunDeviceTag_RequiresClient(t *testing.T) {
	// No real backend/config exists in the test environment: this fails either
	// at client-init or, if a sibling test left a fake authenticated config in
	// place, at the HTTP call itself — either way, no error is unexpected.
	err := runDeviceTag(context.Background(), "web-01", []string{"env=production"})
	if err == nil {
		t.Fatal("expected an error with no real backend configured, got nil")
	}
}
