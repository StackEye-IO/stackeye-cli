// Package cmd implements the CLI commands for StackEye.
// Task stackeye-5859.
package cmd

import (
	"context"
	"testing"
)

func TestNewDeviceRegionCmd(t *testing.T) {
	cmd := NewDeviceRegionCmd()

	if cmd.Use != "region" {
		t.Errorf("Use = %q, want %q", cmd.Use, "region")
	}

	wantSubs := map[string]bool{"assign": false, "list": false, "unassign": false}
	for _, sub := range cmd.Commands() {
		if _, ok := wantSubs[sub.Name()]; ok {
			wantSubs[sub.Name()] = true
		}
	}
	for name, found := range wantSubs {
		if !found {
			t.Errorf("expected device region to register a %q subcommand", name)
		}
	}
}

func TestNewDeviceCmd(t *testing.T) {
	cmd := NewDeviceCmd()

	if cmd.Use != "device" {
		t.Errorf("Use = %q, want %q", cmd.Use, "device")
	}

	wantSubs := map[string]bool{"tag": false, "untag": false, "region": false}
	for _, sub := range cmd.Commands() {
		if _, ok := wantSubs[sub.Name()]; ok {
			wantSubs[sub.Name()] = true
		}
	}
	for name, found := range wantSubs {
		if !found {
			t.Errorf("expected device to register a %q subcommand", name)
		}
	}
}

func TestNewDeviceRegionAssignCmd(t *testing.T) {
	cmd := NewDeviceRegionAssignCmd()
	if cmd.Use != "assign <device-id> <region-id>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "assign <device-id> <region-id>")
	}
}

func TestDeviceRegionAssignCmd_RequiresExactlyTwoArgs(t *testing.T) {
	cmd := NewDeviceRegionAssignCmd()
	cmd.SetArgs([]string{"web-01"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error when only one argument provided, got nil")
	}
}

func TestRunDeviceRegionAssign_RequiresClient(t *testing.T) {
	err := runDeviceRegionAssign(context.Background(), "web-01", "us-east")
	if err == nil {
		t.Fatal("expected an error with no real backend configured, got nil")
	}
}

func TestNewDeviceRegionListCmd(t *testing.T) {
	cmd := NewDeviceRegionListCmd()
	if cmd.Use != "list <device-id>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "list <device-id>")
	}
	if len(cmd.Aliases) != 1 || cmd.Aliases[0] != "ls" {
		t.Errorf("expected aliases [ls], got %v", cmd.Aliases)
	}
}

func TestRunDeviceRegionList_RequiresClient(t *testing.T) {
	err := runDeviceRegionList(context.Background(), "web-01")
	if err == nil {
		t.Fatal("expected an error with no real backend configured, got nil")
	}
}

func TestNewDeviceRegionUnassignCmd(t *testing.T) {
	cmd := NewDeviceRegionUnassignCmd()
	if cmd.Use != "unassign <device-id> <region-id>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "unassign <device-id> <region-id>")
	}
}

func TestDeviceRegionUnassignCmd_RequiresExactlyTwoArgs(t *testing.T) {
	cmd := NewDeviceRegionUnassignCmd()
	cmd.SetArgs([]string{"web-01"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error when only one argument provided, got nil")
	}
}

func TestRunDeviceRegionUnassign_RequiresClient(t *testing.T) {
	err := runDeviceRegionUnassign(context.Background(), "web-01", "us-east")
	if err == nil {
		t.Fatal("expected an error with no real backend configured, got nil")
	}
}
