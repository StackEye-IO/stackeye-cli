package cmd

import (
	"context"
	"slices"
	"strings"
	"testing"
)

func TestNewDashboardCmd(t *testing.T) {
	cmd := NewDashboardCmd()

	if cmd.Use != "dashboard" {
		t.Errorf("expected Use='dashboard', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Short != "Display organization monitoring overview" {
		t.Errorf("expected Short='Display organization monitoring overview', got %q", cmd.Short)
	}
}

func TestNewDashboardCmd_Aliases(t *testing.T) {
	cmd := NewDashboardCmd()

	expectedAliases := []string{"dash", "status", "overview"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("expected %d aliases, got %d", len(expectedAliases), len(cmd.Aliases))
	}

	for _, expected := range expectedAliases {
		if !slices.Contains(cmd.Aliases, expected) {
			t.Errorf("expected alias %q not found", expected)
		}
	}
}

func TestNewDashboardCmd_Long(t *testing.T) {
	cmd := NewDashboardCmd()

	long := cmd.Long

	// Should contain period documentation
	periods := []string{"24h", "7d", "30d"}
	for _, period := range periods {
		if !strings.Contains(long, period) {
			t.Errorf("expected Long description to mention period %q", period)
		}
	}

	// Should have usage examples
	if !strings.Contains(long, "stackeye dashboard") {
		t.Error("expected Long description to contain example commands")
	}

	// Should mention output format
	if !strings.Contains(long, "json") {
		t.Error("expected Long description to mention JSON output option")
	}
}

func TestNewDashboardCmd_Flags(t *testing.T) {
	cmd := NewDashboardCmd()

	// Verify period flag exists with correct default
	flag := cmd.Flags().Lookup("period")
	if flag == nil {
		t.Error("expected flag 'period' to exist")
		return
	}
	if flag.DefValue != "24h" {
		t.Errorf("flag 'period': expected default '24h', got %q", flag.DefValue)
	}
}

func TestRunDashboard_InvalidPeriod(t *testing.T) {
	tests := []struct {
		name         string
		period       string
		wantErrorMsg string
	}{
		{
			name:         "invalid period 1h",
			period:       "1h",
			wantErrorMsg: `invalid value "1h" for --period`,
		},
		{
			name:         "invalid period 1d",
			period:       "1d",
			wantErrorMsg: `invalid value "1d" for --period`,
		},
		{
			name:         "invalid period arbitrary",
			period:       "invalid",
			wantErrorMsg: `invalid value "invalid" for --period`,
		},
		{
			name:         "empty period",
			period:       "",
			wantErrorMsg: `invalid value "" for --period`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := &dashboardFlags{
				period: tt.period,
			}

			// Call runDashboard with a background context.
			// It should fail on validation before needing API client.
			err := runDashboard(context.Background(), flags)

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

func TestRunDashboard_ValidPeriod(t *testing.T) {
	validPeriods := []string{"24h", "7d", "30d"}

	for _, period := range validPeriods {
		t.Run("valid_period_"+period, func(t *testing.T) {
			flags := &dashboardFlags{
				period: period,
			}

			err := runDashboard(context.Background(), flags)

			// Should fail on API client initialization, not validation
			if err == nil {
				t.Error("expected error (no API client configured), got nil")
				return
			}

			// Error should NOT be a period validation error
			if strings.Contains(err.Error(), "invalid period") {
				t.Errorf("got unexpected validation error for valid period %q: %s", period, err.Error())
			}
		})
	}
}

func TestGetStatusIcon(t *testing.T) {
	tests := []struct {
		status   string
		expected string
	}{
		{"up", "●"},
		{"down", "○"},
		{"degraded", "◐"},
		{"paused", "⏸"},
		{"pending", "◌"},
		{"unknown", "?"},
		{"", "?"},
	}

	for _, tt := range tests {
		t.Run("status_"+tt.status, func(t *testing.T) {
			result := getStatusIcon(tt.status)
			if result != tt.expected {
				t.Errorf("getStatusIcon(%q) = %q, want %q", tt.status, result, tt.expected)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"hello", 10, "hello"},
		{"hello", 5, "hello"},
		{"hello world", 5, "he..."},
		{"hello world", 8, "hello..."},
		{"hello world", 11, "hello world"},
		{"hello world", 3, "hel"},
		{"hello world", 2, "he"},
		{"", 5, ""},
	}

	for _, tt := range tests {
		t.Run(tt.input+"_"+string(rune('0'+tt.maxLen)), func(t *testing.T) {
			result := truncate(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}
