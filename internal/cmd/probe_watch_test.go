package cmd

import (
	"context"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestNewProbeWatchCmd(t *testing.T) {
	cmd := NewProbeWatchCmd()

	if cmd.Use != "watch [id]" {
		t.Errorf("expected Use='watch [id]', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Short != "Watch probe status with live updates" {
		t.Errorf("expected Short='Watch probe status with live updates', got %q", cmd.Short)
	}
}

func TestNewProbeWatchCmd_Aliases(t *testing.T) {
	cmd := NewProbeWatchCmd()

	expectedAliases := []string{"w"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("expected %d aliases, got %d", len(expectedAliases), len(cmd.Aliases))
	}

	for _, expected := range expectedAliases {
		if !slices.Contains(cmd.Aliases, expected) {
			t.Errorf("expected alias %q not found", expected)
		}
	}
}

func TestNewProbeWatchCmd_Long(t *testing.T) {
	cmd := NewProbeWatchCmd()

	long := cmd.Long

	// Should describe the polling behavior
	if !strings.Contains(long, "interval") {
		t.Error("expected Long description to mention interval")
	}

	// Should mention Ctrl+C
	if !strings.Contains(long, "Ctrl+C") {
		t.Error("expected Long description to mention Ctrl+C")
	}

	// Should have usage examples
	if !strings.Contains(long, "stackeye probe watch") {
		t.Error("expected Long description to contain example commands")
	}

	// Should mention non-interactive mode
	if !strings.Contains(long, "non-interactive") {
		t.Error("expected Long description to mention non-interactive mode")
	}
}

func TestNewProbeWatchCmd_Flags(t *testing.T) {
	cmd := NewProbeWatchCmd()

	flags := []struct {
		name         string
		shorthand    string
		defaultValue string
	}{
		{"interval", "i", "5s"},
		{"status", "s", ""},
	}

	for _, f := range flags {
		flag := cmd.Flags().Lookup(f.name)
		if flag == nil {
			t.Errorf("expected flag %q to exist", f.name)
			continue
		}
		if flag.DefValue != f.defaultValue {
			t.Errorf("flag %q: expected default %q, got %q", f.name, f.defaultValue, flag.DefValue)
		}
		if f.shorthand != "" && flag.Shorthand != f.shorthand {
			t.Errorf("flag %q: expected shorthand %q, got %q", f.name, f.shorthand, flag.Shorthand)
		}
	}
}

func TestNewProbeWatchCmd_Args(t *testing.T) {
	cmd := NewProbeWatchCmd()

	// Should accept zero arguments (watch all probes)
	if err := cmd.Args(cmd, []string{}); err != nil {
		t.Errorf("expected zero args to be valid, got error: %v", err)
	}

	// Should accept one argument (watch specific probe)
	if err := cmd.Args(cmd, []string{"my-probe"}); err != nil {
		t.Errorf("expected one arg to be valid, got error: %v", err)
	}

	// Should reject two arguments
	if err := cmd.Args(cmd, []string{"probe1", "probe2"}); err == nil {
		t.Error("expected two args to be rejected")
	}
}

func TestRunProbeWatch_Validation(t *testing.T) {
	tests := []struct {
		name         string
		interval     time.Duration
		status       string
		wantErrorMsg string
	}{
		{
			name:         "interval too short",
			interval:     500 * time.Millisecond,
			wantErrorMsg: "invalid interval 500ms: minimum is 1s",
		},
		{
			name:         "invalid status",
			interval:     5 * time.Second,
			status:       "badstatus",
			wantErrorMsg: `invalid value "badstatus" for --status: must be one of: up, down, degraded, paused, pending`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := &probeWatchFlags{
				interval: tt.interval,
				status:   tt.status,
			}

			err := runProbeWatch(context.Background(), "", flags)

			if err == nil {
				t.Errorf("expected error containing %q, got nil", tt.wantErrorMsg)
				return
			}

			if !strings.Contains(err.Error(), tt.wantErrorMsg) {
				t.Errorf("expected error containing %q, got %q", tt.wantErrorMsg, err.Error())
			}
		})
	}
}

func TestRunProbeWatch_ValidFlags(t *testing.T) {
	flags := &probeWatchFlags{
		interval: 5 * time.Second,
		status:   "up",
	}

	err := runProbeWatch(context.Background(), "", flags)

	// Should fail on API client initialization, not validation
	if err == nil {
		t.Error("expected error (no API client configured), got nil")
		return
	}

	// Error should NOT be a validation error
	validationErrors := []string{
		"invalid interval",
		"invalid value",
	}
	for _, ve := range validationErrors {
		if strings.Contains(err.Error(), ve) {
			t.Errorf("got unexpected validation error: %s", err.Error())
		}
	}
}
