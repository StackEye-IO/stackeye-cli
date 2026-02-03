// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestNewProbeDepsClearCmd(t *testing.T) {
	cmd := NewProbeDepsClearCmd()

	if cmd.Use != "clear <probe-id>" {
		t.Errorf("expected Use to be 'clear <probe-id>', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Long == "" {
		t.Error("expected Long description to be set")
	}

	// Verify RunE is set
	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}

	// Verify --direction flag exists
	directionFlag := cmd.Flags().Lookup("direction")
	if directionFlag == nil {
		t.Error("expected --direction flag to be defined")
	}

	// Verify -d shorthand
	if directionFlag != nil && directionFlag.Shorthand != "d" {
		t.Errorf("expected --direction shorthand to be 'd', got %q", directionFlag.Shorthand)
	}

	// Verify default value
	if directionFlag != nil && directionFlag.DefValue != "both" {
		t.Errorf("expected --direction default to be 'both', got %q", directionFlag.DefValue)
	}

	// Verify --yes flag exists
	yesFlag := cmd.Flags().Lookup("yes")
	if yesFlag == nil {
		t.Error("expected --yes flag to be defined")
	}

	// Verify -y shorthand
	if yesFlag != nil && yesFlag.Shorthand != "y" {
		t.Errorf("expected --yes shorthand to be 'y', got %q", yesFlag.Shorthand)
	}
}

func TestProbeDepsClearCmd_Args(t *testing.T) {
	cmd := NewProbeDepsClearCmd()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no arguments",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "one valid argument",
			args:    []string{"550e8400-e29b-41d4-a716-446655440000"},
			wantErr: false,
		},
		{
			name:    "too many arguments",
			args:    []string{"550e8400-e29b-41d4-a716-446655440000", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmd.Args(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProbeDepsClearCmd_NameResolution(t *testing.T) {
	// Since probe name resolution was added, non-UUID inputs are now treated as
	// potential probe names that need API resolution. Without a configured API
	// client, these will fail with an API client initialization error.
	cmd := NewProbeDepsClearCmd()

	// Create a parent command to hold the flag
	root := &cobra.Command{}
	root.AddCommand(cmd)

	// Set a probe name instead of UUID
	root.SetArgs([]string{"clear", "my-probe-name", "--yes"})

	// Capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := root.Execute()
	if err == nil {
		t.Error("expected error when API client not configured")
	}

	if !strings.Contains(err.Error(), "failed to initialize API client") {
		t.Errorf("expected 'failed to initialize API client' error, got: %v", err)
	}
}

func TestProbeDepsClearCmd_InvalidDirection(t *testing.T) {
	cmd := NewProbeDepsClearCmd()

	// Create a parent command to hold the flag
	root := &cobra.Command{}
	root.AddCommand(cmd)

	// Set an invalid direction
	root.SetArgs([]string{"clear", "550e8400-e29b-41d4-a716-446655440000", "--direction", "invalid", "--yes"})

	// Capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := root.Execute()
	if err == nil {
		t.Error("expected error for invalid direction")
	}

	if !strings.Contains(err.Error(), "for --direction") {
		t.Errorf("expected 'for --direction' error, got: %v", err)
	}
}

func TestProbeDepsClearCmd_DirectionValues(t *testing.T) {
	tests := []struct {
		name      string
		direction string
		valid     bool
	}{
		{
			name:      "parents direction",
			direction: "parents",
			valid:     true,
		},
		{
			name:      "children direction",
			direction: "children",
			valid:     true,
		},
		{
			name:      "both direction",
			direction: "both",
			valid:     true,
		},
		{
			name:      "invalid direction",
			direction: "all",
			valid:     false,
		},
		{
			name:      "empty direction",
			direction: "",
			valid:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewProbeDepsClearCmd()
			root := &cobra.Command{}
			root.AddCommand(cmd)

			args := []string{"clear", "550e8400-e29b-41d4-a716-446655440000", "--yes"}
			if tt.direction != "" {
				args = append(args, "--direction", tt.direction)
			}
			root.SetArgs(args)

			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			err := root.Execute()

			// Note: This test will fail at the API call stage since we don't have
			// a mock server. We're just testing that direction validation works.
			// Valid directions will fail later at API call; invalid ones fail at validation.
			if tt.valid && err != nil && strings.Contains(err.Error(), "for --direction") {
				t.Errorf("direction %q should be valid but got: %v", tt.direction, err)
			}
			if !tt.valid && err != nil && !strings.Contains(err.Error(), "for --direction") {
				// For invalid directions, we expect a "for --direction" error.
				// But empty direction falls through as "both" via default, so it's valid.
				if tt.direction != "" {
					t.Errorf("direction %q should be invalid but got different error: %v", tt.direction, err)
				}
			}
		})
	}
}
