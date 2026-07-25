// Package cmd implements the CLI commands for StackEye.
// Task stackeye-5859.
package cmd

import (
	"context"
	"testing"
)

func TestNewDeviceTagListCmd(t *testing.T) {
	cmd := NewDeviceTagListCmd()

	if cmd.Use != "list <device-id>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "list <device-id>")
	}
	if len(cmd.Aliases) != 1 || cmd.Aliases[0] != "ls" {
		t.Errorf("expected aliases [ls], got %v", cmd.Aliases)
	}
}

func TestDeviceTagListCmd_RequiresExactlyOneArg(t *testing.T) {
	cmd := NewDeviceTagListCmd()
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err == nil {
		t.Error("expected error when no device id provided, got nil")
	}
}

func TestRunDeviceTagList_RequiresClient(t *testing.T) {
	err := runDeviceTagList(context.Background(), "web-01")
	if err == nil {
		t.Fatal("expected an error with no real backend configured, got nil")
	}
}
