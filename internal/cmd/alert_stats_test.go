// Package cmd implements the CLI commands for StackEye.
// Task #7445
package cmd

import (
	"strings"
	"testing"
)

func TestNewAlertStatsCmd(t *testing.T) {
	cmd := NewAlertStatsCmd()

	// Verify command basic properties
	if cmd.Use != "stats" {
		t.Errorf("expected Use to be 'stats', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Long == "" {
		t.Error("expected Long description to be set")
	}

	// Verify aliases
	expectedAliases := []string{"statistics", "summary"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Fatalf("expected %d aliases, got %d", len(expectedAliases), len(cmd.Aliases))
	}
	for i, alias := range expectedAliases {
		if cmd.Aliases[i] != alias {
			t.Errorf("expected alias[%d] to be %q, got %q", i, alias, cmd.Aliases[i])
		}
	}

	// Verify flags exist with correct defaults
	periodFlag := cmd.Flags().Lookup("period")
	if periodFlag == nil {
		t.Error("expected 'period' flag to be defined")
	} else if periodFlag.DefValue != "24h" {
		t.Errorf("expected period flag default to be '24h', got %q", periodFlag.DefValue)
	}
}

func TestRunAlertStats_InvalidPeriod(t *testing.T) {
	tests := []struct {
		name    string
		period  string
		wantErr string
	}{
		{
			name:    "invalid period value",
			period:  "invalid",
			wantErr: "invalid period",
		},
		{
			name:    "uppercase period",
			period:  "24H",
			wantErr: "invalid period",
		},
		{
			name:    "unsupported duration",
			period:  "1h",
			wantErr: "invalid period",
		},
		{
			name:    "unsupported days",
			period:  "14d",
			wantErr: "invalid period",
		},
		{
			name:    "empty period",
			period:  "",
			wantErr: "invalid period",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := &alertStatsFlags{period: tt.period}
			err := runAlertStats(t.Context(), flags)

			if err == nil {
				t.Error("expected error for invalid period, got nil")
				return
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestRunAlertStats_ValidPeriods(t *testing.T) {
	validPeriods := []string{"24h", "7d", "30d"}

	for _, period := range validPeriods {
		t.Run(period, func(t *testing.T) {
			flags := &alertStatsFlags{period: period}
			err := runAlertStats(t.Context(), flags)

			// We expect an API client error since we're not mocking,
			// but we should NOT get a period validation error
			if err != nil && strings.Contains(err.Error(), "invalid period") {
				t.Errorf("period %q should be valid, got validation error: %v", period, err)
			}
		})
	}
}
