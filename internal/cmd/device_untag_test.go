// Package cmd implements the CLI commands for StackEye.
// Task stackeye-5859.
package cmd

import (
	"context"
	"strings"
	"testing"
)

func TestNewDeviceUntagCmd(t *testing.T) {
	cmd := NewDeviceUntagCmd()

	if cmd.Use != "untag <device-id> <keys...>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "untag <device-id> <keys...>")
	}
}

func TestDeviceUntagCmd_NoArgs(t *testing.T) {
	cmd := NewDeviceUntagCmd()
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

func TestRunDeviceUntag_RequiresClient(t *testing.T) {
	err := runDeviceUntag(context.Background(), "web-01", []string{"env"})
	if err == nil {
		t.Fatal("expected an error with no real backend configured, got nil")
	}
}
